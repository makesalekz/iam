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
	"golang.org/x/exp/rand"
)

const otpLength = 6
const digits = "0123456789"
const debugOtpCode = "777333"

const verifiablePhone = "+77710012030"
const verifiableOtpCode = "667423"

const defaultRegion = "KZ"
const authOtpDuration = time.Duration(5) * time.Minute
const defaultAccessTokenDuration = time.Duration(10) * time.Minute
const defaultRefreshTokenDuration = time.Duration(30*24) * time.Hour
const personalWorkspace = "My Workspace"

// GreeterUsecase is a Greeter usecase.
type AuthUsecase struct {
	log                  *log.Helper
	queue                *u_nats.QueueManager
	jwt                  *u_jwt.JwtProcessor
	usersRepo            data.UsersRepo
	otpRepo              data.OtpRepo
	tenants              *data.TenantsRemote
	notifications        *data.NotificationsRemote
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
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
	uc := &AuthUsecase{
		log:           log.NewHelper(logger),
		jwt:           jwt,
		usersRepo:     usersRepo,
		otpRepo:       otpRepo,
		queue:         queue,
		tenants:       tenants,
		notifications: notifications,
	}

	// set default access token duration
	uc.accessTokenDuration = defaultAccessTokenDuration

	// set default refresh token duration if debug mode is enabled
	debug := os.Getenv("DEBUG")
	if debug != "" { // set access token duration to 1 month in debug mode
		uc.accessTokenDuration = defaultRefreshTokenDuration
	}

	// set access token duration from environment variable if it is set
	accessDuration := os.Getenv("TOKEN_DURATION")
	if accessDuration != "" {
		duration, err := time.ParseDuration(accessDuration)
		if err == nil {
			uc.accessTokenDuration = duration
		}
	}

	// set default refresh token duration
	uc.refreshTokenDuration = defaultRefreshTokenDuration

	// set refresh token duration from environment variable if it is set
	refreshDuration := os.Getenv("REFRESH_TOKEN_DURATION")
	if refreshDuration != "" {
		duration, err := time.ParseDuration(refreshDuration)
		if err == nil {
			uc.refreshTokenDuration = duration
		}
	}

	return uc, nil
}

func (uc *AuthUsecase) AuthUserByPhone(ctx context.Context, phone string) (int64, error) {
	phoneNumber, err := phonenumbers.Parse(phone, defaultRegion)
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

	var code string
	switch {
	case phone == verifiablePhone:
		// use fixed code for verifiable phone
		code = verifiableOtpCode

	case os.Getenv("DEBUG") != "":
		// use fixed code in debug mode
		code = debugOtpCode

	default:
		code = generateRandomNumber(otpLength)
	}

	otp, err := uc.otpRepo.CreateOneTimePassword(ctx, user.ID, enum.Phone, code, authOtpDuration)
	if err != nil {
		return 0, v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}

	err = uc.notifications.PersonalSmsSender(ctx, phone, fmt.Sprintf("Calendaria: %s", otp.Code))
	if err != nil {
		uc.log.Errorf("notifications.PersonalSmsSender: %s", err.Error())
	}

	return user.ID, nil
}

func (uc *AuthUsecase) AuthUserByEmail(ctx context.Context, email, lang string) (int64, error) {
	user, err := uc.usersRepo.GetUserByEmail(ctx, email)
	if err != nil {
		if ent.IsNotFound(err) {
			user, err = uc.usersRepo.CreateUserWithEmail(ctx, email)
		}
		if err != nil {
			return 0, v1.ErrorDatabaseQuery("database error: %s", err.Error())
		}
	}

	var code string
	if os.Getenv("DEBUG") != "" {
		// use fixed code in debug mode
		code = debugOtpCode
	} else {
		code = generateRandomNumber(otpLength)
	}

	otp, err := uc.otpRepo.CreateOneTimePassword(ctx, user.ID, enum.Email, code, authOtpDuration)
	if err != nil {
		return 0, v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}

	emailDetails := map[string]string{
		"AuthCode": otp.Code,
		"Email":    email,
	}
	err = uc.notifications.PersonalEmailSender(ctx, email, "confirm_email", lang, emailDetails)
	if err != nil {
		return 0, v1.ErrorServiceFailed("notification: %s", err.Error())
	}

	return user.ID, nil
}

func (uc *AuthUsecase) GetUserByID(ctx context.Context, userID int64) (*ent.User, error) {
	user, err := uc.usersRepo.GetUserById(ctx, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, v1.ErrorUserNotFound("user not found")
		}
		return nil, v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}

	return user, nil
}

func (uc *AuthUsecase) AuthUserByCode(ctx context.Context, user *ent.User, code string) error {
	otp, err := uc.otpRepo.CheckOneTimePassword(ctx, user.ID, code)
	if err != nil {
		if ent.IsNotFound(err) {
			return v1.ErrorInvalidCode("invalid code")
		}
		return v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}

	return uc.handleUserVerification(ctx, user, otp)
}

func (uc *AuthUsecase) handleUserVerification(ctx context.Context, user *ent.User, otp *ent.OneTimePassword) error {
	userShort := userShortFromDto(user)

	if user.DefaultTenantID == nil {
		tenantContext := u_auth.AppendAuthIds(ctx, user.ID, 0)
		personalTenant, err := uc.tenants.CreateTenants(tenantContext, personalWorkspace)
		if err != nil {
			return v1.ErrorGrpcConnection("CreateTenants error: %s", err.Error())
		}

		_, err = uc.usersRepo.UpdateUserData(
			tenantContext, user, data.UpdateUserDto{TenantId: personalTenant.GetId()},
		)
		if err != nil {
			return v1.ErrorDatabaseQuery("UpdateUserData gone wrong: %s", err.Error())
		}

		tenantID := personalTenant.GetId()
		user.DefaultTenantID = &tenantID

		uc.queue.GetRemote(QueueEventsDefaultCalendars).Pub(
			&u_struc.AuthIds{
				ActorId:  user.ID,
				TenantId: personalTenant.GetId(),
			},
		)
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

func (uc *AuthUsecase) GenerateIDToken(ctx context.Context, userID int64) (string, error) {
	claims := &jwt.RegisteredClaims{
		Issuer:    "iam",
		Audience:  jwt.ClaimStrings{"refresh"},
		Subject:   strconv.FormatInt(userID, 10),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(uc.refreshTokenDuration)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	result, err := token.SignedString(uc.jwt.GetSecret())
	if err != nil {
		uc.log.Errorf("token.SignedString: %s", err.Error())
		return "", v1.ErrorInternal("internal error")
	}

	return result, nil
}

func (uc *AuthUsecase) GenerateAccessToken(ctx context.Context, user *ent.User) (string, error) {
	if user.DefaultTenantID == nil {
		return "", v1.ErrorInternal("personal tenant non existent")
	}

	return uc.GenerateTenantToken(ctx, *user.DefaultTenantID, user.ID)
}

func (uc *AuthUsecase) GenerateTenantToken(ctx context.Context, tenantID, userID int64) (string, error) {
	claims := &u_jwt.TenantClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "iam",
			Audience:  jwt.ClaimStrings{"tenant"},
			Subject:   strconv.FormatInt(userID, 10),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(uc.accessTokenDuration)),
		},
		TenantId: tenantID,
	}

	reply, err := uc.tenants.GetMemberIdentities(ctx, tenantID, userID)
	if err != nil {
		return "", tenants_v1.ErrorServiceFailed("tenants: %s", err.Error())
	}

	claims.MemberId = reply.GetMember()
	claims.GroupsIds = reply.GetGroups()

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

func generateRandomNumber(n int) string {
	result := make([]byte, n)
	for i := range result {
		result[i] = digits[rand.Int63()%int64(len(digits))]
	}
	return string(result)
}
