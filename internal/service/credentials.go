package service

import (
	"context"
	"time"

	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/internal/biz"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
	"gitlab.calendaria.team/services/utils/v2/auth"
	u_struc "gitlab.calendaria.team/services/utils/v2/struc"

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

func userCredentialToV1Credential(userCredentials *ent.UserCredentials) *iam_v1.UserCredential {
	if userCredentials == nil {
		return &iam_v1.UserCredential{}
	}

	replyUserCredentials := &iam_v1.UserCredential{
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

func userCredentialToV1CredentialShort(userCredentials *ent.UserCredentials) *iam_v1.UserCredentialShort {
	if userCredentials == nil {
		return &iam_v1.UserCredentialShort{}
	}

	replyUserCredentials := &iam_v1.UserCredentialShort{
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

func userCredentialsToV1CredentialShorts(userCredentials []*ent.UserCredentials) []*iam_v1.UserCredentialShort {
	trUserCredentials := make([]*iam_v1.UserCredentialShort, len(userCredentials))
	for i, userCredential := range userCredentials {
		trUserCredentials[i] = userCredentialToV1CredentialShort(userCredential)
	}

	return trUserCredentials
}

func (s *CredentialsService) AuthByGoogle(
	ctx context.Context,
	req *iam_v1.AuthByGoogleRequest,
) (*utils_v1.EmptyReply, error) {
	actorID := auth.GetActorIdFromContext(ctx)
	if actorID == 0 {
		return nil, iam_v1.ErrorEmptyActorId("empty actor id")
	}

	err := s.uc.AuthByGoogle(ctx, actorID, req.GetAuthCode())
	if err != nil {
		return nil, err
	}

	return &utils_v1.EmptyReply{}, nil
}

func (s *CredentialsService) RefreshCredential(
	ctx context.Context,
	req *iam_v1.CredentialRequest,
) (*iam_v1.CredentialReply, error) {
	actorID := auth.GetActorIdFromContext(ctx)
	if actorID == 0 {
		return nil, iam_v1.ErrorEmptyActorId("empty actor id")
	}

	credential, err := s.uc.RefreshCredential(ctx, actorID, req.GetCredentialId())
	if err != nil {
		return nil, err
	}

	return &iam_v1.CredentialReply{Credential: userCredentialToV1Credential(credential)}, nil
}

func (s *CredentialsService) GetCredential(
	ctx context.Context,
	req *iam_v1.CredentialRequest,
) (*iam_v1.CredentialReply, error) {
	actorID := auth.GetActorIdFromContext(ctx)
	if actorID == 0 {
		return nil, iam_v1.ErrorEmptyActorId("empty actor id")
	}

	credential, err := s.uc.GetCredential(ctx, actorID, req.GetCredentialId())
	if err != nil {
		return nil, err
	}

	return &iam_v1.CredentialReply{Credential: userCredentialToV1Credential(credential)}, nil
}

func (s *CredentialsService) ListCredentials(
	ctx context.Context,
	req *iam_v1.ListCredentialsRequest,
) (*iam_v1.ListCredentialsReply, error) {
	actorID := auth.GetActorIdFromContext(ctx)
	if actorID == 0 {
		return nil, iam_v1.ErrorEmptyActorId("empty actor id")
	}

	var provider *u_struc.Provider
	if req.GetProvider() != "" {
		p := u_struc.Provider(req.GetProvider())
		if !p.IsValid() {
			return nil, iam_v1.ErrorInvalidProvider("invalid provider")
		}
		provider = &p
	}

	credentials, err := s.uc.ListCredentials(ctx, actorID, provider)
	if err != nil {
		return nil, err
	}

	return &iam_v1.ListCredentialsReply{Credentials: userCredentialsToV1CredentialShorts(credentials)}, nil
}

func (s *CredentialsService) DeleteCredential(
	ctx context.Context,
	req *iam_v1.CredentialRequest,
) (*utils_v1.EmptyReply, error) {
	actorID := auth.GetActorIdFromContext(ctx)
	if actorID == 0 {
		return nil, iam_v1.ErrorEmptyActorId("empty actor id")
	}

	err := s.uc.DeleteCredential(ctx, actorID, req.GetCredentialId())
	if err != nil {
		return nil, err
	}

	return &utils_v1.EmptyReply{}, nil
}
