package dto

import (
	"fmt"
	"math/rand"
	"os"

	v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/utils/v2/struc"

	"github.com/nyaruka/phonenumbers"
)

const (
	DefaultRegion = "KZ"
	OtpLength     = 6
	DebugOtpCode  = "777333"

	digits            = "0123456789"
	verifiablePhone   = "+77710012030"
	verifiableOtpCode = "667423"
)

type AuthPhoneDto struct {
	AppID                struc.ApplicationID
	Phone                string
	IsRegistrationNeeded bool
	IsRegistration       bool
	AppSignature         string
	code                 string
}

func NewAuthPhoneDto(appID string, req *v1.AuthByPhoneRequest) *AuthPhoneDto {
	return &AuthPhoneDto{
		AppID:                struc.ApplicationID(appID),
		Phone:                req.GetPhone(),
		IsRegistrationNeeded: req.GetIsRegistrationNeeded(),
		IsRegistration:       req.GetIsRegistration(),
		AppSignature:         req.GetAppSignature(),
	}
}

func (dto *AuthPhoneDto) Validate() error {
	if !dto.AppID.IsValid() {
		return v1.ErrorInvalidRequest("invalid app id: %s", dto.AppID)
	}

	phoneNumber, err := phonenumbers.Parse(dto.Phone, DefaultRegion)
	if err != nil {
		return v1.ErrorInvalidPhoneNumber("parse error: %s", err.Error())
	}
	if !phonenumbers.IsValidNumber(phoneNumber) {
		return v1.ErrorInvalidPhoneNumber("invalid phone number: %s", dto.Phone)
	}

	dto.Phone = phonenumbers.Format(phoneNumber, phonenumbers.E164)

	return nil
}

func (dto *AuthPhoneDto) GenerateCode() string {
	switch {
	case dto.Phone == verifiablePhone:
		// use fixed code for verifiable phone
		dto.code = verifiableOtpCode

	case os.Getenv("DEBUG") != "":
		// use fixed code in debug mode
		dto.code = DebugOtpCode

	default:
		dto.code = GenerateRandomNumber(OtpLength)
	}

	return dto.code
}

func (dto *AuthPhoneDto) GetOtpMessage() string {
	sms := fmt.Sprintf("%s Kod: %s\n%s", dto.AppID.BrandName(), dto.code, dto.AppID.CompanyFullName())

	if dto.AppSignature != "" {
		sms = fmt.Sprintf("%s\n%s", sms, dto.AppSignature)
	}

	return sms
}

func GenerateRandomNumber(n int) string {
	result := make([]byte, n)
	for i := range result {
		//nolint:gosec // we don't need cryptographically secure random numbers here
		result[i] = digits[rand.Int63()%int64(len(digits))]
	}
	return string(result)
}
