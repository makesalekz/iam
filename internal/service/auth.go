package service

import (
	"context"

	pb "iam/api/iam/v1"
	"iam/internal/biz"
	"iam/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
)

type AuthService struct {
	pb.UnimplementedAuthServer

	uc     *biz.UsersUsecase
	conf   *conf.Bootstrap
	logger log.Logger
}

func NewAuthService(cfg *conf.Bootstrap, logger log.Logger, uc *biz.UsersUsecase) *AuthService {
	return &AuthService{
		uc:     uc,
		conf:   cfg,
		logger: logger,
	}
}

func (s *AuthService) AuthByLoginPassword(ctx context.Context, req *pb.AuthByLoginPasswordRequest) (*pb.AuthByLoginPasswordReply, error) {
	return &pb.AuthByLoginPasswordReply{}, nil
}

func (s *AuthService) AuthByPhone(ctx context.Context, req *pb.AuthByPhoneRequest) (*pb.AuthByPhoneReply, error) {
	code, err := s.uc.AuthUserByPhone(ctx, req.Phone)
	if err != nil {
		s.logger.Log(log.LevelError, "uc.AuthUserByPhone", err)
		return &pb.AuthByPhoneReply{}, nil
	}

	return &pb.AuthByPhoneReply{
		State: code,
	}, nil
}

func (s *AuthService) AuthByPhoneCode(ctx context.Context, req *pb.AuthByPhoneCodeRequest) (*pb.AuthByPhoneCodeReply, error) {
	return &pb.AuthByPhoneCodeReply{}, nil
}
