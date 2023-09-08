package biz

import (
	"context"
	_ "embed"
	"fmt"
	"strconv"
	"time"

	auth "iam/api/auth/v1"
	"iam/ent"
	"iam/ent/property"
	"iam/internal/conf"
	"iam/internal/data"
	sender "iam/third_party/api/send/v1"

	consul "github.com/go-kratos/consul/registry"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hashicorp/consul/api"
	"github.com/nyaruka/phonenumbers"
)

const DEFAULT_REGION = "KZ"
const AUTH_OTP_DURATION = 5 * time.Minute
const AUTH_TOKEN_DURATION = 30 * time.Minute
const GENERAL_TOKEN_DURATION = 30 * 24 * time.Hour

// TODO: move to vault
//
//go:embed jwt.key
var jwtSecret []byte

// GreeterUsecase is a Greeter usecase.
type UsersUsecase struct {
	conf      *conf.Bootstrap
	log       *log.Helper
	discovery registry.Discovery
	usersRepo data.UsersRepo
	otpRepo   data.OtpRepo
}

// NewGreeterUsecase new a Greeter usecase.
func NewUsersUsecase(c *conf.Bootstrap, logger log.Logger, consulClient *api.Client, usersRepo data.UsersRepo, otpRepo data.OtpRepo) (*UsersUsecase, error) {
	dis := consul.New(consulClient)

	return &UsersUsecase{
		conf:      c,
		log:       log.NewHelper(logger),
		discovery: dis,
		usersRepo: usersRepo,
		otpRepo:   otpRepo,
	}, nil
}

func makeJwtTokenFor(userId int64, duration time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": "iam",
		"aud": "personal",
		"sub": strconv.FormatInt(userId, 10),
		"exp": time.Now().Add(duration).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", auth.ErrorInternal("Token generation error: %s", err.Error())
	}

	return tokenString, nil
}

func (uc *UsersUsecase) dialNotifications(ctx context.Context) (sender.SenderClient, error) {
	conn, err := grpc.DialInsecure(
		ctx,
		grpc.WithEndpoint(uc.conf.Discovery.Notifications),
		grpc.WithDiscovery(uc.discovery),
		grpc.WithTimeout(10*time.Second),
	)
	if err != nil {
		return nil, err
	}
	return sender.NewSenderClient(conn), nil
}

func (uc *UsersUsecase) AuthUserByPhone(ctx context.Context, phone string) (string, error) {
	phoneNumber, err := phonenumbers.Parse(phone, DEFAULT_REGION)
	if err != nil {
		return "", auth.ErrorInvalidPhoneNumber("Parse error: %s", err.Error())
	}
	if !phonenumbers.IsValidNumber(phoneNumber) {
		return "", auth.ErrorInvalidPhoneNumber("Invalid phone number: %s", phone)
	}

	phone = phonenumbers.Format(phoneNumber, phonenumbers.E164)

	uc.log.Infof("phone, %v", phone)

	user, err := uc.usersRepo.GetUserByPhone(ctx, phone)
	if err != nil {
		if ent.IsNotFound(err) {
			user, err = uc.usersRepo.CreateUserWithPhone(ctx, phone)
		}
		if err != nil {
			return "", auth.ErrorDatabaseQuery("DB Error (UsersRepo): %s", err.Error())
		}
	}

	otp, err := uc.otpRepo.CreateOneTimePassword(ctx, int64(user.ID), property.Phone, AUTH_OTP_DURATION)
	if err != nil {
		return "", auth.ErrorDatabaseQuery("DB Error (OtpRepo): %s", err.Error())
	}

	senderClient, err := uc.dialNotifications(ctx)
	if err != nil {
		return "", auth.ErrorGrpcConnection("dialNotifications: %s", err.Error())
	}

	reply, err := senderClient.PersonalSmsSender(ctx, &sender.PersonalSmsSenderRequest{
		Phone:   phone,
		Message: fmt.Sprintf("Enter this code to sign in: %s", otp.Code),
	})
	if err != nil {
		return "", auth.ErrorServiceFailed("sender.PersonalSmsSender: %s", err.Error())
	}
	uc.log.Infof("sender.PersonalSmsSender: %s", reply.Result)

	return makeJwtTokenFor(int64(user.ID), AUTH_TOKEN_DURATION)
}

func (uc *UsersUsecase) AuthUserByCode(ctx context.Context, tokenString string, code string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		return "", auth.ErrorInvalidToken("Token parse error: %s", err.Error())
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", auth.ErrorInvalidToken("Token is not valid")
	}

	sub, err := claims.GetSubject()
	if err != nil {
		return "", auth.ErrorInvalidToken("Token sub is not valid %v", err)
	}

	id, err := strconv.ParseInt(sub, 10, 64)
	if err != nil {
		return "", auth.ErrorInvalidToken("Token sub is not valid %v", err)
	}

	user, err := uc.usersRepo.GetUserById(ctx, id)
	if err != nil {
		return "", auth.ErrorDatabaseQuery("DB Error (UsersRepo): %s", err.Error())
	}

	ok, err = uc.otpRepo.CheckOneTimePassword(ctx, user.ID, code)
	if err != nil {
		return "", auth.ErrorDatabaseQuery("DB Error (OtpRepo): %s", err.Error())
	}
	if !ok {
		return "", auth.ErrorInvalidCode("Invalid code")
	}

	return makeJwtTokenFor(user.ID, GENERAL_TOKEN_DURATION)
}
