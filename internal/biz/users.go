package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	contacts_v1 "gitlab.calendaria.team/services/contacts/api/contacts/v1"
	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/internal/data"
	tenants_v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
	"gitlab.calendaria.team/services/utils/v1/jwt"
)

type UserItem struct {
	*ent.User

	Relation     *v1.Relation
	Privacies    map[string]string
	WithVerified bool
}

// UsersUsecase .
type UsersUsecase struct {
	jwt           *jwt.JwtProcessor
	usersRepo     data.UsersRepo
	otpRepo       data.OtpRepo
	privaciesRepo data.PrivacyRepo
	contacts      *data.ContactsRemote
	tenants       *data.TenantsRemote
}

// NewUsersUsecase .
func NewUsersUsecase(logger log.Logger,
	jwt *jwt.JwtProcessor,
	usersRepo data.UsersRepo,
	otpRepo data.OtpRepo,
	privaciesRepo data.PrivacyRepo,
	contacts *data.ContactsRemote,
	tenants *data.TenantsRemote,
) (*UsersUsecase, error) {
	return &UsersUsecase{
		jwt:           jwt,
		usersRepo:     usersRepo,
		otpRepo:       otpRepo,
		privaciesRepo: privaciesRepo,
		contacts:      contacts,
		tenants:       tenants,
	}, nil
}

func (uc *UsersUsecase) includeRelations(ctx context.Context, users ...*UserItem) error {
	userIds := make([]int64, len(users))
	for i, user := range users {
		userIds[i] = user.ID
	}

	relationsReply, err := uc.contacts.GetRelations(ctx, &contacts_v1.GetRelationsRequest{UserIds: userIds})
	if err != nil {
		if contacts_v1.IsNotFound(err) {
			return nil
		}
		return iam_v1.ErrorServiceFailed("contacts: %s", err.Error())
	}

	relations := relationsReply.GetRelations()
	if relations == nil {
		return nil
	}

	relationMap := make(map[int64]*contacts_v1.Relation)
	for _, relation := range relations {
		relationMap[relation.GetUserId()] = relation
	}

	for _, user := range users {
		relation, ok := relationMap[user.ID]
		if !ok {
			continue
		}

		user.Relation = &iam_v1.Relation{
			IsBlocked: relation.GetIsBlocked(),
			IsMuted:   relation.GetIsMuted(),
		}
	}

	return nil
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
			privaciesMap[userPrivacies.UserID] = make(map[string]string)
		}
		privaciesMap[userPrivacies.UserID][string(userPrivacies.Setting)] = string(userPrivacies.Option)
	}

	for _, user := range users {
		privacy, ok := privaciesMap[user.ID]
		if !ok {
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

func (uc *UsersUsecase) GetUsers(ctx context.Context, filter data.GetUsersFilterDto, sort *utils_v1.SortRequest, paginate *utils_v1.PaginateRequest) ([]*UserItem, error) {
	if paginate == nil {
		paginate = &utils_v1.PaginateRequest{}
	}

	users, err := uc.usersRepo.GetUsers(ctx, filter, sort, paginate)
	if err != nil {
		return nil, v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}

	replyUsers := make([]*UserItem, len(users))
	for i, user := range users {
		replyUsers[i] = &UserItem{User: user}
	}

	if filter.WithRelation {
		err = uc.includeRelations(ctx, replyUsers...)
		if err != nil {
			return nil, err
		}
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
	claims, ok := uc.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserRequest() {
		return nil, v1.ErrorUnauthorized("invalid token")
	}

	tenants, err := uc.tenants.GetUserTenants(ctx, claims)
	if err != nil {
		return nil, tenants_v1.ErrorServiceFailed("tenants: %s", err.Error())
	}

	return tenants, nil
}
