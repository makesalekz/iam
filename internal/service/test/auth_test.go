package test_test

import (
	"context"
	"testing"
	"time"

	v1 "github.com/makesalekz/iam/api/iam/v1"
	"github.com/makesalekz/iam/ent"
	"github.com/makesalekz/iam/ent/enum"
	"github.com/makesalekz/iam/internal/biz"
	"github.com/makesalekz/iam/internal/data/dto"
	"github.com/makesalekz/iam/internal/data/mock"
	tenants_v1 "github.com/makesalekz/tenants/api/tenants/v1"
	jwt_mock "github.com/makesalekz/utils/v4/jwt/mock"
	nats_mock "github.com/makesalekz/utils/v4/nats/mock"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/makesalekz/utils/v2/zap"
)

const (
	aigendaAppID    = "AIgenda"
	calendariaAppID = "calendaria"
	smsCode         = "777333"
	smsText         = "AIgenda Kod: 777333\nTOO \"AXIO\""
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

	authUseCase, err := biz.NewAuthUsecase(logger, jwt, queue, tenantRemote, notificationsRemote, usersRepo, otpRepo)
	require.NoError(t, err)

	ctx := context.Background()

	phone := "+77777777777"
	var userID int64 = 123
	otp := &ent.OneTimePassword{
		Code: smsCode,
	}

	expectedUser := &ent.User{
		ID: userID,
	}

	usersRepo.EXPECT().GetUserByPhone(gomock.Any(), phone, true).Return(expectedUser, nil)
	otpRepo.EXPECT().CreateOneTimePassword(gomock.Any(), userID, enum.Phone, smsCode, 5*time.Minute).Return(otp, nil)
	notificationsRemote.EXPECT().PersonalSmsSender(gomock.Any(), aigendaAppID, phone, smsText).Return(nil)

	_, err = authUseCase.AuthUserByPhone(ctx, &dto.AuthPhoneDto{
		AppID: calendariaAppID,
		Phone: phone,
	})
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

	authUseCase, err := biz.NewAuthUsecase(logger, jwt, queue, tenantRemote, notificationsRemote, usersRepo, otpRepo)
	require.NoError(t, err)

	ctx := context.Background()

	phone := "+77777777777"
	var userID int64 = 123
	otp := &ent.OneTimePassword{
		Code: smsCode,
	}

	expectedUser := &ent.User{
		ID: userID,
	}

	usersRepo.EXPECT().GetUserByPhone(gomock.Any(), phone, true).Return(nil, &ent.NotFoundError{})
	usersRepo.EXPECT().CreateUserWithPhone(gomock.Any(), phone).Return(expectedUser, nil)
	otpRepo.EXPECT().CreateOneTimePassword(gomock.Any(), userID, enum.Phone, smsCode, 5*time.Minute).Return(otp, nil)
	notificationsRemote.EXPECT().PersonalSmsSender(gomock.Any(), aigendaAppID, phone, smsText).Return(nil)

	_, err = authUseCase.AuthUserByPhone(ctx, &dto.AuthPhoneDto{
		AppID: calendariaAppID,
		Phone: phone,
	})
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

	authUseCase, err := biz.NewAuthUsecase(logger, jwt, queue, tenantRemote, notificationsRemote, usersRepo, otpRepo)
	require.NoError(t, err)

	ctx := context.Background()

	phone := "+77777777777"

	usersRepo.EXPECT().GetUserByPhone(gomock.Any(), phone, true).Return(nil, &ent.NotFoundError{})
	_, err = authUseCase.AuthUserByPhone(ctx, &dto.AuthPhoneDto{
		AppID:                calendariaAppID,
		Phone:                phone,
		IsRegistrationNeeded: true,
	})
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

	authUseCase, err := biz.NewAuthUsecase(logger, jwt, queue, tenantRemote, notificationsRemote, usersRepo, otpRepo)
	require.NoError(t, err)

	ctx := context.Background()

	phone := "+77777777777"

	var userID int64 = 123
	otp := &ent.OneTimePassword{
		Code: smsCode,
	}

	expectedUser := &ent.User{
		ID: userID,
	}

	usersRepo.EXPECT().GetUserByPhone(gomock.Any(), phone, true).Return(nil, &ent.NotFoundError{})
	usersRepo.EXPECT().CreateUserWithPhone(gomock.Any(), phone).Return(expectedUser, nil)
	otpRepo.EXPECT().CreateOneTimePassword(gomock.Any(), userID, enum.Phone, smsCode, 5*time.Minute).Return(otp, nil)
	notificationsRemote.EXPECT().PersonalSmsSender(gomock.Any(), aigendaAppID, phone, smsText).Return(nil)

	_, err = authUseCase.AuthUserByPhone(ctx, &dto.AuthPhoneDto{
		AppID:          calendariaAppID,
		Phone:          phone,
		IsRegistration: true,
	})
	require.NoError(t, err)
}

func TestAuthService_RefreshToken(t *testing.T) {
	ctx, repo, authUseCase := createAuthService(t)
	ids := getIDs()

	// create request
	req := &v1.TenantRequest{
		TenantId: ids.tenantID,
	}

	// Success Case 1: Refresh token for active user
	{
		// create user
		user := &ent.User{
			ID:   ids.actorID,
			Name: "tester",
		}
		repo.usersRepo.EXPECT().GetUserByID(ctx, user.ID, false).Return(user, nil)

		// tenant identities reply
		tenantIdentities := &tenants_v1.GetMemberIdentitiesReply{
			Member: "member1",
			Groups: []string{"group1", "group2"},
		}
		repo.tenants.EXPECT().GetMemberIdentities(ctx, req.GetTenantId(), user.ID).Return(tenantIdentities, nil)

		// jwt secret
		secret := []byte{1}
		repo.jwt.EXPECT().GetSecret().Return(secret).AnyTimes()

		result, err := authUseCase.RefreshToken(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, result)
	}

	// Error Case 1: Refresh token for inactive or non-existing user
	{
		repo.usersRepo.EXPECT().GetUserByID(ctx, ids.actorID, false).Return(nil, &ent.NotFoundError{})

		result, err := authUseCase.RefreshToken(ctx, req)
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, v1.ErrorUserNotFound("user not found"), err)
	}
}
