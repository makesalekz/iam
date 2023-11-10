package biz

import (
	"context"
	_ "embed"
	"slices"

	contacts_v1 "contacts/api/contacts/v1"
	v1 "iam/api/iam/v1"
	"iam/ent"
	"iam/internal/data"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
)

type UserItem struct {
	*ent.User

	Relation *v1.Relation
	Contact  *v1.Contact
}

// UsersUsecase .
type UsersUsecase struct {
	log       *log.Helper
	discovery registry.Discovery
	usersRepo data.UsersRepo
	otpRepo   data.OtpRepo
	dialer    *data.Dialer
}

// NewUsersUsecase .
func NewUsersUsecase(logger log.Logger,
	c *data.Config,
	jwt *data.JwtProcessor,
	usersRepo data.UsersRepo,
	otpRepo data.OtpRepo,
	dialer *data.Dialer,
) (*UsersUsecase, error) {
	return &UsersUsecase{
		log:       log.NewHelper(logger),
		discovery: c.GetRegistry(),
		usersRepo: usersRepo,
		otpRepo:   otpRepo,
		dialer:    dialer,
	}, nil
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
		return nil, v1.ErrorInvalidRequest("Invalid request, please read documentations")
	}

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, v1.ErrorUserNotFound("User not found: %v", err)
		}
		return nil, v1.ErrorDatabaseQuery("Internal error")
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

	return replyUser, nil
}

func (uc *UsersUsecase) UpdateUserProfile(ctx context.Context, userId int64, data data.UpdateUserDto) (*UserItem, error) {
	user, err := uc.usersRepo.UpdateUserData(ctx, userId, data)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, v1.ErrorUserNotFound("User not found: %v", err)
		}
		return nil, v1.ErrorDatabaseQuery("Internal error")
	}

	return &UserItem{
		User: user,
	}, nil
}

func (uc *UsersUsecase) DeleteUser(ctx context.Context, userId int64) error {
	return uc.usersRepo.DeleteUser(ctx, userId)
}

func (uc *UsersUsecase) GetUsers(ctx context.Context, filter data.GetUsersFilterDto) ([]*UserItem, error) {
	users, err := uc.usersRepo.GetUsers(ctx, filter)
	if err != nil {
		return nil, v1.ErrorDatabaseQuery("Internal error")
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

func (uc *UsersUsecase) getUserContactLabel(ctx context.Context, userId int64) (*v1.Contact, error) {
	contactClient, err := uc.dialer.Contacts(ctx)
	if err != nil {
		return nil, v1.ErrorGrpcConnection("dialer.Users: %s", err.Error())
	}

	labels, err := contactClient.GetLabelsByUserId(ctx, &contacts_v1.GetLabelsByUserIdRequest{UserId: userId})
	if err != nil {
		if contacts_v1.IsNotFound(err) {
			return nil, v1.ErrorContactNotFound("there is not such contact")
		}
		return nil, v1.ErrorGrpcConnection("Contact user internal error")
	}

	contact := &v1.Contact{}
	if len(labels.GetLabels()) == 0 {
		return nil, v1.ErrorContactNotFound("there is not such contact")
	}

	label := slices.MaxFunc(labels.GetLabels(), func(a, b string) int { return len(a) - len(b) })
	contact.Label = label

	return contact, nil
}

func (uc *UsersUsecase) includeRelations(ctx context.Context, users ...*UserItem) error {
	userIds := make([]int64, len(users))
	for i, user := range users {
		userIds[i] = user.User.ID
	}

	relations, err := uc.getUsersRelations(ctx, userIds)
	if err != nil {
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
		return nil, v1.ErrorGrpcConnection("dialer.Users: %s", err.Error())
	}

	relations, err := relationsClient.GetRelations(ctx, &contacts_v1.GetRelationsRequest{UserIds: userIds})
	if err != nil {
		if contacts_v1.IsNotFound(err) {
			return nil, v1.ErrorContactNotFound("there is not such relation")
		}
		return nil, v1.ErrorGrpcConnection("dialer.Users: %s", err.Error())
	}

	return relations.GetRelations(), nil
}
