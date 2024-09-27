package data

import (
	"time"

	"gitlab.calendaria.team/services/iam/ent"
)

type UpdateUserDto struct {
	Phone    string
	Email    string
	Name     string
	Username string
	Bio      *string
	Avatar   string
	Timezone string
	TenantID int64

	shouldUpdate bool
	updateQuery  *ent.UserUpdateOne
	user         *ent.User
}

func (dto *UpdateUserDto) ForQuery(query *ent.UserUpdateOne) *UpdateUserDto {
	dto.updateQuery = query

	return dto
}

func (dto *UpdateUserDto) ForUser(user *ent.User) *UpdateUserDto {
	dto.user = user

	return dto
}
func (dto *UpdateUserDto) ApplyPhone() *UpdateUserDto {
	// TODO: allow to update verified phone and email, using additional tables
	if dto.Phone != "" && !dto.user.PhoneVerified { // update only if phone is not verified
		if dto.user.Phone == nil || *dto.user.Phone != dto.Phone { // check if new phone is different from the old one
			dto.shouldUpdate = true
			dto.updateQuery.SetPhone(dto.Phone)
		}
	}

	return dto
}

func (dto *UpdateUserDto) ApplyEmail() *UpdateUserDto {
	if dto.Email != "" && !dto.user.EmailVerified { // update only if email is not verified
		if dto.user.Email == nil || *dto.user.Email != dto.Email { // check if new phone is different from the old one
			dto.shouldUpdate = true
			dto.updateQuery.SetEmail(dto.Email)
		}
	}

	return dto
}

func (dto *UpdateUserDto) ApplyName() *UpdateUserDto {
	if dto.Name != "" && dto.Name != dto.user.Name { // unnecessary to finish the registration
		dto.shouldUpdate = true
		dto.updateQuery.SetName(dto.Name)
	}

	return dto
}

func (dto *UpdateUserDto) ApplyUsername() *UpdateUserDto {
	// unnecessary to finish the registration
	if dto.Username != "" && (dto.user.Username == nil || dto.Username != *dto.user.Username) {
		dto.shouldUpdate = true
		dto.updateQuery.SetUsername(dto.Username)
	}

	return dto
}

func (dto *UpdateUserDto) ApplyBio() *UpdateUserDto {
	if dto.Bio != nil && *dto.Bio != dto.user.Bio { // unnecessary to finish the registration
		dto.shouldUpdate = true
		dto.updateQuery.SetBio(*dto.Bio).SetBioUpdatedAt(time.Now())
	}

	return dto
}

func (dto *UpdateUserDto) ApplyAvatar() *UpdateUserDto {
	if dto.Avatar != "" { // unnecessary to finish the registration
		if dto.user.Avatar == nil || *dto.user.Avatar != dto.Avatar { // check if new phone is different from the old one
			dto.shouldUpdate = true
			dto.updateQuery.SetAvatar(dto.Avatar)
		}
	}

	return dto
}

func (dto *UpdateUserDto) ApplyTimezone() *UpdateUserDto {
	if dto.Timezone != "" { // !required to finish the registration
		dto.shouldUpdate = true
		dto.updateQuery.SetTimezone(dto.Timezone).SetIsActive(true)
	}

	return dto
}

func (dto *UpdateUserDto) ApplyTenantID() *UpdateUserDto {
	if dto.TenantID != 0 {
		dto.shouldUpdate = true
		dto.updateQuery.SetDefaultTenantID(dto.TenantID)
	}

	return dto
}

func (dto *UpdateUserDto) GetQuery() *ent.UserUpdateOne {
	return dto.updateQuery
}

func (dto *UpdateUserDto) ShouldUpdate() bool {
	return dto.shouldUpdate
}
