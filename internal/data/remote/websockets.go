package remote

import (
	"context"

	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/internal/conf"
	u_dialer "gitlab.calendaria.team/services/utils/v4/dialer"
	websockets_v1 "gitlab.calendaria.team/services/websockets/api/websockets/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type IWebsocketsRemote interface {
	GetUserStatus(ctx context.Context, userID int64) (*websockets_v1.UserStatusReply, error)
	ListUsersStatuses(ctx context.Context, usersIDs []int64) (map[int64]*websockets_v1.UserStatusModel, error)
}

type WebsocketsRemote struct {
	log    *log.Helper
	dialer u_dialer.IDialer
}

func NewWebsocketsRemote(
	logger log.Logger,
	conf *conf.Bootstrap,
	dm u_dialer.IDialerManager,
) (IWebsocketsRemote, func(), error) {
	dialer, err := dm.NewServiceDialer("websockets", conf.GetDiscovery().GetWebsockets())
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		dialer.Close()
	}

	return &WebsocketsRemote{
		log:    log.NewHelper(log.With(logger, "module", "data/websockets")),
		dialer: dialer,
	}, cleanup, nil
}

func (r *WebsocketsRemote) getUserStatusClient(ctx context.Context) (websockets_v1.UserStatusClient, error) {
	conn, err := r.dialer.Connect(ctx)
	if err != nil {
		return nil, iam_v1.ErrorGrpcConnection("can't connect to websockets: %s", err.Error())
	}

	return websockets_v1.NewUserStatusClient(conn), nil
}

func (r *WebsocketsRemote) GetUserStatus(
	ctx context.Context,
	userID int64,
) (*websockets_v1.UserStatusReply, error) {
	client, err := r.getUserStatusClient(ctx)
	if err != nil {
		return nil, err
	}

	userStatus, err := client.GetUserStatus(ctx, &websockets_v1.GetUserStatusRequest{UserId: userID})
	if err != nil {
		return nil, err
	}

	return userStatus, nil
}

func (r *WebsocketsRemote) ListUsersStatuses(
	ctx context.Context,
	usersIDs []int64,
) (map[int64]*websockets_v1.UserStatusModel, error) {
	client, err := r.getUserStatusClient(ctx)
	if err != nil {
		return nil, err
	}

	reply, err := client.ListUsersStatuses(ctx, &websockets_v1.ListUsersStatusesRequest{UsersIds: usersIDs})
	if err != nil {
		return nil, err
	}

	mapStatuses := make(map[int64]*websockets_v1.UserStatusModel)
	for _, userStatus := range reply.GetUsersStatuses() {
		mapStatuses[userStatus.GetUserId()] = userStatus
	}

	return mapStatuses, nil
}
