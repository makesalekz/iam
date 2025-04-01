package test_test

import (
	"context"
	"strconv"
	"testing"

	"gitlab.calendaria.team/services/iam/internal/biz"
	"gitlab.calendaria.team/services/iam/internal/data/mock"
	"gitlab.calendaria.team/services/iam/internal/service"
	u_zap "gitlab.calendaria.team/services/utils/v2/zap"
	u_config_mock "gitlab.calendaria.team/services/utils/v4/config/mock"
	jwt_mock "gitlab.calendaria.team/services/utils/v4/jwt/mock"
	u_nats_mock "gitlab.calendaria.team/services/utils/v4/nats/mock"

	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

type dataMock struct {
	config          *u_config_mock.MockIConfig
	qm              *u_nats_mock.MockIQueueManager
	queue           *u_nats_mock.MockIQueue
	jwt             *jwt_mock.MockIJwtProcessor
	usersRepo       *mock.MockUsersRepo
	otpRepo         *mock.MockOtpRepo
	privacyRepo     *mock.MockPrivacyRepo
	credentialsRepo *mock.MockCredentialsRepo
	chats           *mock.MockIChatsRemote
	contacts        *mock.MockIContactsRemote
	events          *mock.MockIEventsRemote
	media           *mock.MockIMediaRemote
	tenants         *mock.MockITenantsRemote
	notifications   *mock.MockINotificationsRemote
	provider        *mock.MockIProviderManager
	providerGateway *mock.MockIProviderGateway
}

type idCollection struct {
	actorID   int64 // actorID is 111
	actor2ID  int64 // actor2ID is 1112
	ownerID   int64 // ownerID is 777
	owner2ID  int64 // owner2ID is 7772
	adminID   int64 // adminID is 999
	admin2ID  int64 // admin2ID is 9992
	tenantID  int64 // tenantID is 444
	tenant2ID int64 // tenant2ID is 4442
	userID    int64 // userID is 123
	user2ID   int64 // user2ID is 1232
}

func getIDs() idCollection {
	return idCollection{
		actorID:   111,
		actor2ID:  1112,
		ownerID:   777,
		owner2ID:  7772,
		adminID:   999,
		admin2ID:  9992,
		tenantID:  444,
		tenant2ID: 4442,
		userID:    123,
		user2ID:   1232,
	}
}

func mockServerContext() context.Context {
	ids := getIDs()
	md := metadata.Metadata{
		"x-md-global-tenant-id":  []string{strconv.Itoa(int(ids.tenantID))},
		"x-md-global-actor-id":   []string{strconv.Itoa(int(ids.actorID))},
		"x-md-global-identities": []string{"identity1", "identity2"},
		"x-md-global-app-id":     []string{"app-id"},
	}
	return metadata.NewServerContext(context.Background(), md)
}

func createAuthService(t *testing.T) (context.Context, *dataMock, *service.AuthService) {
	// create context
	ctx := mockServerContext()

	// create controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create logger
	logger := u_zap.NewZapLogger(true)

	// create mocks
	qm := u_nats_mock.NewMockIQueueManager(ctrl)
	queue := u_nats_mock.NewMockIQueue(ctrl)
	jwt := jwt_mock.NewMockIJwtProcessor(ctrl)
	usersRepo := mock.NewMockUsersRepo(ctrl)
	otpRepo := mock.NewMockOtpRepo(ctrl)
	tenantRemote := mock.NewMockITenantsRemote(ctrl)
	notificationsRemote := mock.NewMockINotificationsRemote(ctrl)

	// collect repo
	repo := &dataMock{
		qm:            qm,
		queue:         queue,
		jwt:           jwt,
		usersRepo:     usersRepo,
		otpRepo:       otpRepo,
		tenants:       tenantRemote,
		notifications: notificationsRemote,
	}

	// create service
	eu, err := biz.NewAuthUsecase(logger, jwt, qm, tenantRemote, notificationsRemote, usersRepo, otpRepo)
	require.NoError(t, err)

	return ctx, repo, service.NewAuthService(eu)
}

func createCredentialsService(t *testing.T) (context.Context, *dataMock, *service.CredentialsService) {
	// create context
	ctx := mockServerContext()

	// create controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create logger
	logger := u_zap.NewZapLogger(true)

	// create mocks
	config := u_config_mock.NewMockIConfig(ctrl)
	qm := u_nats_mock.NewMockIQueueManager(ctrl)
	queue := u_nats_mock.NewMockIQueue(ctrl)
	jwt := jwt_mock.NewMockIJwtProcessor(ctrl)
	provider := mock.NewMockIProviderManager(ctrl)
	providerGateway := mock.NewMockIProviderGateway(ctrl)
	events := mock.NewMockIEventsRemote(ctrl)
	credentialsRepo := mock.NewMockCredentialsRepo(ctrl)

	// collect repo
	repo := &dataMock{
		config:          config,
		qm:              qm,
		queue:           queue,
		jwt:             jwt,
		provider:        provider,
		providerGateway: providerGateway,
		events:          events,
		credentialsRepo: credentialsRepo,
	}

	// create service
	eu, err := biz.NewCredentialsUsecase(
		config,
		logger,
		qm,
		jwt,
		provider,
		events,
		credentialsRepo,
	)
	require.NoError(t, err)

	return ctx, repo, service.NewCredentialsService(logger, eu)
}

func createUsersService(t *testing.T) (context.Context, *dataMock, *service.UsersService) {
	// create context
	ctx := mockServerContext()

	// create controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// create logger
	logger := u_zap.NewZapLogger(true)

	// create mocks
	qm := u_nats_mock.NewMockIQueueManager(ctrl)
	queue := u_nats_mock.NewMockIQueue(ctrl)
	jwt := jwt_mock.NewMockIJwtProcessor(ctrl)
	tenants := mock.NewMockITenantsRemote(ctrl)
	contacts := mock.NewMockIContactsRemote(ctrl)
	chats := mock.NewMockIChatsRemote(ctrl)
	events := mock.NewMockIEventsRemote(ctrl)
	media := mock.NewMockIMediaRemote(ctrl)
	usersRepo := mock.NewMockUsersRepo(ctrl)
	otpRepo := mock.NewMockOtpRepo(ctrl)
	privacyRepo := mock.NewMockPrivacyRepo(ctrl)

	// collect repo
	repo := &dataMock{
		qm:          qm,
		queue:       queue,
		jwt:         jwt,
		tenants:     tenants,
		contacts:    contacts,
		chats:       chats,
		events:      events,
		media:       media,
		usersRepo:   usersRepo,
		otpRepo:     otpRepo,
		privacyRepo: privacyRepo,
	}

	// create service
	eu, err := biz.NewUsersUsecase(
		logger,
		jwt,
		qm,
		tenants,
		contacts,
		chats,
		events,
		media,
		usersRepo,
		otpRepo,
		privacyRepo,
	)
	require.NoError(t, err)

	return ctx, repo, service.NewUsersService(logger, eu)
}
