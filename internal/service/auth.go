package service

import (
	"context"
	"os"
	"strconv"
	"time"

	v1 "iam/api/auth/v1"
	"iam/internal/biz"
	"iam/internal/data"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	jwtv4 "github.com/golang-jwt/jwt/v4"
)

const TOKEN_DURATION = 24 * time.Hour

type AuthService struct {
	v1.UnimplementedAuthServer

	log *log.Helper
	jwt *data.JwtProcessor
	au  *biz.AuthUsecase
}

func NewAuthService(logger log.Logger, jwt *data.JwtProcessor, au *biz.AuthUsecase) *AuthService {
	return &AuthService{
		log: log.NewHelper(logger),
		jwt: jwt,
		au:  au,
	}
}

func (s *AuthService) AuthByPhone(ctx context.Context, req *v1.AuthByPhoneRequest) (*v1.AuthByPhoneReply, error) {
	userId, err := s.au.AuthUserByPhone(ctx, req.Phone)
	if err != nil {
		if v1.IsInvalidPhoneNumber(err) {
			return nil, err
		}
		s.log.Errorf("au.AuthUserByPhone: ", err)
		return nil, errors.InternalServer("internal", "internal error")
	}

	return &v1.AuthByPhoneReply{UserId: userId}, nil
}

func (s *AuthService) AuthByCode(ctx context.Context, req *v1.AuthByCodeRequest) (*v1.AuthByCodeReply, error) {
	err := s.au.AuthUserByCode(ctx, req.UserId, req.Code)
	if err != nil {
		if v1.IsInvalidCode(err) {
			return nil, err
		}
		s.log.Errorf("au.AuthUserByCode: ", err)
		return nil, errors.InternalServer("internal", "internal error")
	}

	claims := &jwtv4.RegisteredClaims{
		Issuer:    "iam",
		Audience:  jwtv4.ClaimStrings{"personal"},
		Subject:   strconv.FormatInt(req.UserId, 10),
		IssuedAt:  jwtv4.NewNumericDate(time.Now()),
		ExpiresAt: jwtv4.NewNumericDate(time.Now().Add(TOKEN_DURATION)),
	}
	token := jwtv4.NewWithClaims(jwtv4.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(s.jwt.GetSecret())
	if err != nil {
		s.log.Errorf("token.SignedString: ", err)
		return nil, errors.InternalServer("internal", "internal error")
	}

	return &v1.AuthByCodeReply{Token: tokenString}, nil
}

func (s *AuthService) TempAuthBySuperCode(ctx context.Context, req *v1.AuthByCodeRequest) (*v1.AuthByCodeReply, error) {
	automigrate := os.Getenv("AUTOMIGRATE") // check if we are in dev mode
	if automigrate == "" || req.Code != "sup3rcaL2033" {
		return nil, errors.InternalServer("internal", "internal error")
	}

	claims := &jwtv4.RegisteredClaims{
		Issuer:    "iam",
		Audience:  jwtv4.ClaimStrings{"personal"},
		Subject:   strconv.FormatInt(req.UserId, 10),
		IssuedAt:  jwtv4.NewNumericDate(time.Now()),
		ExpiresAt: jwtv4.NewNumericDate(time.Now().Add(TOKEN_DURATION * 30)),
	}
	token := jwtv4.NewWithClaims(jwtv4.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(s.jwt.GetSecret())
	if err != nil {
		s.log.Errorf("token.SignedString: ", err)
		return nil, errors.InternalServer("internal", "internal error")
	}

	return &v1.AuthByCodeReply{Token: tokenString}, nil
}
