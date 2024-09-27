package service

import (
	"context"

	v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/internal/biz"
	"gitlab.calendaria.team/services/utils/v2/auth"
	"gitlab.calendaria.team/services/utils/v2/struc"
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
	appID := auth.GetAppIdFromContext(ctx)
	if appID == "" {
		return nil, v1.ErrorEmptyAppId("empty app id")
	}

	dto := &biz.AuthPhoneDto{
		AppID:                struc.ApplicationID(appID),
		Phone:                req.GetPhone(),
		IsRegistrationNeeded: req.GetIsRegistrationNeeded(),
		IsRegistration:       req.GetIsRegistration(),
		AppSignature:         req.GetAppSignature(),
	}
	if err := dto.Validate(); err != nil {
		return nil, err
	}

	userID, err := s.au.AuthUserByPhone(ctx, dto)
	if err != nil {
		return nil, err
	}

	return &v1.AuthByPhoneReply{UserId: userID}, nil
}

func (s *AuthService) AuthByEmail(ctx context.Context, req *v1.AuthByEmailRequest) (*v1.AuthByPhoneReply, error) {
	userID, err := s.au.AuthUserByEmail(
		ctx, req.GetEmail(), req.GetLanguage(), req.GetIsRegistrationNeeded(), req.GetIsRegistration(),
	)
	if err != nil {
		return nil, err
	}

	return &v1.AuthByPhoneReply{UserId: userID}, nil
}

func (s *AuthService) AuthByCode(ctx context.Context, req *v1.AuthByCodeRequest) (*v1.TokenReply, error) {
	user, err := s.au.GetUserByID(ctx, req.GetUserId(), true)
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

	// get/check user
	user, err := s.au.GetUserByID(ctx, actorID, false)
	if err != nil {
		return nil, err
	}

	var accessToken string
	if req.GetTenantId() != 0 {
		accessToken, err = s.au.GenerateTenantToken(ctx, req.GetTenantId(), actorID)
	} else {
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
