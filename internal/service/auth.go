package service

import (
	"context"
	"strconv"
	"time"

	v1 "iam/api/auth/v1"
	"iam/internal/biz"
	"iam/internal/conf"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	jwtv4 "github.com/golang-jwt/jwt/v4"
)

const TOKEN_DURATION = 30 * 24 * time.Hour

type AuthService struct {
	v1.UnimplementedAuthServer

	conf *conf.Bootstrap
	log  *log.Helper
	jwt  *biz.JwtProcessor
	au   *biz.AuthUsecase
}

func NewAuthService(cfg *conf.Bootstrap, logger log.Logger, jwt *biz.JwtProcessor, au *biz.AuthUsecase) *AuthService {
	return &AuthService{
		conf: cfg,
		log:  log.NewHelper(logger),
		jwt:  jwt,
		au:   au,
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

	return &v1.AuthByPhoneReply{UserId: strconv.FormatInt(userId, 10)}, nil
}

func (s *AuthService) AuthByCode(ctx context.Context, req *v1.AuthByCodeRequest) (*v1.AuthByCodeReply, error) {
	userId, err := strconv.ParseInt(req.UserId, 10, 64)
	if err != nil {
		return nil, v1.ErrorInvalidRequest("Invalid request (user ID)")
	}

	err = s.au.AuthUserByCode(ctx, userId, req.Code)
	if err != nil {
		if v1.IsInvalidCode(err) {
			return nil, err
		}
		s.log.Errorf("au.AuthUserByCode: ", err)
		return nil, errors.InternalServer("internal", "internal error")
	}

	token := jwtv4.RegisteredClaims{
		Issuer:    "iam",
		Audience:  jwtv4.ClaimStrings{"personal"},
		Subject:   strconv.FormatInt(userId, 10),
		ExpiresAt: jwtv4.NewNumericDate(time.Now().Add(TOKEN_DURATION)),
	}
	jwtv4Token := jwtv4.NewWithClaims(jwtv4.SigningMethodHS256, token)

	tokenString, err := jwtv4Token.SignedString(s.jwt.GetSecret())
	if err != nil {
		s.log.Errorf("jwtv4Token.SignedString: ", err)
		return nil, errors.InternalServer("internal", "internal error")
	}

	return &v1.AuthByCodeReply{Token: tokenString}, nil
}
