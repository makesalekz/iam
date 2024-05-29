package biz

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/ent/enum"
	"gitlab.calendaria.team/services/iam/internal/data"
	tenants_v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	u_jwt "gitlab.calendaria.team/services/utils/v1/jwt"
	u_nats "gitlab.calendaria.team/services/utils/v1/nats"
	u_auth "gitlab.calendaria.team/services/utils/v2/auth"
	u_struc "gitlab.calendaria.team/services/utils/v2/struc"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/nyaruka/phonenumbers"
)

const DEFAULT_REGION = "KZ"
const AUTH_OTP_DURATION = 5 * time.Minute
const ACCESS_TOKEN_DURATION = 10 * time.Minute
const REFRESH_TOKEN_DURATION = 30 * 24 * time.Hour
const PERSONAL_WORKSPACE = "My Workspace"

// GreeterUsecase is a Greeter usecase.
type AuthUsecase struct {
	log           *log.Helper
	queue         *u_nats.QueueManager
	jwt           *u_jwt.JwtProcessor
	usersRepo     data.UsersRepo
	otpRepo       data.OtpRepo
	tenants       *data.TenantsRemote
	notifications *data.NotificationsRemote
}

// NewAuthUsecase new a Greeter usecase.
func NewAuthUsecase(
	logger log.Logger,
	jwt *u_jwt.JwtProcessor,
	usersRepo data.UsersRepo,
	otpRepo data.OtpRepo,
	queue *u_nats.QueueManager,
	tenants *data.TenantsRemote,
	notifications *data.NotificationsRemote,
) (*AuthUsecase, error) {
	return &AuthUsecase{
		log:           log.NewHelper(logger),
		jwt:           jwt,
		usersRepo:     usersRepo,
		otpRepo:       otpRepo,
		queue:         queue,
		tenants:       tenants,
		notifications: notifications,
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

	user, err := uc.usersRepo.GetUserByPhone(ctx, phone)
	if err != nil {
		if ent.IsNotFound(err) {
			user, err = uc.usersRepo.CreateUserWithPhone(ctx, phone)
		}
		if err != nil {
			return 0, v1.ErrorDatabaseQuery("database error: %s", err.Error())
		}
	}

	otp, err := uc.otpRepo.CreateOneTimePassword(ctx, user.ID, enum.Phone, AUTH_OTP_DURATION)
	if err != nil {
		return 0, v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}

	err = uc.notifications.PersonalSmsSender(ctx, phone, fmt.Sprintf("Calendaria: %s", otp.Code))
	if err != nil {
		return 0, v1.ErrorServiceFailed("notification: %s", err.Error())
	}

	return user.ID, nil
}

func (uc *AuthUsecase) AuthUserByCode(ctx context.Context, userId int64, code string) error {
	user, err := uc.usersRepo.GetUserById(ctx, userId)
	if err != nil {
		return v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}

	otp, err := uc.otpRepo.CheckOneTimePassword(ctx, user.ID, code)
	if err != nil {
		if ent.IsNotFound(err) {
			return v1.ErrorInvalidCode("invalid code")
		}
		return v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}

	err = uc.handleUserVerification(ctx, user, otp)
	if err != nil {
		return err
	}

	return nil
}

func (uc *AuthUsecase) handleUserVerification(ctx context.Context, user *ent.User, otp *ent.OneTimePassword) error {
	userShort := userShortFromDto(user)

	if user.DefaultTenantID == nil {
		tenantContext := u_auth.AppendAuthIds(ctx, user.ID, 0)
		personalTenant, err := uc.tenants.CreateTenants(tenantContext, PERSONAL_WORKSPACE)
		if err != nil {
			return v1.ErrorGrpcConnection("CreateTenants error: %s", err.Error())
		}

		_, err = uc.usersRepo.UpdateUserData(tenantContext, user, data.UpdateUserDto{TenantId: personalTenant.Id})
		if err != nil {
			return v1.ErrorDatabaseQuery("UpdateUserData gone wrong: %s", err.Error())
		}

		uc.queue.GetRemote(QueueEventsDefaultCalendars).Pub(&u_struc.AuthIds{
			ActorId:  user.ID,
			TenantId: personalTenant.Id,
		})
	}

	switch otp.Type {
	case enum.Phone:
		if user.PhoneVerified {
			return nil
		}
		uc.queue.GetRemote(QueueContactsPhoneVerified).Pub(userShort)

		return uc.usersRepo.PhoneVerified(ctx, userShort.GetId())
	case enum.Email:
		if user.EmailVerified {
			return nil
		}
		uc.queue.GetRemote(QueueContactsEmailVerified).Pub(userShort)

		return uc.usersRepo.EmailVerified(ctx, userShort.GetId())
	}

	return v1.ErrorInternal("unrecognized otpType")
}

func (uc *AuthUsecase) GenerateIdToken(ctx context.Context, userId int64) (string, error) {
	claims := &jwt.RegisteredClaims{
		Issuer:    "iam",
		Audience:  jwt.ClaimStrings{"refresh"},
		Subject:   strconv.FormatInt(userId, 10),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(REFRESH_TOKEN_DURATION)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	result, err := token.SignedString(uc.jwt.GetSecret())
	if err != nil {
		uc.log.Errorf("token.SignedString: %s", err.Error())
		return "", v1.ErrorInternal("internal error")
	}

	return result, nil
}

func (uc *AuthUsecase) GenerateAccessToken(ctx context.Context, userId int64) (string, error) {
	user, err := uc.usersRepo.GetUserById(ctx, userId)
	if err != nil {
		return "", v1.ErrorDatabaseQuery("get user: %s", err.Error())
	}
	if user.DefaultTenantID == nil {
		return "", v1.ErrorInternal("personal tenant non existent")
	}

	result, err := uc.GenerateTenantToken(ctx, *user.DefaultTenantID, userId)
	if err != nil {
		return "", err
	}

	return result, nil
}

func (uc *AuthUsecase) GenerateTenantToken(ctx context.Context, tenantId, userId int64) (string, error) {
	duration := ACCESS_TOKEN_DURATION
	debug := os.Getenv("DEBUG")
	if debug != "" { // set access token duration to 1 month in debug mode
		duration = REFRESH_TOKEN_DURATION
	}

	claims := &u_jwt.TenantClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "iam",
			Audience:  jwt.ClaimStrings{"tenant"},
			Subject:   strconv.FormatInt(userId, 10),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		},
		TenantId: tenantId,
	}

	reply, err := uc.tenants.GetMemberIdentities(ctx, tenantId, userId)
	if err != nil {
		return "", tenants_v1.ErrorServiceFailed("tenants: %s", err.Error())
	}

	claims.MemberId = reply.Member
	claims.GroupsIds = reply.Groups

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

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

func (uc *AuthUsecase) TempAddDefaultTenants(ctx context.Context) error {
	users, err := uc.usersRepo.TempGetUsersWithoutDefaultTenant(ctx)
	if err != nil {
		return v1.ErrorInternal("can't get users without tenant")
	}

	for _, user := range users {
		tenant, err := uc.tenants.CreateTenants(u_auth.AppendAuthIds(context.Background(), user.ID, 0), PERSONAL_WORKSPACE)
		if err != nil {
			return v1.ErrorInternal("can't create tenant %s", err.Error())
		}

		_, err = uc.usersRepo.UpdateUserData(ctx, user, data.UpdateUserDto{TenantId: tenant.Id})
		if err != nil {
			return v1.ErrorInternal("can't update tenant id %s", err.Error())
		}
	}

	return nil
}
