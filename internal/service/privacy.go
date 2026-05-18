package service

import (
	"context"

	v1 "github.com/makesalekz/iam/api/iam/v1"
	"github.com/makesalekz/iam/ent"
	"github.com/makesalekz/iam/internal/biz"
	utils_v1 "github.com/makesalekz/utils/api/utils/v1"
	"github.com/makesalekz/utils/v2/auth"
)

type PrivacyService struct {
	v1.UnimplementedPrivacyServer

	uc *biz.PrivacyUsecase
}

func NewPrivacyService(
	uc *biz.PrivacyUsecase,
) *PrivacyService {
	return &PrivacyService{
		uc: uc,
	}
}

func (s *PrivacyService) GetPrivacy(ctx context.Context, req *utils_v1.EmptyRequest) (*v1.PrivacyReply, error) {
	actorID := auth.GetActorIdFromContext(ctx)
	if actorID == 0 {
		return nil, v1.ErrorEmptyActorId("empty actor id")
	}

	settings, err := s.uc.GetPrivacy(ctx, actorID)
	if err != nil {
		return nil, v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}

	return &v1.PrivacyReply{
		Settings: settings,
	}, nil
}

func (s *PrivacyService) UpdatePrivacy(ctx context.Context, req *v1.PrivacyRequest) (*v1.PrivacyReply, error) {
	actorID := auth.GetActorIdFromContext(ctx)
	if actorID == 0 {
		return nil, v1.ErrorEmptyActorId("empty actor id")
	}

	settings, err := s.uc.UpdatePrivacy(ctx, actorID, req.GetSettings())
	if err != nil {
		if ent.IsValidationError(err) {
			return nil, v1.ErrorInvalidRequest("invalid request: %s", err.Error())
		}
		return nil, v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}

	return &v1.PrivacyReply{
		Settings: settings,
	}, nil
}

func (s *PrivacyService) GetUsersPrivacies(
	ctx context.Context,
	req *v1.UsersPrivaciesRequest,
) (*v1.UsersPrivaciesReply, error) {
	settings, err := s.uc.GetPrivacies(ctx, req.GetIds())
	if err != nil {
		return nil, err
	}

	return &v1.UsersPrivaciesReply{
		Users: settings,
	}, nil
}
