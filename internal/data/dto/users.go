package dto

import (
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"time"

	v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/ent"

	"github.com/nyaruka/phonenumbers"
)

const (
	MinLength = 3
	MaxLength = 30
)

type UpdateUserDto struct {
	UserID   int64
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

func NewUpdateUserDto(userID int64, req *v1.UpdateOwnProfileRequest) *UpdateUserDto {
	return &UpdateUserDto{
		UserID:   userID,
		Phone:    req.GetPhone(),
		Email:    req.GetEmail(),
		Name:     req.GetName(),
		Username: req.GetUsername(),
		Bio:      req.Bio, //nolint:protogetter // optional field, acquired by ref
		Avatar:   req.GetAvatar(),
		Timezone: req.GetTimezone(),
	}
}

func (dto *UpdateUserDto) Validate() (err error) {
	if dto.Phone != "" {
		dto.Phone, err = parsePhone(dto.Phone)
		if err != nil {
			return err
		}
	}

	var email *mail.Address
	if dto.Email != "" {
		email, err = parseEmail(dto.Email)
		if err != nil {
			return err
		}
		dto.Email = email.Address
	}

	if dto.Timezone != "" {
		err = checkTimezone(dto.Timezone)
		if err != nil {
			return err
		}
	}

	if !isValidUsername(dto.UserID, dto.Username) {
		return v1.ErrorInvalidUsername("invalid username format")
	}

	return nil
}

func isValidUsername(actorID int64, username string) bool {
	// 1. Check if username is empty
	if username == "" {
		return false
	}

	// 2. Check combined validation for length and allowed characters
	// This handles the specific error message requirement
	if len(username) > MaxLength || !regexp.MustCompile(`^[a-z][a-z0-9_]+$`).MatchString(username) {
		return false
	}

	// 3. Check minimum length separately (still required)
	if len(username) < MinLength {
		return false
	}

	// 4. Check reserved username format (user{number})
	// Only allow if the number matches the actor's ID
	if len(username) > 4 && strings.ToLower(username[:4]) == "user" {
		userID, err := strconv.ParseInt(username[4:], 10, 64)
		if err == nil && userID != actorID {
			return false
		}
	}

	return true
}

func parseEmail(email string) (*mail.Address, error) {
	address, err := mail.ParseAddress(email)
	if err != nil {
		return nil, v1.ErrorInvalidEmail("invalid email: %s", email)
	}
	return address, nil
}

func parsePhone(phone string) (string, error) {
	if phone == "" {
		return "", v1.ErrorInvalidPhoneNumber("phone is empty")
	}

	phoneNumber, err := phonenumbers.Parse(phone, DefaultRegion)
	if err != nil {
		return "", v1.ErrorInvalidPhoneNumber("parse error: %s", err.Error())
	}
	if !phonenumbers.IsValidNumber(phoneNumber) {
		return "", v1.ErrorInvalidPhoneNumber("invalid phone number: %s", phone)
	}

	return phonenumbers.Format(phoneNumber, phonenumbers.E164), nil
}

func checkTimezone(tz string) error {
	if tz == "Local" {
		return v1.ErrorInvalidTimezone("invalid timezone (%s)", tz)
	}

	_, err := time.LoadLocation(tz)
	if err != nil {
		return v1.ErrorInvalidTimezone("invalid timezone (%s): %s", tz, err.Error())
	}

	return nil
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
	} else if dto.user.Avatar != nil {
		dto.shouldUpdate = true
		dto.updateQuery.ClearAvatar()
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
