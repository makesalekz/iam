package biz

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	v1 "iam/api/iam/v1"
	"iam/ent"
	"iam/ent/property"
	"iam/internal/data"
	notifications_v1 "notifications/api/notifications/v1"
	tenants_v1 "tenants/api/tenants/v1"

	"github.com/go-kratos/kratos/v2/log"
	jwtv4 "github.com/golang-jwt/jwt/v4"
	"github.com/nyaruka/phonenumbers"
)

const DEFAULT_REGION = "KZ"
const AUTH_OTP_DURATION = 5 * time.Minute
const ACCESS_TOKEN_DURATION = 10 * time.Minute
const REFRESH_TOKEN_DURATION = 30 * 24 * time.Hour

// GreeterUsecase is a Greeter usecase.
type AuthUsecase struct {
	log       *log.Helper
	queue     *QueueManager
	jwt       *data.JwtProcessor
	dialer    *data.Dialer
	usersRepo data.UsersRepo
	otpRepo   data.OtpRepo
}

// NewAuthUsecase new a Greeter usecase.
func NewAuthUsecase(
	logger log.Logger,
	jwt *data.JwtProcessor,
	dialer *data.Dialer,
	usersRepo data.UsersRepo,
	otpRepo data.OtpRepo,
	queue *QueueManager,
) (*AuthUsecase, error) {
	return &AuthUsecase{
		log:       log.NewHelper(logger),
		jwt:       jwt,
		dialer:    dialer,
		usersRepo: usersRepo,
		otpRepo:   otpRepo,
		queue:     queue,
	}, nil
}

func (uc *AuthUsecase) AuthUserByPhone(ctx context.Context, phone string) (int64, error) {
	phoneNumber, err := phonenumbers.Parse(phone, DEFAULT_REGION)
	if err != nil {
		return 0, v1.ErrorInvalidPhoneNumber("parse error: %s", err.Error())
	}
	if !phonenumbers.IsValidNumber(phoneNumber) {
		return 0, v1.ErrorInvalidPhoneNumber("invalid phone number: %s", phone)
	}

	phone = phonenumbers.Format(phoneNumber, phonenumbers.E164)

	uc.log.Infof("phone, %v", phone)

	user, err := uc.usersRepo.GetUserByPhone(ctx, phone)
	if err != nil {
		if ent.IsNotFound(err) {
			user, err = uc.usersRepo.CreateUserWithPhone(ctx, phone)
		}
		if err != nil {
			return 0, v1.ErrorDatabaseQuery("db error (UsersRepo): %s", err.Error())
		}
	}

	otp, err := uc.otpRepo.CreateOneTimePassword(ctx, int64(user.ID), property.Phone, AUTH_OTP_DURATION)
	if err != nil {
		return 0, v1.ErrorDatabaseQuery("db error (OtpRepo): %s", err.Error())
	}

	debug := os.Getenv("DEBUG")
	if debug == "" { // don't send sms in debug mode
		senderClient, err := uc.dialer.Notifications(ctx)
		if err != nil {
			return 0, v1.ErrorGrpcConnection("dialer.Notifications: %s", err.Error())
		}

		_, err = senderClient.PersonalSmsSender(ctx, &notifications_v1.PersonalSmsSenderRequest{
			Phone:   phone,
			Message: fmt.Sprintf("Enter this code to sign in: %s", otp.Code),
		})
		if err != nil {
			return 0, v1.ErrorServiceFailed("senderClient.PersonalSmsSender: %s", err.Error())
		}
	}

	return int64(user.ID), nil
}

func (uc *AuthUsecase) AuthUserByCode(ctx context.Context, userId int64, code string) error {
	user, err := uc.usersRepo.GetUserById(ctx, userId)
	if err != nil {
		return v1.ErrorDatabaseQuery("DB Error (UsersRepo): %s", err.Error())
	}

	otp, err := uc.otpRepo.CheckOneTimePassword(ctx, user.ID, code)
	if err != nil {
		if ent.IsNotFound(err) {
			return v1.ErrorInvalidCode("invalid code")
		}
		return v1.ErrorDatabaseQuery("DB Error (OtpRepo): %s", err.Error())
	}

	err = uc.handleUserVerification(ctx, user, otp)
	if err != nil {
		return err
	}

	return nil
}

func (uc *AuthUsecase) handleUserVerification(ctx context.Context, user *ent.User, otp *ent.OneTimePassword) error {
	userShort := userShortFromDto(user)

	switch otp.Type {
	case property.Phone:
		if user.PhoneVerified {
			return nil
		}
		uc.queue.GetRemote(QueueContactsPhoneVerified).Pub(userShort)

		return uc.usersRepo.PhoneVerified(ctx, userShort.GetId())
	case property.Email:
		if user.EmailVerified {
			return nil
		}
		uc.queue.GetRemote(QueueContactsEmailVerified).Pub(userShort)

		return uc.usersRepo.EmailVerified(ctx, userShort.GetId())
	}

	return v1.ErrorInternal("unrecognized otpType")
}

func (uc *AuthUsecase) CheckIdToken(ctx context.Context) (int64, error) {
	userId, ok := uc.jwt.GetUserIdFromContext(ctx)
	if !ok {
		return 0, v1.ErrorInvalidToken("access denied")
	}

	// TODO: use JTI and save refresh token in DB

	return userId, nil
}

func (uc *AuthUsecase) GenerateIdToken(ctx context.Context, userId int64) (string, error) {
	claims := &jwtv4.RegisteredClaims{
		Issuer:    "iam",
		Audience:  jwtv4.ClaimStrings{"refresh"},
		Subject:   strconv.FormatInt(userId, 10),
		IssuedAt:  jwtv4.NewNumericDate(time.Now()),
		ExpiresAt: jwtv4.NewNumericDate(time.Now().Add(REFRESH_TOKEN_DURATION)),
	}
	token := jwtv4.NewWithClaims(jwtv4.SigningMethodHS256, claims)

	result, err := token.SignedString(uc.jwt.GetSecret())
	if err != nil {
		uc.log.Errorf("token.SignedString: %s", err.Error())
		return "", v1.ErrorInternal("internal error")
	}

	return result, nil
}

func (uc *AuthUsecase) GenerateAccessToken(ctx context.Context, userId int64) (string, error) {
	duration := ACCESS_TOKEN_DURATION
	debug := os.Getenv("DEBUG")
	if debug != "" { // set access token duration to 1 month in debug mode
		duration = REFRESH_TOKEN_DURATION
	}

	claims := &jwtv4.RegisteredClaims{
		Issuer:    "iam",
		Audience:  jwtv4.ClaimStrings{"personal"},
		Subject:   strconv.FormatInt(userId, 10),
		IssuedAt:  jwtv4.NewNumericDate(time.Now()),
		ExpiresAt: jwtv4.NewNumericDate(time.Now().Add(duration)),
	}
	token := jwtv4.NewWithClaims(jwtv4.SigningMethodHS256, claims)

	result, err := token.SignedString(uc.jwt.GetSecret())
	if err != nil {
		uc.log.Errorf("token.SignedString: %s", err.Error())
		return "", v1.ErrorInternal("internal error")
	}

	return result, nil
}

func (uc *AuthUsecase) GenerateTenantToken(ctx context.Context, userId, tenantId int64) (string, error) {
	duration := ACCESS_TOKEN_DURATION
	debug := os.Getenv("DEBUG")
	if debug != "" { // set access token duration to 1 month in debug mode
		duration = REFRESH_TOKEN_DURATION
	}

	claims := &data.TenantClaims{
		RegisteredClaims: jwtv4.RegisteredClaims{
			Issuer:    "iam",
			Audience:  jwtv4.ClaimStrings{"tenant"},
			Subject:   strconv.FormatInt(userId, 10),
			IssuedAt:  jwtv4.NewNumericDate(time.Now()),
			ExpiresAt: jwtv4.NewNumericDate(time.Now().Add(duration)),
		},
		TenantId: tenantId,
	}
	uc.log.Debugf("claims: %+v", claims)

	tenantMemberClient, err := uc.dialer.TenantsMembers(ctx, claims)
	if err != nil {
		return "", v1.ErrorGrpcConnection("dialer.TenantsMembers: %s", err.Error())
	}

	reply, err := tenantMemberClient.GetMember(ctx, &tenants_v1.GetMemberRequest{
		UserId: userId,
	})
	if err != nil {
		return "", v1.ErrorServiceFailed("tenantMemberClient.GetMember: %s", err.Error())
	}

	claims.MemberId = reply.Member
	claims.GroupsIds = reply.Groups

	token := jwtv4.NewWithClaims(jwtv4.SigningMethodHS256, claims)

	result, err := token.SignedString(uc.jwt.GetSecret())
	if err != nil {
		uc.log.Errorf("token.SignedString: %s", err.Error())
		return "", v1.ErrorInternal("internal error")
	}

	return result, nil
}

func userShortFromDto(user *ent.User) *v1.UserShort {
	replyUser := &v1.UserShort{
		Id:          user.ID,
		Name:        user.Name,
		LastLoginAt: user.LastLoginAt.Format(time.RFC3339),
	}

	if user.Phone != nil {
		replyUser.Phone = *user.Phone
	}
	if user.Email != nil {
		replyUser.Email = *user.Email
	}
	if user.Avatar != nil {
		replyUser.Avatar = *user.Avatar
	}

	return replyUser
}
