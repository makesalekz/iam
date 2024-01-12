package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/internal/biz"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
)

type PrivacyService struct {
	v1.UnimplementedPrivacyServer

	log *log.Helper
	sh  *ServiceHelper
	uc  *biz.PrivacyUsecase
}

func NewPrivacyService(
	logger log.Logger,
	sh *ServiceHelper,
	uc *biz.PrivacyUsecase,
) *PrivacyService {
	return &PrivacyService{
		log: log.NewHelper(logger),
		sh:  sh,
		uc:  uc,
	}
}

func (s *PrivacyService) GetPrivacy(ctx context.Context, req *utils_v1.ActorRequest) (*v1.PrivacyReply, error) {
	actorId, err := s.sh.GetActorId(ctx, req.ActorId)
	if err != nil {
		return nil, err
	}

	settings, err := s.uc.GetPrivacy(ctx, actorId)
	if err != nil {
		return nil, v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}

	return &v1.PrivacyReply{
		Settings: settings,
	}, nil
}

func (s *PrivacyService) UpdatePrivacy(ctx context.Context, req *v1.PrivacyRequest) (*v1.PrivacyReply, error) {
	actorId, err := s.sh.GetActorId(ctx, req.ActorId)
	if err != nil {
		return nil, err
	}

	settings, err := s.uc.UpdatePrivacy(ctx, actorId, req.Settings)
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

func (s *PrivacyService) GetUsersPrivacies(ctx context.Context, req *v1.UsersPrivaciesRequest) (*v1.UsersPrivaciesReply, error) {
	settings, err := s.uc.GetPrivacies(ctx, req.Ids)
	if err != nil {
		return nil, err
	}

	return &v1.UsersPrivaciesReply{
		Users: settings,
	}, nil
}
