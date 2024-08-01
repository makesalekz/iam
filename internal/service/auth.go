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
	userID, err := s.au.AuthUserByPhone(ctx, req.GetPhone(), req.GetIsRegistrationNeeded())
	if err != nil {
		return nil, err
	}

	return &v1.AuthByPhoneReply{UserId: userID}, nil
}

func (s *AuthService) AuthByEmail(ctx context.Context, req *v1.AuthByEmailRequest) (*v1.AuthByPhoneReply, error) {
	userID, err := s.au.AuthUserByEmail(
		ctx, req.GetEmail(), req.GetLanguage(), req.GetIsRegistrationNeeded(),
	)
	if err != nil {
		return nil, err
	}

	return &v1.AuthByPhoneReply{UserId: userID}, nil
}

func (s *AuthService) AuthByCode(ctx context.Context, req *v1.AuthByCodeRequest) (*v1.TokenReply, error) {
	user, err := s.au.GetUserByID(ctx, req.GetUserId())
	if err != nil {
		return nil, err
	}

	err = s.au.AuthUserByCode(ctx, user, req.GetCode())
	if err != nil {
		return nil, err
	}

	accessToken, err := s.au.GenerateAccessToken(ctx, user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.au.GenerateIDToken(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &v1.TokenReply{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req *v1.TenantRequest) (*v1.TokenReply, error) {
	actorID := auth.GetActorIdFromContext(ctx)
	if actorID == 0 {
		return nil, v1.ErrorEmptyActorId("empty actor id")
	}

	var err error
	var accessToken string
	if req.GetTenantId() != 0 {
		accessToken, err = s.au.GenerateTenantToken(ctx, req.GetTenantId(), actorID)
	} else {
		user, err2 := s.au.GetUserByID(ctx, actorID)
		if err2 != nil {
			return nil, err2
		}

		accessToken, err = s.au.GenerateAccessToken(ctx, user)
	}
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.au.GenerateIDToken(ctx, actorID)
	if err != nil {
		return nil, err
	}

	return &v1.TokenReply{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
