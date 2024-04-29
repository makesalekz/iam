package service

import (
	"context"
	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/ent/property"
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

func userCredentialToV1Credential(userCredentials *ent.UserCredentials) *iam_v1.UserCredentials {
	if userCredentials == nil {
		return &iam_v1.UserCredentials{}
	}

	replyUserCredentials := &iam_v1.UserCredentials{
		Id:           userCredentials.ID,
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

func userCredentialToV1CredentialShort(userCredentials *ent.UserCredentials) *iam_v1.UserCredentialsShort {
	if userCredentials == nil {
		return &iam_v1.UserCredentialsShort{}
	}

	replyUserCredentials := &iam_v1.UserCredentialsShort{
		Id:          userCredentials.ID,
		Mail:        userCredentials.Mail,
		DisplayName: userCredentials.DisplayName,
	}

	if userCredentials.Provider != nil {
		provider := userCredentials.Provider.Value()
		replyUserCredentials.Provider = &provider
	}

	return replyUserCredentials
}

func userCredentialsToV1CredentialShorts(userCredentials []*ent.UserCredentials) []*iam_v1.UserCredentialsShort {
	trUserCredentials := make([]*iam_v1.UserCredentialsShort, len(userCredentials))
	for i, userCredential := range userCredentials {
		trUserCredentials[i] = userCredentialToV1CredentialShort(userCredential)
	}

	return trUserCredentials
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

func (s *CredentialsService) GetCredential(ctx context.Context, req *iam_v1.GetCredentialRequest) (*iam_v1.CredentialReply, error) {
	actorId := auth.GetActorIdFromContext(ctx)
	if actorId == 0 {
		return nil, iam_v1.ErrorEmptyActorId("empty actor id")
	}

	provider := property.Provider(req.Provider)
	if !provider.IsValid() {
		return nil, iam_v1.ErrorInvalidProvider("invalid provider")
	}

	credential, err := s.uc.GetCredential(ctx, actorId, provider)
	if err != nil {
		return nil, err
	}

	return &iam_v1.CredentialReply{Credential: userCredentialToV1Credential(credential)}, nil
}

func (s *CredentialsService) ListCredentials(ctx context.Context, req *utils_v1.EmptyRequest) (*iam_v1.ListCredentialsReply, error) {
	actorId := auth.GetActorIdFromContext(ctx)
	if actorId == 0 {
		return nil, iam_v1.ErrorEmptyActorId("empty actor id")
	}

	credentials, err := s.uc.ListCredentials(ctx, actorId)
	if err != nil {
		return nil, err
	}

	return &iam_v1.ListCredentialsReply{Credentials: userCredentialsToV1CredentialShorts(credentials)}, nil
}

func (s *CredentialsService) DeleteCredential(ctx context.Context, req *iam_v1.CredentialsRequest) (*utils_v1.EmptyReply, error) {
	actorId := auth.GetActorIdFromContext(ctx)
	if actorId == 0 {
		return nil, iam_v1.ErrorEmptyActorId("empty actor id")
	}

	err := s.uc.DeleteCredential(ctx, actorId, req.CredentialId)
	if err != nil {
		return nil, err
	}

	return &utils_v1.EmptyReply{}, nil
}
