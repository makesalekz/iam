package service

import (
	"context"

	v1 "iam/api/auth/v1"
	"iam/internal/biz"
	"iam/internal/conf"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

type AuthService struct {
	v1.UnimplementedAuthServer

	uc   *biz.UsersUsecase
	conf *conf.Bootstrap
	log  *log.Helper
}

func NewAuthService(cfg *conf.Bootstrap, logger log.Logger, uc *biz.UsersUsecase) *AuthService {
	return &AuthService{
		uc:   uc,
		conf: cfg,
		log:  log.NewHelper(logger),
	}
}

func (s *AuthService) AuthByPhone(ctx context.Context, req *v1.AuthByPhoneRequest) (*v1.AuthByPhoneReply, error) {
	token, err := s.uc.AuthUserByPhone(ctx, req.Phone)
	if err != nil {
		if v1.IsInvalidPhoneNumber(err) {
			return nil, err
		}
		s.log.Errorf("uc.AuthUserByPhone: ", err)
		return nil, errors.InternalServer("internal", "internal error")
	}

	return &v1.AuthByPhoneReply{Token: token}, nil
}

func (s *AuthService) AuthByCode(ctx context.Context, req *v1.AuthByCodeRequest) (*v1.AuthByCodeReply, error) {
	token, err := s.uc.AuthUserByCode(ctx, req.Token, req.Code)
	if err != nil {
		if v1.IsInvalidToken(err) || v1.IsInvalidCode(err) {
			return nil, err
		}
		s.log.Errorf("uc.AuthUserByCode: ", err)
		return nil, errors.InternalServer("internal", "internal error")
	}

	return &v1.AuthByCodeReply{Token: token}, nil
}
