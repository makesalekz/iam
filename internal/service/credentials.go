package service

import (
	"context"
	"gitlab.calendaria.team/services/iam/ent"
	"time"

	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/internal/biz"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
	"gitlab.calendaria.team/services/utils/v2/auth"

	"github.com/go-kratos/kratos/v2/log"
)

type CredentialsService struct {
	iam_v1.UnimplementedCredentialsServer

	log *log.Helper
	uc  *biz.CredentialsUsecase
}

func NewCredentialsService(
	logger log.Logger,
	uc *biz.CredentialsUsecase,
) *CredentialsService {
	return &CredentialsService{
		log: log.NewHelper(log.With(logger, "module", "service/credentials")),
		uc:  uc,
	}
}

func userCredentialsToV1Credentials(userCredentials *ent.UserCredentials) *iam_v1.UserCredentials {
	if userCredentials == nil {
		return &iam_v1.UserCredentials{}
	}

	replyUserCredentials := &iam_v1.UserCredentials{
		Id:           userCredentials.ID,
		UserId:       userCredentials.UserID,
		Mail:         userCredentials.Mail,
		DisplayName:  userCredentials.DisplayName,
		AccessToken:  userCredentials.AccessToken,
		TokenType:    userCredentials.TokenType,
		RefreshToken: userCredentials.RefreshToken,
	}

	if userCredentials.Provider != nil {
		provider := userCredentials.Provider.Value()
		replyUserCredentials.Provider = &provider
	}

	if userCredentials.ExpiresAt != nil {
		expiresAt := userCredentials.ExpiresAt.Format(time.RFC3339)
		replyUserCredentials.ExpiresAt = &expiresAt
	}

	return replyUserCredentials
}

func (s *CredentialsService) AuthByGoogle(ctx context.Context, req *iam_v1.AuthByGoogleRequest) (*utils_v1.EmptyReply, error) {
	actorId := auth.GetActorIdFromContext(ctx)
	if actorId == 0 {
		return nil, iam_v1.ErrorEmptyActorId("empty actor id")
	}

	err := s.uc.AuthByGoogle(ctx, actorId, req.AuthCode)
	if err != nil {
		return nil, err
	}

	return &utils_v1.EmptyReply{}, nil
}

func (s *CredentialsService) GetOwnCredentials(ctx context.Context, _ *utils_v1.EmptyRequest) (*iam_v1.CredentialsReply, error) {
	actorId := auth.GetActorIdFromContext(ctx)
	if actorId == 0 {
		return nil, iam_v1.ErrorEmptyActorId("empty actor id")
	}

	credentials, err := s.uc.GetOwnCredentials(ctx, actorId)
	if err != nil {
		return nil, err
	}

	return &iam_v1.CredentialsReply{Credential: userCredentialsToV1Credentials(credentials)}, nil
}

func (s *CredentialsService) DeleteCredentials(ctx context.Context, req *iam_v1.CredentialsRequest) (*utils_v1.EmptyReply, error) {
	actorId := auth.GetActorIdFromContext(ctx)
	if actorId == 0 {
		return nil, iam_v1.ErrorEmptyActorId("empty actor id")
	}

	err := s.uc.DeleteCredentials(ctx, actorId, req.CredentialId)
	if err != nil {
		return nil, err
	}

	return &utils_v1.EmptyReply{}, nil
}
