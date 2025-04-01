package biz

import (
	"context"
	"errors"
	"os"
	"sync"
	"time"

	v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/ent"
	"gitlab.calendaria.team/services/iam/internal/data"
	"gitlab.calendaria.team/services/iam/internal/data/dialer"
	"gitlab.calendaria.team/services/iam/internal/data/dto"
	tenants_v1 "gitlab.calendaria.team/services/tenants/api/tenants/v1"
	utils_v1 "gitlab.calendaria.team/services/utils/api/utils/v1"
	u_error "gitlab.calendaria.team/services/utils/v1/error"
	u_jwt "gitlab.calendaria.team/services/utils/v4/jwt"
	u_nats "gitlab.calendaria.team/services/utils/v4/nats"

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
	log           *log.Helper
	jwt           u_jwt.IJwtProcessor
	queue         u_nats.IQueueManager
	tenants       dialer.ITenantsRemote
	contacts      dialer.IContactsRemote
	chats         dialer.IChatsRemote
	events        dialer.IEventsRemote
	media         dialer.IMediaRemote
	usersRepo     data.UsersRepo
	otpRepo       data.OtpRepo
	privaciesRepo data.PrivacyRepo
}

type ConstraintKey string

const (
	USERNAME ConstraintKey = "users_username_key"
	EMAIL    ConstraintKey = "users_email_key"
	PHONE    ConstraintKey = "users_phone_key"

	DeleteDuration        = time.Duration(30*24) * time.Hour
	DeleteDurationInDebug = time.Duration(5) * time.Minute

	phoneMask  = 0b001
	emailMask  = 0b010
	userIDMask = 0b100
)

// NewUsersUsecase .
func NewUsersUsecase(
	logger log.Logger,
	jwt u_jwt.IJwtProcessor,
	queue u_nats.IQueueManager,
	tenants dialer.ITenantsRemote,
	contacts dialer.IContactsRemote,
	chats dialer.IChatsRemote,
	events dialer.IEventsRemote,
	media dialer.IMediaRemote,
	usersRepo data.UsersRepo,
	otpRepo data.OtpRepo,
	privaciesRepo data.PrivacyRepo,
) (*UsersUsecase, error) {
	return &UsersUsecase{
		log:           log.NewHelper(log.With(logger, "module", "usecase/users")),
		jwt:           jwt,
		queue:         queue,
		tenants:       tenants,
		contacts:      contacts,
		chats:         chats,
		events:        events,
		media:         media,
		usersRepo:     usersRepo,
		otpRepo:       otpRepo,
		privaciesRepo: privaciesRepo,
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

func (uc *UsersUsecase) DeleteUserData(ctx context.Context) {
	users, err := uc.usersRepo.GetUsersToDelete(ctx)
	if err != nil {
		uc.log.Errorf("failed to get users to delete: %v", err)
		return
	}

	if len(users) == 0 {
		return
	}

	usersIDs := make([]int64, len(users))
	avatars := make([]string, 0, len(users))
	for i, user := range users {
		usersIDs[i] = user.ID
		if user.Avatar != nil {
			avatars = append(avatars, *user.Avatar)
		}
	}

	wg := sync.WaitGroup{}

	// Delete tenants, groups, members of users in tenants service
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = uc.tenants.DeleteUsersTenants(ctx, usersIDs)
		if err != nil {
			uc.log.Errorf("failed to delete data in tenants: %v", err)
		}
	}()

	// Delete contacts, relations of users in contacts service
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = uc.contacts.DeleteUsersDataInContacts(ctx, usersIDs)
		if err != nil {
			uc.log.Errorf("failed to delete data in contacts: %v", err)
		}
	}()

	// Delete members of users in chats service
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = uc.chats.DeleteUsersDataInChats(ctx, usersIDs)
		if err != nil {
			uc.log.Errorf("failed to delete data in chats: %v", err)
		}
	}()

	// Delete members of users in events service
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = uc.events.DeleteUsersDataInEvents(ctx, usersIDs)
		if err != nil {
			uc.log.Errorf("failed to delete data in chats: %v", err)
		}
	}()

	// Delete avatars of users in media service including s3 (aws)
	wg.Add(1)
	go func() {
		defer wg.Done()
		err = uc.media.DeleteAvatar(ctx, avatars)
		if err != nil {
			uc.log.Errorf("failed to delete avatars: %v", err)
		}
	}()

	wg.Wait()

	err = uc.usersRepo.DeleteUsers(ctx, usersIDs)
	if err != nil {
		uc.log.Errorf("failed to delete users: %v", err)
	} else {
		uc.log.Infof("users deleted from services: %v", usersIDs)
	}
}

func (uc *UsersUsecase) GetUserProfile(ctx context.Context, filter data.GetUserFilterDto) (*UserItem, error) {
	var user *ent.User
	var err error

	switch btoi(filter.Phone != "") |
		btoi(filter.Email != "")<<1 |
		btoi(filter.UserID != 0)<<2 {
	case phoneMask:
		user, err = uc.usersRepo.GetUserByPhone(ctx, filter.Phone, false)
	case emailMask:
		user, err = uc.usersRepo.GetUserByEmail(ctx, filter.Email, false)
	case userIDMask:
		user, err = uc.usersRepo.GetUserByID(ctx, filter.UserID, false)
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

func (uc *UsersUsecase) UpdateUserProfile(
	ctx context.Context,
	userID int64,
	dto dto.UpdateUserDto,
) (*UserItem, error) {
	var err error

	user, err := uc.usersRepo.GetUserByID(ctx, userID, false)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, v1.ErrorUserNotFound("user not found")
		}
		return nil, v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}

	if dto.Avatar == "" && user.Avatar != nil {
		err = uc.media.DeleteAvatar(ctx, []string{*user.Avatar})
		if err != nil {
			uc.log.Errorf("failed to delete avatar: %v", err)
			dto.Avatar = *user.Avatar
		}
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

func (uc *UsersUsecase) ScheduleUserDeletion(ctx context.Context, actorID int64) error {
	deleteDuration := DeleteDuration
	if os.Getenv("DEBUG") != "" {
		deleteDuration = DeleteDurationInDebug
	}

	err := uc.usersRepo.ScheduleUserDeletion(ctx, actorID, deleteDuration)
	if err != nil {
		return v1.ErrorDatabaseQuery("database error: %s", err.Error())
	}
	return nil
}

func (uc *UsersUsecase) ListUsers(
	ctx context.Context,
	filter data.GetUsersFilterDto,
	sort *utils_v1.SortRequest,
	paginate *utils_v1.PaginateRequest,
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

func (uc *UsersUsecase) GetUsers(ctx context.Context, filter data.GetUsersFilterDto) ([]*UserItem, error) {
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
