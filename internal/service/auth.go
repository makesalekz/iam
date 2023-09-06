package service

import (
	"context"

	v1 "iam/api/iam/v1"
	"iam/internal/biz"
	"iam/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
)

type AuthService struct {
	v1.UnimplementedAuthServer

	conf *conf.Bootstrap
	logger log.Logger
}

func NewAuthService(cfg *conf.Bootstrap, logger log.Logger, dummy *biz.Dummy) *AuthService {
	return &AuthService{
		conf: cfg,
		logger: logger,
	}
}

func (s *AuthService) AuthByLoginPassword(ctx context.Context, req *v1.AuthByLoginPasswordRequest) (*v1.AuthByLoginPasswordReply, error) {
	return &v1.AuthByLoginPasswordReply{}, nil
}

func (s *AuthService) AuthByPhone(ctx context.Context, req *v1.AuthByPhoneRequest) (*v1.AuthByPhoneReply, error) {
	return &v1.AuthByPhoneReply{}, nil
}

func (s *AuthService) AuthByPhoneCode(ctx context.Context, req *v1.AuthByPhoneCodeRequest) (*v1.AuthByPhoneCodeReply, error) {
	return &v1.AuthByPhoneCodeReply{}, nil
}
