package biz

import (
	"context"
	"slices"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
	chats_v1 "gitlab.calendaria.team/services/chats/api/chats/v1"
	contacts_v1 "gitlab.calendaria.team/services/contacts/api/contacts/v1"
	v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/internal/data"
	tenants_v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
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
	dialer    *data.Dialer
	tenants   *data.TenantsRemote
}

// NewUsersUsecase .
func NewUsersUsecase(logger log.Logger,
	c *config.Config,
	jwt *jwt.JwtProcessor,
	usersRepo data.UsersRepo,
	otpRepo data.OtpRepo,
	dialer *data.Dialer,
	tenants *data.TenantsRemote,
) (*UsersUsecase, error) {
	return &UsersUsecase{
		log:       log.NewHelper(logger),
		discovery: c.GetRegistry(),
		jwt:       jwt,
		usersRepo: usersRepo,
		otpRepo:   otpRepo,
		dialer:    dialer,
		tenants:   tenants,
	}, nil
}

func (uc *UsersUsecase) getUserContactLabel(ctx context.Context, userId int64) (*v1.Contact, error) {
	contactClient, err := uc.dialer.Contacts(ctx)
	if err != nil {
		return nil, v1.ErrorGrpcConnection("contacts: %s", err.Error())
	}

	labels, err := contactClient.GetLabelsByUserId(ctx, &contacts_v1.GetLabelsByUserIdRequest{UserId: userId})
	if err != nil {
		if contacts_v1.IsNotFound(err) {
			return nil, v1.ErrorContactNotFound("there is not such contact")
		}
		return nil, v1.ErrorGrpcConnection("contacts: %s", err.Error())
	}

	contact := &v1.Contact{}
	if len(labels.GetLabels()) == 0 {
		return nil, v1.ErrorContactNotFound("there is not such contact")
	}

	label := slices.MaxFunc(labels.GetLabels(), func(a, b string) int { return len(a) - len(b) })
	contact.Label = label

	return contact, nil
}

func (uc *UsersUsecase) getChatMembership(ctx context.Context, userId int64) (*chats_v1.Membership, error) {
	membersClient, err := uc.dialer.Members(ctx)
	if err != nil {
		return nil, v1.ErrorGrpcConnection("chats: %s", err.Error())
	}

	chatMembership, err := membersClient.GetDirectChatMembership(ctx, &chats_v1.DirectChatMembershipRequest{UserId: userId})
	if err != nil {
		if chats_v1.IsNotFound(err) {
			return nil, v1.ErrorCommonChatNotFound("there is not such chat")
		}
		return nil, v1.ErrorGrpcConnection("chats: %s", err.Error())
	}

	return chatMembership.GetMembership(), nil
}

func (uc *UsersUsecase) includeRelations(ctx context.Context, users ...*UserItem) error {
	userIds := make([]int64, len(users))
	for i, user := range users {
		userIds[i] = user.User.ID
	}

	relations, err := uc.getUsersRelations(ctx, userIds)
	if err != nil {
		if v1.IsRelationNotFound(err) {
			return nil
		}

		return err
	}

	relationMap := make(map[int64]*contacts_v1.Relation)
	for _, relation := range relations {
		relationMap[relation.GetUserId()] = relation
	}

	for _, user := range users {
		relation, ok := relationMap[user.User.ID]
		if !ok {
			continue
		}

		user.Relation = &v1.Relation{
			IsBlocked: relation.GetIsBlocked(),
			IsMuted:   relation.GetIsMuted(),
		}
	}

	return nil
}

func (uc *UsersUsecase) getUsersRelations(ctx context.Context, userIds []int64) ([]*contacts_v1.Relation, error) {
	relationsClient, err := uc.dialer.Relations(ctx)
	if err != nil {
		return nil, v1.ErrorGrpcConnection("contacts: %s", err.Error())
	}

	relations, err := relationsClient.GetRelations(ctx, &contacts_v1.GetRelationsRequest{UserIds: userIds})
	if err != nil {
		if contacts_v1.IsNotFound(err) {
			return nil, v1.ErrorRelationNotFound("there is not such relation")
		}
		return nil, v1.ErrorGrpcConnection("contacts: %s", err.Error())
	}

	return relations.GetRelations(), nil
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
		contactLabel, err := uc.getUserContactLabel(ctx, filter.UserId)
		if err != nil {
			if !v1.IsContactNotFound(err) {
				return nil, err
			}
		} else {
			replyUser.Contact = &v1.Contact{Label: contactLabel.Label}
		}
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
			if !v1.IsCommonChatNotFound(err) {
				return nil, err
			}
		} else {
			replyUser.CommonChat = data.FromChatsToIam(membership)
		}
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

func (uc *UsersUsecase) GetUsers(ctx context.Context, filter data.GetUsersFilterDto) ([]*UserItem, error) {
	users, err := uc.usersRepo.GetUsers(ctx, filter)
	if err != nil {
		return nil, v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}

	replyUsers := make([]*UserItem, len(users))
	for i, user := range users {
		replyUsers[i] = &UserItem{User: user}
	}

	return replyUsers, nil
}

func (uc *UsersUsecase) GetUserTenants(ctx context.Context) ([]*tenants_v1.Tenant, error) {
	claims, ok := uc.jwt.GetClaimsFromContext(ctx)
	if !ok || !claims.IsUserRequest() {
		return nil, v1.ErrorUnauthorized("invalid token")
	}

	uc.log.Debug("GetUserTenants: ", "claims", claims)

	tenantClient, err := uc.tenants.Tenants(ctx, claims)
	if err != nil {
		return nil, v1.ErrorGrpcConnection("tenants: %s", err.Error())
	}

	tenants, err := tenantClient.ListTenants(ctx, &tenants_v1.ListTenantsRequest{})
	if err != nil {
		return nil, v1.ErrorGrpcConnection("tenants: %s", err.Error())
	}

	return tenants.Tenants, nil
}
