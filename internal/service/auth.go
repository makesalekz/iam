package service

import (
	"context"

	v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/internal/biz"
	"gitlab.calendaria.team/services/utils/v2/auth"
)

type AuthService struct {
	v1.UnimplementedAuthServer

	au *biz.AuthUsecase
}

func NewAuthService(
	au *biz.AuthUsecase,
) *AuthService {
	return &AuthService{
		au: au,
	}
}

func (s *AuthService) AuthByPhone(ctx context.Context, req *v1.AuthByPhoneRequest) (*v1.AuthByPhoneReply, error) {
	userId, err := s.au.AuthUserByPhone(ctx, req.Phone)
	if err != nil {
		return nil, err
	}

	return &v1.AuthByPhoneReply{UserId: userId}, nil
}

func (s *AuthService) AuthByCode(ctx context.Context, req *v1.AuthByCodeRequest) (*v1.TokenReply, error) {
	err := s.au.AuthUserByCode(ctx, req.UserId, req.Code)
	if err != nil {
		return nil, err
	}

	accessToken, err := s.au.GeneratePersonalToken(ctx, req.UserId)
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

func (s *AuthService) RefreshToken(ctx context.Context, req *v1.TenantRequest) (*v1.TokenReply, error) {
	actorId := auth.GetActorIdFromContext(ctx)
	if actorId == 0 {
		return nil, v1.ErrorEmptyActorId("empty actor id")
	}

	var err error
	var accessToken string
	if req.TenantId != 0 {
		accessToken, err = s.au.GenerateTenantToken(ctx, req.TenantId, actorId)
	} else {
		accessToken, err = s.au.GeneratePersonalToken(ctx, actorId)
	}
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.au.GenerateIdToken(ctx, actorId)
	if err != nil {
		return nil, err
	}

	return &v1.TokenReply{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
