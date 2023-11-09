package utils

import (
	v1 "iam/api/iam/v1"
	"iam/ent"
	"time"
)

func UserFromDto(user *ent.User) *v1.User {
	result := &v1.User{
		Id:          user.ID,
		Phone:       user.Phone,
		Email:       user.Email,
		Name:        user.Name,
		Bio:         user.Bio,
		Avatar:      user.Avatar,
		Timezone:    user.Timezone,
		CreatedAt:   user.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   user.UpdatedAt.Format(time.RFC3339),
		LastLoginAt: user.LastLoginAt.Format(time.RFC3339),
		IsActive:    user.IsActive,
	}

	if user.BioUpdatedAt != nil {
		bioUpdatedAt := user.BioUpdatedAt.Format(time.RFC3339)
		result.BioUpdatedAt = &bioUpdatedAt
	}
	return result
}

func UserShortFromDto(user *ent.User) *v1.UserShort {
	result := &v1.UserShort{
		Id:          user.ID,
		Name:        user.Name,
		LastLoginAt: user.LastLoginAt.Format(time.RFC3339),
	}

	if user.Phone != nil {
		result.Phone = *user.Phone
	}
	if user.Email != nil {
		result.Email = *user.Email
	}
	if user.Avatar != nil {
		result.Avatar = *user.Avatar
	}

	return result
}

func UsersShortFromDtos(users []*ent.User) []*v1.UserShort {
	replies := make([]*v1.UserShort, len(users))
	for _, user := range users {
		replies = append(replies, UserShortFromDto(user))
	}

	return replies
}

func UsersFromDtos(users []*ent.User) []*v1.User {
	replies := make([]*v1.User, len(users))
	for _, user := range users {
		replies = append(replies, UserFromDto(user))
	}

	return replies
}

func UserToUserShort(user *v1.User) *v1.UserShort {
	replyUser := &v1.UserShort{}
	replyUser = &v1.UserShort{
		Id:          user.Id,
		Name:        user.Name,
		LastLoginAt: user.LastLoginAt,
	}

	if user.Phone != nil {
		replyUser.Phone = *user.Phone
	}
	if user.Email != nil {
		replyUser.Email = *user.Email
	}
	if user.Avatar != nil {
		replyUser.Avatar = *user.Avatar
	}

	if user.Relation != nil {
		relation := &v1.Relation{}
		relation.IsBlocked = user.GetRelation().IsBlocked
		relation.IsMuted = user.GetRelation().IsMuted

		replyUser.Relation = relation
	}

	return replyUser
}

func UsersToUsersShort(users []*v1.User) []*v1.UserShort {
	replies := make([]*v1.UserShort, len(users))
	for _, user := range users {
		UserToUserShort(user)
	}

	return replies
}
