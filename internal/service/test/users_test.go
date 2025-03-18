package test_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/internal/data/dto"

	"github.com/lib/pq"
)

func TestUsersService_UpdateOwnProfile_SuccessCases(t *testing.T) {
	ctx, repo, usersService := createUsersService(t)
	ids := getIDs()

	// Success Case: Valid profile update
	{
		req := &v1.UpdateOwnProfileRequest{
			Email:    "user@example.com",
			Phone:    "+77071231212",
			Username: "valid_username",
			Timezone: "Europe/Moscow",
		}

		updatedUser := &ent.User{
			ID:       ids.actorID,
			Email:    &req.Email,
			Phone:    &req.Phone,
			Username: &req.Username,
			Timezone: req.Timezone,
		}

		repo.usersRepo.EXPECT().
			GetUserByID(ctx, ids.actorID, false).
			Return(updatedUser, nil)

		userDto := dto.NewUpdateUserDto(ids.actorID, req)
		err := userDto.Validate()
		require.NoError(t, err)

		repo.usersRepo.EXPECT().
			UpdateUserData(ctx, updatedUser, *userDto).
			Return(updatedUser, nil)

		result, err := usersService.UpdateOwnProfile(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, ids.actorID, result.User.Id)
		require.Equal(t, req.Email, *result.User.Email)
		require.Equal(t, req.Phone, *result.User.Phone)
		require.Equal(t, req.Username, *result.User.Username)
		require.Equal(t, req.Timezone, result.User.Timezone)
	}
}

func TestUsersService_UpdateOwnProfile_ErrorCases(t *testing.T) {
	ctx, repo, usersService := createUsersService(t)
	ids := getIDs()

	// Error Case 1: Empty Actor ID
	{
		ctxNoActor := context.Background()
		req := &v1.UpdateOwnProfileRequest{
			Username: "validusername",
		}

		result, err := usersService.UpdateOwnProfile(ctxNoActor, req)
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, v1.ErrorEmptyActorId("empty actor id"), err)
	}

	// Error Case 2: Invalid Email
	{
		req := &v1.UpdateOwnProfileRequest{
			Email: "invalid-email",
		}

		result, err := usersService.UpdateOwnProfile(ctx, req)
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, v1.ErrorInvalidEmail("invalid email: invalid-email"), err)
	}

	// Error Case 3: Invalid Phone
	{
		req := &v1.UpdateOwnProfileRequest{
			Phone: "+1231212",
		}

		result, err := usersService.UpdateOwnProfile(ctx, req)
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "invalid phone number")
	}

	// Error Case 4: Invalid Timezone
	{
		req := &v1.UpdateOwnProfileRequest{
			Timezone: "Invalid/Timezone",
		}

		result, err := usersService.UpdateOwnProfile(ctx, req)
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "invalid timezone")
	}

	// Error Case 5: Invalid Username
	{
		req := &v1.UpdateOwnProfileRequest{
			Username: "Invalid-Username!",
		}

		result, err := usersService.UpdateOwnProfile(ctx, req)
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, v1.ErrorInvalidUsername("invalid username format"), err)
	}

	// Error Case 6: GetUserByID NotFound
	{
		req := &v1.UpdateOwnProfileRequest{
			Username: "validusername",
		}

		repo.usersRepo.EXPECT().
			GetUserByID(ctx, ids.actorID, false).
			Return(nil, &ent.NotFoundError{})

		result, err := usersService.UpdateOwnProfile(ctx, req)
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, v1.ErrorUserNotFound("user not found"), err)
	}

	// Error Case 7: GetUserByID DB Error
	{
		req := &v1.UpdateOwnProfileRequest{
			Username: "validusername",
		}

		repo.usersRepo.EXPECT().
			GetUserByID(ctx, ids.actorID, false).
			Return(nil, errors.New("db error"))

		result, err := usersService.UpdateOwnProfile(ctx, req)
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, v1.ErrorDatabaseQuery("database error: db error"), err)
	}

	// Error Case 8: UpdateUserData Unique Violation Username
	{
		req := &v1.UpdateOwnProfileRequest{
			Username: "existingusername",
		}

		user := &ent.User{ID: ids.actorID}

		repo.usersRepo.EXPECT().
			GetUserByID(ctx, ids.actorID, false).
			Return(user, nil)

		repo.usersRepo.EXPECT().
			UpdateUserData(ctx, user, dto.UpdateUserDto{
				UserID:   ids.actorID,
				Username: req.Username,
			}).
			Return(nil, &pq.Error{Code: "23505", Constraint: "users_username_key"})

		result, err := usersService.UpdateOwnProfile(ctx, req)
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, v1.ErrorInvalidUsername("user with such username already exists"), err)
	}

	// Error Case 9: UpdateUserData General DB Error
	{
		req := &v1.UpdateOwnProfileRequest{
			Username: "newusername",
		}

		user := &ent.User{ID: ids.actorID}

		repo.usersRepo.EXPECT().
			GetUserByID(ctx, ids.actorID, false).
			Return(user, nil)

		repo.usersRepo.EXPECT().
			UpdateUserData(ctx, user, dto.UpdateUserDto{
				UserID:   ids.actorID,
				Username: req.Username,
			}).
			Return(nil, errors.New("general db error"))

		result, err := usersService.UpdateOwnProfile(ctx, req)
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, v1.ErrorDatabaseQuery("database error: general db error"), err)
	}

	// Error Case 10: UpdateUserData Unique Violation Email
	{
		req := &v1.UpdateOwnProfileRequest{
			Username: "some_user",
			Email:    "existingemail@example.com",
		}

		user := &ent.User{ID: ids.actorID}

		repo.usersRepo.EXPECT().
			GetUserByID(ctx, ids.actorID, false).
			Return(user, nil)

		repo.usersRepo.EXPECT().
			UpdateUserData(ctx, user, dto.UpdateUserDto{
				Username: req.Username,
				UserID:   ids.actorID,
				Email:    req.Email,
			}).
			Return(nil, &pq.Error{Code: "23505", Constraint: "users_email_key"})

		result, err := usersService.UpdateOwnProfile(ctx, req)
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, v1.ErrorInvalidEmail("user with such email already exists"), err)
	}

	// Error Case 10: UpdateUserData Unique Violation Email
	{
		req := &v1.UpdateOwnProfileRequest{
			Username: "some_user",
			Phone:    "+77071231212",
		}

		user := &ent.User{ID: ids.actorID}

		repo.usersRepo.EXPECT().
			GetUserByID(ctx, ids.actorID, false).
			Return(user, nil)

		repo.usersRepo.EXPECT().
			UpdateUserData(ctx, user, dto.UpdateUserDto{
				UserID:   ids.actorID,
				Username: req.Username,
				Phone:    req.Phone,
			}).
			Return(nil, &pq.Error{Code: "23505", Constraint: "users_phone_key"})

		result, err := usersService.UpdateOwnProfile(ctx, req)
		require.Error(t, err)
		require.Nil(t, result)
		require.Equal(t, v1.ErrorInvalidPhoneNumber("user with such phone number already exists"), err)
	}
}
