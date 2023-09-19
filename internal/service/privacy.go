package service

import (
	"context"

	v1 "iam/api/privacy/v1"
	"iam/ent"
	"iam/internal/biz"
	"iam/internal/data"

	"github.com/go-kratos/kratos/v2/log"
)

type PrivacyService struct {
	v1.UnimplementedPrivacyServer

	log *log.Helper
	jwt *data.JwtProcessor
	uc  *biz.PrivacyUsecase
}

func NewPrivacyService(logger log.Logger, jwt *data.JwtProcessor, uc *biz.PrivacyUsecase) *PrivacyService {
	return &PrivacyService{
		log: log.NewHelper(logger),
		jwt: jwt,
		uc:  uc,
	}
}

func (s *PrivacyService) GetPrivacy(ctx context.Context, req *v1.EmptyRequest) (*v1.PrivacyReply, error) {
	userId, ok := s.jwt.GetUserIdFromContext(ctx)
	if !ok {
		return nil, v1.ErrorUnauthorized("Unauthorized")
	}

	settings, err := s.uc.GetPrivacy(ctx, userId)
	if err != nil {
		return nil, v1.ErrorDatabaseQuery("Internal error")
	}

	return &v1.PrivacyReply{
		Settings: settings,
	}, nil
}

func (s *PrivacyService) UpdatePrivacy(ctx context.Context, req *v1.PrivacyRequest) (*v1.PrivacyReply, error) {
	userId, ok := s.jwt.GetUserIdFromContext(ctx)
	if !ok {
		return nil, v1.ErrorUnauthorized("Unauthorized")
	}

	settings, err := s.uc.UpdatePrivacy(ctx, userId, req.Settings)
	if err != nil {
		if ent.IsValidationError(err) {
			return nil, v1.ErrorInvalidRequest(err.Error())
		}
		s.log.Errorf("UpdatePrivacy error: %v", err)
		return nil, v1.ErrorDatabaseQuery("Internal error")
	}

	return &v1.PrivacyReply{
		Settings: settings,
	}, nil
}
