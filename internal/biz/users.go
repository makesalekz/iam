package biz

import (
	"context"
	_ "embed"
	"slices"

	iam_v1 "iam/api/iam/v1"
	"iam/ent"
	"iam/internal/data"
	contacts_v1 "iam/third_party/contacts/api/contacts/v1"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/registry"
)

// UsersUsecase .
type UsersUsecase struct {
	log       *log.Helper
	jwt       *data.JwtProcessor
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
		jwt:       jwt,
		usersRepo: usersRepo,
		otpRepo:   otpRepo,
		dialer:    dialer,
	}, nil
}

func (uc *UsersUsecase) GetUserProfile(ctx context.Context, filter data.GetUserFilterDto) (*ent.User, error) {
	if filter.Phone != "" && filter.Email == "" && filter.UserId == 0 {
		return uc.usersRepo.GetUserByPhone(ctx, filter.Phone)
	} else if filter.Email != "" && filter.Phone == "" && filter.UserId == 0 {
		return uc.usersRepo.GetUserByEmail(ctx, filter.Email)
	} else if filter.UserId != 0 && filter.Email == "" && filter.Phone == "" {
		return uc.usersRepo.GetUserById(ctx, filter.UserId)
	}

	return nil, iam_v1.ErrorInvalidRequest("invalid request, please read documentations")
}

func (uc *UsersUsecase) UpdateUserProfile(ctx context.Context, userId int64, data data.UpdateUserDto) (*ent.User, error) {
	return uc.usersRepo.UpdateUserData(ctx, userId, data)
}

func (uc *UsersUsecase) DeleteUser(ctx context.Context, userId int64) error {
	return uc.usersRepo.DeleteUser(ctx, userId)
}

func (uc *UsersUsecase) GetUsers(ctx context.Context, filter data.GetUsersFilterDto) ([]*ent.User, error) {
	return uc.usersRepo.GetUsers(ctx, filter)
}

func (uc *UsersUsecase) GetUserContactLabel(ctx context.Context, userId int64) (*iam_v1.Contact, error) {
	ownerId, ok := uc.jwt.GetUserIdFromContext(ctx)
	if !ok {
		return nil, iam_v1.ErrorUnauthorized("Unauthorized")
	}

	contactClient, err := uc.dialer.Contacts(ctx)
	if err != nil {
		return &iam_v1.Contact{}, iam_v1.ErrorGrpcConnection("dialer.Users: %s", err.Error())
	}

	labels, err := contactClient.GetLabelsByUserId(ctx, &contacts_v1.GetLabelsByUserIdRequest{UserId: ownerId})
	if err != nil {
		if contacts_v1.IsNotFound(err) {
			return &iam_v1.Contact{}, iam_v1.ErrorContactNotFound("there is not such contact")
		}
	}

	contact := iam_v1.Contact{}
	if len(labels.GetLabels()) == 0 {
		return &contact, iam_v1.ErrorContactNotFound("there is not such contact")
	}

	label := slices.MaxFunc(labels.GetLabels(), func(a, b string) int { return len(a) - len(b) })
	contact.Label = &label

	return &contact, nil
}
