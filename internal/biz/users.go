package biz

import (
	"context"
	"errors"

	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/internal/data"
	tenants_v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
	u_error "gitlab.calendaria.team/services/utils/v1/error"
	"gitlab.calendaria.team/services/utils/v1/jwt"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/lib/pq"
)

type UserItem struct {
	*ent.User

	Privacies    map[string]string
	WithVerified bool
}

// UsersUsecase .
type UsersUsecase struct {
	jwt           *jwt.JwtProcessor
	usersRepo     data.UsersRepo
	otpRepo       data.OtpRepo
	privaciesRepo data.PrivacyRepo
	tenants       *data.TenantsRemote
}

type ConstraintKey string

const (
	USERNAME ConstraintKey = "users_username_key"
	EMAIL    ConstraintKey = "users_email_key"
	PHONE    ConstraintKey = "users_phone_key"
)

// NewUsersUsecase .
func NewUsersUsecase(logger log.Logger,
	jwt *jwt.JwtProcessor,
	usersRepo data.UsersRepo,
	otpRepo data.OtpRepo,
	privaciesRepo data.PrivacyRepo,
	tenants *data.TenantsRemote,
) (*UsersUsecase, error) {
	return &UsersUsecase{
		jwt:           jwt,
		usersRepo:     usersRepo,
		otpRepo:       otpRepo,
		privaciesRepo: privaciesRepo,
		tenants:       tenants,
	}, nil
}

func (uc *UsersUsecase) includePrivacies(ctx context.Context, users ...*UserItem) error {
	userIds := make([]int64, len(users))
	for i, user := range users {
		userIds[i] = user.ID
	}

	usersPrivacies, err := uc.privaciesRepo.GetPrivacies(ctx, userIds)
	if err != nil {
		return iam_v1.ErrorServiceFailed("privacy: %s", err.Error())
	}

	privaciesMap := make(map[int64]map[string]string)
	for _, userPrivacies := range usersPrivacies {
		if privaciesMap[userPrivacies.UserID] == nil {
			privaciesMap[userPrivacies.UserID] = data.DefaultPrivacies()
		}
		privaciesMap[userPrivacies.UserID][string(userPrivacies.Setting)] = string(userPrivacies.Option)
	}

	for _, user := range users {
		privacy, ok := privaciesMap[user.ID]
		if !ok {
			user.Privacies = data.DefaultPrivacies()

			continue
		}

		user.Privacies = privacy
	}

	return nil
}

func (uc *UsersUsecase) GetUserProfile(ctx context.Context, filter data.GetUserFilterDto) (*UserItem, error) {
	var user *ent.User
	var err error

	if filter.Phone != "" && filter.Email == "" && filter.UserId == 0 {
		user, err = uc.usersRepo.GetUserByPhone(ctx, filter.Phone)
	} else if filter.Email != "" && filter.Phone == "" && filter.UserId == 0 {
		user, err = uc.usersRepo.GetUserByEmail(ctx, filter.Email)
	} else if filter.UserId != 0 && filter.Email == "" && filter.Phone == "" {
		user, err = uc.usersRepo.GetUserById(ctx, filter.UserId)
	} else {
		return nil, v1.ErrorInvalidRequest("invalid request")
	}

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, v1.ErrorUserNotFound("user not found")
		}
		return nil, v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}
	replyUser := &UserItem{
		User: user,
	}

	return replyUser, nil
}

func (uc *UsersUsecase) UpdateUserProfile(ctx context.Context, userId int64, dto data.UpdateUserDto) (*UserItem, error) {
	var err error

	if dto.Phone != "" {
		dto.Phone, err = ParsePhone(dto.Phone)
		if err != nil {
			return nil, err
		}
	}

	if dto.Email != "" {
		email, err := ParseEmail(dto.Email)
		if err != nil {
			return nil, err
		}
		dto.Email = email.Address
	}

	if dto.Timezone != "" {
		err = CheckTimezone(dto.Timezone)
		if err != nil {
			return nil, err
		}
	}

	user, err := uc.usersRepo.GetUserById(ctx, userId)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, v1.ErrorUserNotFound("user not found")
		}
		return nil, v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}

	updatedUser, err := uc.usersRepo.UpdateUserData(ctx, user, dto)
	if err != nil {
		if u_error.IsUniqueViolation(err) {
			var pqError *pq.Error
			ok := errors.As(err, &pqError)
			if !ok {
				return nil, v1.ErrorDatabaseQuery("database error: %s", err.Error())
			}

			switch pqError.Constraint {
			case string(USERNAME):
				return nil, v1.ErrorInvalidUsername("user with such username already exists")
			case string(EMAIL):
				return nil, v1.ErrorInvalidEmail("user with such email already exists")
			case string(PHONE):
				return nil, v1.ErrorInvalidPhoneNumber("user with such phone number already exists")
			default:
				return nil, v1.ErrorInvalidRequest("some user details are already exists")
			}
		}
		return nil, v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}

	return &UserItem{
		User: updatedUser,
	}, nil
}

func (uc *UsersUsecase) DeleteUser(ctx context.Context, userId int64) error {
	err := uc.usersRepo.DeleteUser(ctx, userId)
	if err != nil {
		_, notFound := err.(*ent.NotFoundError)
		if notFound {
			return v1.ErrorUserNotFound("user not found")
		}
		return v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}
	return nil
}

func (uc *UsersUsecase) ListUsers(ctx context.Context, filter data.GetUsersFilterDto, sort *utils_v1.SortRequest, paginate *utils_v1.PaginateRequest) ([]*UserItem, error) {
	if paginate == nil {
		paginate = &utils_v1.PaginateRequest{}
	}

	users, err := uc.usersRepo.ListUsers(ctx, filter, sort, paginate)
	if err != nil {
		return nil, v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}

	replyUsers := make([]*UserItem, len(users))
	for i, user := range users {
		replyUsers[i] = &UserItem{User: user}
	}

	return replyUsers, nil
}

func (uc *UsersUsecase) GetUsers(ctx context.Context, actorId int64, filter data.GetUsersFilterDto) ([]*UserItem, error) {
	users, err := uc.usersRepo.GetUsers(ctx, filter)
	if err != nil {
		return nil, v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}

	replyUsers := make([]*UserItem, len(users))
	for i, user := range users {
		replyUsers[i] = &UserItem{User: user}
	}

	if filter.WithPrivacies {
		err = uc.includePrivacies(ctx, replyUsers...)
		if err != nil {
			return nil, err
		}
	}

	if filter.WithVerified {
		for i := 0; i < len(users); i++ {
			replyUsers[i].WithVerified = filter.WithVerified
		}
	}

	return replyUsers, nil
}

func (uc *UsersUsecase) GetUserTenants(ctx context.Context) ([]*tenants_v1.Tenant, error) {
	tenants, err := uc.tenants.GetUserTenants(ctx)
	if err != nil {
		return nil, tenants_v1.ErrorServiceFailed("tenants: %s", err.Error())
	}

	return tenants, nil
}
