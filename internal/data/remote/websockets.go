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
	GetUserPresence(ctx context.Context, userID int64) (*websockets_v1.UserPresenceModel, error)
	ListUsersPresences(ctx context.Context, usersIDs []int64) (map[int64]*websockets_v1.UserPresenceModel, error)
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

func (r *WebsocketsRemote) getUserPresenceClient(ctx context.Context) (websockets_v1.UserPresenceClient, error) {
	conn, err := r.dialer.Connect(ctx)
	if err != nil {
		return nil, iam_v1.ErrorGrpcConnection("can't connect to websockets: %s", err.Error())
	}

	return websockets_v1.NewUserPresenceClient(conn), nil
}

func (r *WebsocketsRemote) GetUserPresence(
	ctx context.Context,
	userID int64,
) (*websockets_v1.UserPresenceModel, error) {
	client, err := r.getUserPresenceClient(ctx)
	if err != nil {
		return nil, err
	}

	presence, err := client.GetUserPresence(ctx, &websockets_v1.GetUserPresenceRequest{UserId: userID})
	if err != nil {
		return nil, err
	}

	return presence.GetUserPresence(), nil
}

func (r *WebsocketsRemote) ListUsersPresences(
	ctx context.Context,
	usersIDs []int64,
) (map[int64]*websockets_v1.UserPresenceModel, error) {
	client, err := r.getUserPresenceClient(ctx)
	if err != nil {
		return nil, err
	}

	presence, err := client.ListUsersPresences(ctx, &websockets_v1.ListUsersPresencesRequest{UsersIds: usersIDs})
	if err != nil {
		return nil, err
	}

	mapPresences := make(map[int64]*websockets_v1.UserPresenceModel)
	for _, userPresence := range presence.GetUsersPresences() {
		mapPresences[userPresence.GetUserId()] = userPresence
	}

	return mapPresences, nil
}
