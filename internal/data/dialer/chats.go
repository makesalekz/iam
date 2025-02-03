package dialer

import (
	"context"

	chats_v1 "gitlab.calendaria.team/services/chats/api/chats/v1"
	v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/internal/conf"
	u_dialer "gitlab.calendaria.team/services/utils/v2/dialer"
)

type IChatsRemote interface {
	DeleteUsersDataInChats(ctx context.Context, usersIDs []int64) error
}

type ChatsRemote struct {
	dialer u_dialer.IDialer
}

func NewChatsRemote(
	conf *conf.Bootstrap,
	dm u_dialer.IDialerManager,
) (IChatsRemote, func(), error) {
	dialer, err := dm.NewServiceDialer("chats", conf.GetDiscovery().GetChats())
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		dialer.Close()
	}

	return &ChatsRemote{
		dialer: dialer,
	}, cleanup, nil
}

func (r *ChatsRemote) getMembersClient(ctx context.Context) (chats_v1.MembersClient, error) {
	conn, err := r.dialer.Connect(ctx)
	if err != nil {
		return nil, v1.ErrorGrpcConnection("chats: %s", err.Error())
	}

	return chats_v1.NewMembersClient(conn), nil
}

func (r *ChatsRemote) DeleteUsersDataInChats(ctx context.Context, usersIDs []int64) error {
	client, err := r.getMembersClient(ctx)
	if err != nil {
		return err
	}

	_, err = client.DeleteUsersData(ctx, &chats_v1.DeleteUsersDataRequest{UsersIds: usersIDs})

	return err
}
