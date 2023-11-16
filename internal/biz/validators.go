package biz

import (
	"net/mail"
	"time"

	"github.com/nyaruka/phonenumbers"
	v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
)

func ParseEmail(email string) (*mail.Address, error) {
	address, err := mail.ParseAddress(email)
	if err != nil {
		return nil, v1.ErrorInvalidEmail("invalid email: %s", email)
	}
	return address, nil
}

func ParsePhone(phone string) (string, error) {
	if phone == "" {
		return "", v1.ErrorInvalidPhoneNumber("phone is empty")
	}

	phoneNumber, err := phonenumbers.Parse(phone, DEFAULT_REGION)
	if err != nil {
		return "", v1.ErrorInvalidPhoneNumber("parse error: %s", err.Error())
	}
	if !phonenumbers.IsValidNumber(phoneNumber) {
		return "", v1.ErrorInvalidPhoneNumber("invalid phone number: %s", phone)
	}

	return phonenumbers.Format(phoneNumber, phonenumbers.E164), nil
}

func CheckTimezone(tz string) error {
	if tz == "Local" {
		return v1.ErrorInvalidTimezone("invalid timezone (%s)", tz)
	}

	_, err := time.LoadLocation(tz)
	if err != nil {
		return v1.ErrorInvalidTimezone("invalid timezone (%s): %s", tz, err.Error())
	}

	return nil
}
