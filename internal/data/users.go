package data

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/ent/user"
)

type UpdateUserDto struct {
	Phone    string
	Email    string
	Name     string
	Bio      string
	Avatar   string
	Timezone string
}
type GetUserFilterDto struct {
	WithRelation   bool
	WithContact    bool
	WithMembership bool
	UserId         int64
	Phone          string
	Email          string
}

type GetUsersFilterDto struct {
	UsersIds []int64
	Phones   []string
	Emails   []string
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
	GetUsers(ctx context.Context, filter GetUsersFilterDto) ([]*ent.User, error)
	PhoneVerified(ctx context.Context, userId int64) error
	EmailVerified(ctx context.Context, userId int64) error
}

type usersRepo struct {
	db *ent.Client
}

// NewUsersRepo .
func NewUsersRepo(d *Data, logger log.Logger) UsersRepo {
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
	if dto.Bio != "" && dto.Bio != user.Bio { // unnecessary to finish the registration
		shouldUpdate = true
		query.SetBio(dto.Bio)
		query.SetBioUpdatedAt(now)
	}
	if dto.Avatar != "" && dto.Avatar != *user.Avatar { // unnecessary to finish the registration
		shouldUpdate = true
		query.SetAvatar(dto.Avatar)
	}
	if dto.Timezone != "" { // !required to finish the registration
		shouldUpdate = true
		query.SetTimezone(dto.Timezone).SetIsActive(true)
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

func (r *usersRepo) GetUsers(ctx context.Context, filter GetUsersFilterDto) ([]*ent.User, error) {
	return r.db.User.Query().Where(
		user.Or(
			user.IDIn(filter.UsersIds...),
			user.PhoneIn(filter.Phones...),
			user.EmailIn(filter.Emails...),
		)).
		All(ctx)
}

func (r *usersRepo) PhoneVerified(ctx context.Context, userId int64) error {
	return r.db.User.UpdateOneID(userId).
		Where(
			user.PhoneVerified(false),
		).
		SetPhoneVerified(true).
		Exec(ctx)
}

func (r *usersRepo) EmailVerified(ctx context.Context, userId int64) error {
	return r.db.User.UpdateOneID(userId).
		Where(
			user.EmailVerified(false),
		).
		SetEmailVerified(true).
		Exec(ctx)
}
