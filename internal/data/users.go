package data

import (
	"context"
	"fmt"
	"time"

	"gitlab.calendaria.team/services/iam/ent/mixins"

	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/ent/user"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
)

type GetUserFilterDto struct {
	UserID int64
	Phone  string
	Email  string
}

type GetUsersFilterDto struct {
	UsersIDs      []int64
	Phones        []string
	Emails        []string
	Search        string
	WithRelation  bool
	WithPrivacies bool
	WithVerified  bool
}

// UsersRepo.
type UsersRepo interface {
	GetUserByID(ctx context.Context, id int64, skipRemoveAt bool) (*ent.User, error)
	GetUserByPhone(ctx context.Context, phone string, skipRemoveAt bool) (*ent.User, error)
	GetUserByEmail(ctx context.Context, email string, skipRemoveAt bool) (*ent.User, error)
	CreateUserWithPhone(ctx context.Context, phone string) (*ent.User, error)
	CreateUserWithEmail(ctx context.Context, email string) (*ent.User, error)
	UpdateUserData(ctx context.Context, user *ent.User, dto UpdateUserDto) (*ent.User, error)
	ScheduleUserDeletion(ctx context.Context, id int64, deleteDuration time.Duration) error
	GetUsersToDelete(ctx context.Context) ([]*ent.User, error)
	DeleteUsers(ctx context.Context, usersIDs []int64) error
	ListUsers(
		ctx context.Context,
		filter GetUsersFilterDto,
		sort *utils_v1.SortRequest,
		paginate *utils_v1.PaginateRequest,
	) ([]*ent.User, error)
	GetUsers(ctx context.Context, filter GetUsersFilterDto) ([]*ent.User, error)
	PhoneVerified(ctx context.Context, userID int64) error
	EmailVerified(ctx context.Context, userID int64) error

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
	now := time.Now()
	query := r.db.User.UpdateOne(user).SetLastLoginAt(now).SetUpdatedAt(now)

	query = dto.ForUser(user).
		ForQuery(query).
		ApplyEmail().
		ApplyAvatar().
		ApplyBio().
		ApplyName().
		ApplyTenantID().
		ApplyTimezone().
		ApplyUsername().
		GetQuery()

	if !dto.ShouldUpdate() {
		return user, nil
	}

	return query.Save(ctx)
}

func (r *usersRepo) ScheduleUserDeletion(ctx context.Context, id int64, deleteDuration time.Duration) error {
	return r.db.User.UpdateOneID(id).SetRemoveAt(time.Now().Add(deleteDuration)).Exec(ctx)
}

func (r *usersRepo) GetUsersToDelete(ctx context.Context) ([]*ent.User, error) {
	return r.db.User.Query().Where(user.RemoveAtLTE(time.Now())).All(ctx)
}

func (r *usersRepo) DeleteUsers(ctx context.Context, usersIDs []int64) error {
	return r.db.User.Update().
		Where(user.IDIn(usersIDs...)).
		ClearPhone().
		ClearEmail().
		ClearUsername().
		SetName("").
		SetBio("").
		ClearAvatar().
		SetTimezone("UTC").
		SetIsActive(false).
		SetPhoneVerified(false).
		SetEmailVerified(false).
		SetLastLoginAt(time.Time{}).
		SetCreatedAt(time.Time{}).
		SetUpdatedAt(time.Time{}).
		SetBioUpdatedAt(time.Time{}).
		ClearDefaultTenantID().
		SetDeletedAt(time.Now()).
		Exec(ctx)
}

func (r *usersRepo) GetUserByID(ctx context.Context, id int64, skipRemoveAt bool) (*ent.User, error) {
	if skipRemoveAt {
		ctx = mixins.SkipSoftRemove(ctx)
	}

	return r.db.User.Query().Where(user.ID(id)).First(ctx)
}

func (r *usersRepo) GetUserByPhone(ctx context.Context, phone string, skipRemoveAt bool) (*ent.User, error) {
	if skipRemoveAt {
		ctx = mixins.SkipSoftRemove(ctx)
	}

	return r.db.User.Query().Where(user.Phone(phone)).First(ctx)
}

func (r *usersRepo) GetUserByEmail(ctx context.Context, email string, skipRemoveAt bool) (*ent.User, error) {
	if skipRemoveAt {
		ctx = mixins.SkipSoftRemove(ctx)
	}

	return r.db.User.Query().Where(user.Email(email)).First(ctx)
}

func (r *usersRepo) ListUsers(
	ctx context.Context,
	filter GetUsersFilterDto,
	sort *utils_v1.SortRequest,
	paginate *utils_v1.PaginateRequest,
) ([]*ent.User, error) {
	if len(filter.UsersIDs) == 0 && len(filter.Phones) == 0 && len(filter.Emails) == 0 {
		return []*ent.User{}, nil
	}

	query := r.db.User.Query().Where(
		user.Or(
			user.IDIn(filter.UsersIDs...),
			user.PhoneIn(filter.Phones...),
			user.EmailIn(filter.Emails...),
		),
	)

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
		orderFunc := ent.Asc
		if sort.GetDescending() {
			orderFunc = ent.Desc
		}

		switch sort.GetField() {
		case "email":
			query.Order(orderFunc(user.FieldEmail))
		case "phone":
			query.Order(orderFunc(user.FieldPhone))
		case "name":
			query.Order(orderFunc(user.FieldName))
		default: // case "id"
			query.Order(orderFunc(user.FieldID))
		}
	} else {
		if paginate.GetFromId() != 0 {
			query.Where(user.IDGT(paginate.GetFromId()))
		}

		query.Order(ent.Asc(user.FieldID))
	}

	if paginate.GetLimit() == 0 {
		paginate.Limit = 100
	}

	if paginate.GetPage() != 0 {
		query.Offset(int((paginate.GetPage() - 1) * paginate.GetLimit()))
	}

	return query.Limit(int(paginate.GetLimit())).All(ctx)
}

func (r *usersRepo) GetUsers(ctx context.Context, filter GetUsersFilterDto) ([]*ent.User, error) {
	if len(filter.UsersIDs) == 0 && len(filter.Phones) == 0 && len(filter.Emails) == 0 {
		return []*ent.User{}, nil
	}

	query := r.db.User.Query().Where(
		user.Or(
			user.IDIn(filter.UsersIDs...),
			user.PhoneIn(filter.Phones...),
			user.EmailIn(filter.Emails...),
		),
	)

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

func (r *usersRepo) PhoneVerified(ctx context.Context, userID int64) error {
	return r.db.User.UpdateOneID(userID).
		Where(
			user.PhoneVerified(false),
		).
		SetPhoneVerified(true).
		SetUsername(fmt.Sprintf("user%d", userID)).
		Exec(ctx)
}

func (r *usersRepo) EmailVerified(ctx context.Context, userID int64) error {
	return r.db.User.UpdateOneID(userID).
		Where(
			user.EmailVerified(false),
		).
		SetEmailVerified(true).
		SetUsername(fmt.Sprintf("user%d", userID)).
		Exec(ctx)
}

func (r *usersRepo) TempGetUsersWithoutDefaultTenant(ctx context.Context) ([]*ent.User, error) {
	return r.db.User.Query().
		Where(user.DefaultTenantIDIsNil()).
		All(ctx)
}
