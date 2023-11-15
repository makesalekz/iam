package service

import (
	"context"

	v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/internal/biz"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type AuthService struct {
	v1.UnimplementedAuthServer

	log *log.Helper
	au  *biz.AuthUsecase
}

func NewAuthService(logger log.Logger, au *biz.AuthUsecase) *AuthService {
	return &AuthService{
		log: log.NewHelper(logger),
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
		return nil, v1.ErrorInternal("internal error")
	}

	return &v1.AuthByPhoneReply{UserId: userId}, nil
}

func (s *AuthService) AuthByCode(ctx context.Context, req *v1.AuthByCodeRequest) (*v1.TokenReply, error) {
	err := s.au.AuthUserByCode(ctx, req.UserId, req.Code)
	if err != nil {
		if v1.IsInvalidCode(err) {
			return nil, err
		}
		s.log.Errorf("au.AuthUserByCode: ", err)
		return nil, v1.ErrorInternal("internal error")
	}

	accessToken, err := s.au.GenerateAccessToken(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.au.GenerateIdToken(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	return &v1.TokenReply{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) RefreshPersonalToken(ctx context.Context, req *utils_v1.EmptyRequest) (*v1.TokenReply, error) {
	userId, err := s.au.CheckIdToken(ctx)
	if err != nil {
		return nil, err
	}

	accessToken, err := s.au.GenerateAccessToken(ctx, userId)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.au.GenerateIdToken(ctx, userId)
	if err != nil {
		return nil, err
	}

	return &v1.TokenReply{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) RefreshTenantToken(ctx context.Context, req *v1.TenantRequest) (*v1.TokenReply, error) {
	userId, err := s.au.CheckIdToken(ctx)
	if err != nil {
		return nil, err
	}

	accessToken, err := s.au.GenerateTenantToken(ctx, userId, req.TenantId)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.au.GenerateIdToken(ctx, userId)
	if err != nil {
		return nil, err
	}

	return &v1.TokenReply{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
