package biz

import (
	"context"
	"slices"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	chats_v1 "gitlab.calendaria.team/services/chats/api/chats/v1"
	contacts_v1 "gitlab.calendaria.team/services/contacts/api/contacts/v1"
	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/internal/data"
	tenants_v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
	"gitlab.calendaria.team/services/utils/v1/config"
	"gitlab.calendaria.team/services/utils/v1/jwt"
)

type UserItem struct {
	*ent.User

	Relation   *v1.Relation
	Contact    *v1.Contact
	CommonChat *v1.CommonChat
}

// UsersUsecase .
type UsersUsecase struct {
	log       *log.Helper
	discovery registry.Discovery
	jwt       *jwt.JwtProcessor
	usersRepo data.UsersRepo
	otpRepo   data.OtpRepo
	chats     *data.ChatsRemote
	contacts  *data.ContactsRemote
	tenants   *data.TenantsRemote
}

// NewUsersUsecase .
func NewUsersUsecase(logger log.Logger,
	c *config.Config,
	jwt *jwt.JwtProcessor,
	usersRepo data.UsersRepo,
	otpRepo data.OtpRepo,
	chats *data.ChatsRemote,
	contacts *data.ContactsRemote,
	tenants *data.TenantsRemote,
) (*UsersUsecase, error) {
	return &UsersUsecase{
		log:       log.NewHelper(logger),
		discovery: c.GetRegistry(),
		jwt:       jwt,
		usersRepo: usersRepo,
		otpRepo:   otpRepo,
		chats:     chats,
		contacts:  contacts,
		tenants:   tenants,
	}, nil
}

func (uc *UsersUsecase) getUserContactLabel(ctx context.Context, userId int64) (*v1.Contact, error) {
	labels, err := uc.contacts.GetLabelsByUserId(ctx, &contacts_v1.GetLabelsByUserIdRequest{UserId: userId})
	if err != nil {
		if contacts_v1.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	if len(labels.GetLabels()) == 0 {
		return nil, nil
	}

	contact := &v1.Contact{Label: slices.MaxFunc(labels.GetLabels(), func(a, b string) int { return len(a) - len(b) })}

	return contact, nil
}

func (uc *UsersUsecase) getChatMembership(ctx context.Context, userId int64) (*chats_v1.Membership, error) {
	chatMembership, err := uc.chats.GetDirectChatMembership(ctx, &chats_v1.DirectChatMembershipRequest{UserId: userId})
	if err != nil {
		if chats_v1.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return chatMembership.GetMembership(), nil
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
		return err
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

	if filter.WithContact {
		contactLabel, err := uc.getUserContactLabel(ctx, user.ID)
		if err != nil {
			return nil, err
		}

		replyUser.Contact = contactLabel
	}

	if filter.WithRelation {
		err = uc.includeRelations(ctx, replyUser)
		if err != nil {
			return replyUser, err
		}
	}

	if filter.WithMembership {
		membership, err := uc.getChatMembership(ctx, user.ID)
		if err != nil {
			return nil, err
		}

		replyUser.CommonChat = fromChatsToIam(membership)
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

	return replyUsers, nil
}

func (uc *UsersUsecase) GetUserTenants(ctx context.Context) ([]*tenants_v1.Tenant, error) {
	claims, ok := uc.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserRequest() {
		return nil, v1.ErrorUnauthorized("invalid token")
	}

	return uc.tenants.GetUserTenants(ctx, claims)
}

func fromChatsToIam(membership *chats_v1.Membership) *v1.CommonChat {
	if membership == nil {
		return nil
	}

	return &v1.CommonChat{
		ChatId:     membership.ChatId,
		Status:     membership.Status,
		Role:       membership.Role,
		IsPinned:   membership.IsPinned,
		IsMuted:    membership.IsMuted,
		MutedTill:  membership.MutedTill,
		ArchivedAt: membership.ArchivedAt,
		AutoSave:   membership.AutoSave,
	}
}
