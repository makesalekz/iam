
package service_test

import (
	"context"
	"strconv"
	"testing"

	"gitlab.calendaria.team/services/iam/internal/biz"
	"gitlab.calendaria.team/services/iam/internal/data/mock"
	"gitlab.calendaria.team/services/iam/internal/service"
	u_nats_mock "gitlab.calendaria.team/services/utils/v1/nats/mock"
	jwt_mock "gitlab.calendaria.team/services/utils/v2/jwt/mock"
	u_zap "gitlab.calendaria.team/services/utils/v2/zap"

	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

type dataMock struct {
	qm                  *u_nats_mock.MockIQueueManager
	queue               *u_nats_mock.MockIQueue
	jwt                 *jwt_mock.MockIJwtProcessor
	usersRepo           *mock.MockUsersRepo
	otpRepo             *mock.MockOtpRepo
	tenantsRemote       *mock.MockITenantsRemote
	notificationsRemote *mock.MockINotificationsRemote
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
		qm:                  qm,
		queue:               queue,
		jwt:                 jwt,
		usersRepo:           usersRepo,
		otpRepo:             otpRepo,
		tenantsRemote:       tenantRemote,
		notificationsRemote: notificationsRemote,
	}

	// create service
	eu, err := biz.NewAuthUsecase(logger, jwt, qm, tenantRemote, notificationsRemote, usersRepo, otpRepo)
	require.NoError(t, err)

	return ctx, repo, service.NewAuthService(eu)
}
