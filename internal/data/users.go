package data

import (
	"context"
	"time"

	"iam/ent"
	"iam/ent/user"

	"github.com/go-kratos/kratos/v2/log"
)

type UpdateUserDto struct {
	Name     string
	Bio      string
	Avatar   string
	Timezone string
}

// UsersRepo
type UsersRepo interface {
	GetUserById(ctx context.Context, id int64) (*ent.User, error)
	GetUserByPhone(ctx context.Context, phone string) (*ent.User, error)
	GetUserByEmail(ctx context.Context, email string) (*ent.User, error)
	CreateUserWithPhone(ctx context.Context, phone string) (*ent.User, error)
	CreateUserWithEmail(ctx context.Context, email string) (*ent.User, error)
	UpdateUserData(ctx context.Context, id int64, dto UpdateUserDto) (*ent.User, error)
	DeleteUser(ctx context.Context, id int64) error
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

func (r *usersRepo) UpdateUserData(ctx context.Context, id int64, dto UpdateUserDto) (*ent.User, error) {
	user, err := r.db.User.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	shouldUpdate := false
	now := time.Now()
	query := user.Update().SetLastLoginAt(now).SetUpdatedAt(now)

	if dto.Name != "" { // unnecessary to finish the registration
		shouldUpdate = true
		query.SetName(dto.Name)
	}
	if dto.Bio != "" { // unnecessary to finish the registration
		shouldUpdate = true
		query.SetBio(dto.Bio)
	}
	if dto.Avatar != "" { // unnecessary to finish the registration
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

	_, err = query.Save(ctx)

	return user, err
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
