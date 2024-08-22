package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/ent/enum"
	"gitlab.calendaria.team/services/iam/internal/biz"
	"gitlab.calendaria.team/services/iam/internal/data/mock"
	nats_mock "gitlab.calendaria.team/services/utils/v1/nats/mock"
	jwt_mock "gitlab.calendaria.team/services/utils/v2/jwt/mock"
	"gitlab.calendaria.team/services/utils/v2/zap"
)

func TestAuthSuccess(t *testing.T) {
	logger := zap.NewZapLogger(true)
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	usersRepo := mock.NewMockUsersRepo(ctrl)
	otpRepo := mock.NewMockOtpRepo(ctrl)
	tenantRemote := mock.NewMockITenantsRemote(ctrl)
	notificationsRemote := mock.NewMockINotificationsRemote(ctrl)
	jwt := jwt_mock.NewMockIJwtProcessor(ctrl)
	queue := nats_mock.NewMockIQueueManager(ctrl)

	authUseCase, err := biz.NewAuthUsecase(logger, jwt, usersRepo, otpRepo, queue, tenantRemote, notificationsRemote)
	require.NoError(t, err)

	ctx := context.Background()

	phone := "+77777777777"
	var userID int64 = 123
	otp := &ent.OneTimePassword{
		Code: "777333",
	}

	expectedUser := &ent.User{
		ID: userID,
	}

	usersRepo.EXPECT().GetUserByPhone(gomock.Any(), phone, true).Return(expectedUser, nil)
	otpRepo.EXPECT().CreateOneTimePassword(gomock.Any(), userID, enum.Phone, "777333", 5*time.Minute).Return(otp, nil)
	notificationsRemote.EXPECT().PersonalSmsSender(gomock.Any(), phone, "Calendaria: 777333").Return(nil)

	_, err = authUseCase.AuthUserByPhone(ctx, phone, false, false)
	require.NoError(t, err)
}

func TestUserNotExist(t *testing.T) {
	logger := zap.NewZapLogger(true)
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	usersRepo := mock.NewMockUsersRepo(ctrl)
	otpRepo := mock.NewMockOtpRepo(ctrl)
	tenantRemote := mock.NewMockITenantsRemote(ctrl)
	notificationsRemote := mock.NewMockINotificationsRemote(ctrl)
	jwt := jwt_mock.NewMockIJwtProcessor(ctrl)
	queue := nats_mock.NewMockIQueueManager(ctrl)

	authUseCase, err := biz.NewAuthUsecase(logger, jwt, usersRepo, otpRepo, queue, tenantRemote, notificationsRemote)
	require.NoError(t, err)

	ctx := context.Background()

	phone := "+77777777777"
	var userID int64 = 123
	otp := &ent.OneTimePassword{
		Code: "777333",
	}

	expectedUser := &ent.User{
		ID: userID,
	}

	usersRepo.EXPECT().GetUserByPhone(gomock.Any(), phone, true).Return(nil, &ent.NotFoundError{})
	usersRepo.EXPECT().CreateUserWithPhone(gomock.Any(), phone).Return(expectedUser, nil)
	otpRepo.EXPECT().CreateOneTimePassword(gomock.Any(), userID, enum.Phone, "777333", 5*time.Minute).Return(otp, nil)
	notificationsRemote.EXPECT().PersonalSmsSender(gomock.Any(), phone, "Calendaria: 777333").Return(nil)

	_, err = authUseCase.AuthUserByPhone(ctx, phone, false, false)
	require.NoError(t, err)
}

func TestUserRegistrationRequired(t *testing.T) {
	logger := zap.NewZapLogger(true)
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	usersRepo := mock.NewMockUsersRepo(ctrl)
	otpRepo := mock.NewMockOtpRepo(ctrl)
	tenantRemote := mock.NewMockITenantsRemote(ctrl)
	notificationsRemote := mock.NewMockINotificationsRemote(ctrl)
	jwt := jwt_mock.NewMockIJwtProcessor(ctrl)
	queue := nats_mock.NewMockIQueueManager(ctrl)

	authUseCase, err := biz.NewAuthUsecase(logger, jwt, usersRepo, otpRepo, queue, tenantRemote, notificationsRemote)
	require.NoError(t, err)

	ctx := context.Background()

	phone := "+77777777777"

	usersRepo.EXPECT().GetUserByPhone(gomock.Any(), phone, true).Return(nil, &ent.NotFoundError{})
	_, err = authUseCase.AuthUserByPhone(ctx, phone, true, false)
	require.Error(t, err)
}

func TestUserRegistration(t *testing.T) {
	logger := zap.NewZapLogger(true)
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	usersRepo := mock.NewMockUsersRepo(ctrl)
	otpRepo := mock.NewMockOtpRepo(ctrl)
	tenantRemote := mock.NewMockITenantsRemote(ctrl)
	notificationsRemote := mock.NewMockINotificationsRemote(ctrl)
	jwt := jwt_mock.NewMockIJwtProcessor(ctrl)
	queue := nats_mock.NewMockIQueueManager(ctrl)

	authUseCase, err := biz.NewAuthUsecase(logger, jwt, usersRepo, otpRepo, queue, tenantRemote, notificationsRemote)
	require.NoError(t, err)

	ctx := context.Background()

	phone := "+77777777777"

	var userID int64 = 123
	otp := &ent.OneTimePassword{
		Code: "777333",
	}

	expectedUser := &ent.User{
		ID: userID,
	}

	usersRepo.EXPECT().GetUserByPhone(gomock.Any(), phone, true).Return(nil, &ent.NotFoundError{})
	usersRepo.EXPECT().CreateUserWithPhone(gomock.Any(), phone).Return(expectedUser, nil)
	otpRepo.EXPECT().CreateOneTimePassword(gomock.Any(), userID, enum.Phone, "777333", 5*time.Minute).Return(otp, nil)
	notificationsRemote.EXPECT().PersonalSmsSender(gomock.Any(), phone, "Calendaria: 777333").Return(nil)

	_, err = authUseCase.AuthUserByPhone(ctx, phone, false, true)
	require.NoError(t, err)
}
