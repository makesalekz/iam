package data

import (
	"context"

	"iam/ent"

	_ "github.com/lib/pq"
)

// UsersRepo
type UsersRepo interface {
	CreateUser(context.Context) (*ent.User, error)
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

func (r *usersRepo) CreateUser(ctx context.Context) (*ent.User, error) {
	user, err := r.db.User.Create().Save(ctx)

	return user, err
}
