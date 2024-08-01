package biz

import (
	"context"
	"errors"

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

	phoneMask  = 0b001
	emailMask  = 0b010
	userIDMask = 0b100
)

// NewUsersUsecase .
func NewUsersUsecase(
	logger log.Logger,
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

func btoi(b bool) int64 {
	if b {
		return 1
	}

	return 0
}

func (uc *UsersUsecase) includePrivacies(ctx context.Context, users ...*UserItem) error {
	userIDs := make([]int64, len(users))
	for i, user := range users {
		userIDs[i] = user.ID
	}

	usersPrivacies, err := uc.privaciesRepo.GetPrivacies(ctx, userIDs)
	if err != nil {
		return v1.ErrorServiceFailed("privacy: %s", err.Error())
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

	switch btoi(filter.Phone != "") |
		btoi(filter.Email != "")<<1 |
		btoi(filter.UserID != 0)<<2 {
	case phoneMask:
		user, err = uc.usersRepo.GetUserByPhone(ctx, filter.Phone)
	case emailMask:
		user, err = uc.usersRepo.GetUserByEmail(ctx, filter.Email)
	case userIDMask:
		user, err = uc.usersRepo.GetUserByID(ctx, filter.UserID)
	default:
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

func (uc *UsersUsecase) UpdateUserProfile(ctx context.Context, userID int64, dto data.UpdateUserDto) (
	*UserItem, error,
) {
	var err error

	user, err := uc.usersRepo.GetUserByID(ctx, userID)
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

func (uc *UsersUsecase) DeleteUser(ctx context.Context, userID int64) error {
	err := uc.usersRepo.DeleteUser(ctx, userID)
	if err != nil {
		if ent.IsNotFound(err) {
			return v1.ErrorUserNotFound("user not found")
		}
		return v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}
	return nil
}

func (uc *UsersUsecase) ListUsers(
	ctx context.Context, filter data.GetUsersFilterDto, sort *utils_v1.SortRequest, paginate *utils_v1.PaginateRequest,
) ([]*UserItem, error) {
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

func (uc *UsersUsecase) GetUsers(ctx context.Context, actorID int64, filter data.GetUsersFilterDto) (
	[]*UserItem, error,
) {
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
