package data

import (
	"context"
	"fmt"
	"time"

	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/ent/user"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
)

type UpdateUserDto struct {
	Phone    string
	Email    string
	Name     string
	Username string
	Bio      *string
	Avatar   string
	Timezone string
	TenantId int64
}
type GetUserFilterDto struct {
	UserId int64
	Phone  string
	Email  string
}

type GetUsersFilterDto struct {
	UsersIds      []int64
	Phones        []string
	Emails        []string
	Search        string
	WithRelation  bool
	WithPrivacies bool
	WithVerified  bool
}

// UsersRepo
type UsersRepo interface {
	GetUserById(ctx context.Context, id int64) (*ent.User, error)
	GetUserByPhone(ctx context.Context, phone string) (*ent.User, error)
	GetUserByEmail(ctx context.Context, email string) (*ent.User, error)
	CreateUserWithPhone(ctx context.Context, phone string) (*ent.User, error)
	CreateUserWithEmail(ctx context.Context, email string) (*ent.User, error)
	UpdateUserData(ctx context.Context, user *ent.User, dto UpdateUserDto) (*ent.User, error)
	DeleteUser(ctx context.Context, id int64) error
	ListUsers(ctx context.Context, filter GetUsersFilterDto, sort *utils_v1.SortRequest, paginate *utils_v1.PaginateRequest) ([]*ent.User, error)
	GetUsers(ctx context.Context, filter GetUsersFilterDto) ([]*ent.User, error)
	PhoneVerified(ctx context.Context, userId int64) error
	EmailVerified(ctx context.Context, userId int64) error

	TempGetUsersWithoutDefaultTenant(ctx context.Context) ([]*ent.User, error)
}

type usersRepo struct {
	db *ent.Client
}

// NewUsersRepo .
func NewUsersRepo(d *Data) UsersRepo {
	return &usersRepo{
		db: d.db,
	}
}

func (r *usersRepo) CreateUserWithPhone(ctx context.Context, phone string) (*ent.User, error) {
	return r.db.User.Create().SetPhone(phone).Save(ctx)
}

func (r *usersRepo) CreateUserWithEmail(ctx context.Context, email string) (*ent.User, error) {
	return r.db.User.Create().SetEmail(email).Save(ctx)
}

func (r *usersRepo) UpdateUserData(ctx context.Context, user *ent.User, dto UpdateUserDto) (*ent.User, error) {
	shouldUpdate := false
	now := time.Now()
	query := r.db.User.UpdateOne(user).SetLastLoginAt(now).SetUpdatedAt(now)

	// TODO: allow to update verified phone and email, using additional tables
	if dto.Phone != "" && !user.PhoneVerified { // update only if phone is not verified
		if user.Phone == nil || *user.Phone != dto.Phone { // check if new phone is different from the old one
			shouldUpdate = true
			query.SetPhone(dto.Phone)
		}
	}
	if dto.Email != "" && !user.EmailVerified { // update only if email is not verified
		if user.Email == nil || *user.Email != dto.Email { // check if new phone is different from the old one
			shouldUpdate = true
			query.SetEmail(dto.Email)
		}
	}
	if dto.Name != "" && dto.Name != user.Name { // unnecessary to finish the registration
		shouldUpdate = true
		query.SetName(dto.Name)
	}
	if dto.Username != "" && (user.Username == nil || dto.Username != *user.Username) { // unnecessary to finish the registration
		shouldUpdate = true
		query.SetUsername(dto.Username)
	}
	if dto.Bio != nil && *dto.Bio != user.Bio { // unnecessary to finish the registration
		shouldUpdate = true
		query.SetBio(*dto.Bio).SetBioUpdatedAt(now)
	}
	if dto.Avatar != "" { // unnecessary to finish the registration
		if user.Avatar == nil || *user.Avatar != dto.Avatar { // check if new phone is different from the old one
			shouldUpdate = true
			query.SetAvatar(dto.Avatar)
		}
	}
	if dto.Timezone != "" { // !required to finish the registration
		shouldUpdate = true
		query.SetTimezone(dto.Timezone).SetIsActive(true)
	}

	if dto.TenantId != 0 {
		shouldUpdate = true
		query.SetDefaultTenantID(dto.TenantId)
	}

	if !shouldUpdate {
		return user, nil
	}

	return query.Save(ctx)
}

func (r *usersRepo) DeleteUser(ctx context.Context, id int64) error {
	return r.db.User.DeleteOneID(id).Exec(ctx)
}

func (r *usersRepo) GetUserById(ctx context.Context, id int64) (*ent.User, error) {
	return r.db.User.Query().Where(user.ID(id)).First(ctx)
}

func (r *usersRepo) GetUserByPhone(ctx context.Context, phone string) (*ent.User, error) {
	return r.db.User.Query().Where(user.Phone(phone)).First(ctx)
}

func (r *usersRepo) GetUserByEmail(ctx context.Context, email string) (*ent.User, error) {
	return r.db.User.Query().Where(user.Email(email)).First(ctx)
}

func (r *usersRepo) ListUsers(ctx context.Context, filter GetUsersFilterDto, sort *utils_v1.SortRequest, paginate *utils_v1.PaginateRequest) ([]*ent.User, error) {
	if len(filter.UsersIds) == 0 && len(filter.Phones) == 0 && len(filter.Emails) == 0 {
		return []*ent.User{}, nil
	}

	query := r.db.User.Query().Where(
		user.Or(
			user.IDIn(filter.UsersIds...),
			user.PhoneIn(filter.Phones...),
			user.EmailIn(filter.Emails...),
		))

	if filter.Search != "" {
		query = query.Where(
			user.Or(
				user.PhoneContains(filter.Search),
				user.EmailContainsFold(filter.Search),
				user.NameContainsFold(filter.Search),
			),
		)
	}

	if sort != nil {
		switch sort.Field {
		case "email":
			if sort.Descending {
				query.Order(ent.Desc(user.FieldEmail))
			} else {
				query.Order(ent.Asc(user.FieldEmail))
			}
		case "phone":
			if sort.Descending {
				query.Order(ent.Desc(user.FieldPhone))
			} else {
				query.Order(ent.Asc(user.FieldPhone))
			}
		case "name":
			if sort.Descending {
				query.Order(ent.Desc(user.FieldName))
			} else {
				query.Order(ent.Asc(user.FieldName))
			}
		default: // case "id"
			if sort.Descending {
				query.Order(ent.Desc(user.FieldID))
			} else {
				query.Order(ent.Asc(user.FieldID))
			}
		}
	} else {
		if paginate.FromId != 0 {
			query.Where(user.IDGT(paginate.FromId))
		}

		query.Order(ent.Asc(user.FieldID))
	}

	if paginate.Limit == 0 {
		paginate.Limit = 100
	}

	if paginate.Page != 0 {
		query.Offset(int((paginate.Page - 1) * paginate.Limit))
	}

	return query.Limit(int(paginate.Limit)).All(ctx)
}

func (r *usersRepo) GetUsers(ctx context.Context, filter GetUsersFilterDto) ([]*ent.User, error) {
	if len(filter.UsersIds) == 0 && len(filter.Phones) == 0 && len(filter.Emails) == 0 {
		return []*ent.User{}, nil
	}

	query := r.db.User.Query().Where(
		user.Or(
			user.IDIn(filter.UsersIds...),
			user.PhoneIn(filter.Phones...),
			user.EmailIn(filter.Emails...),
		))

	if filter.Search != "" {
		query = query.Where(
			user.Or(
				user.PhoneContains(filter.Search),
				user.EmailContainsFold(filter.Search),
				user.NameContainsFold(filter.Search),
			),
		)
	}

	return query.All(ctx)
}

func (r *usersRepo) PhoneVerified(ctx context.Context, userId int64) error {
	return r.db.User.UpdateOneID(userId).
		Where(
			user.PhoneVerified(false),
		).
		SetPhoneVerified(true).
		SetUsername(fmt.Sprintf("user%v", userId)).
		Exec(ctx)
}

func (r *usersRepo) EmailVerified(ctx context.Context, userId int64) error {
	return r.db.User.UpdateOneID(userId).
		Where(
			user.EmailVerified(false),
		).
		SetEmailVerified(true).
		SetUsername(fmt.Sprintf("user%v", userId)).
		Exec(ctx)
}

func (r *usersRepo) TempGetUsersWithoutDefaultTenant(ctx context.Context) ([]*ent.User, error) {
	return r.db.User.Query().
		Where(user.DefaultTenantIDIsNil()).
		All(ctx)
}
