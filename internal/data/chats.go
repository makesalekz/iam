package data

import (
	"context"

	chats_v1 "gitlab.calendaria.team/services/chats/api/chats/v1"
	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/internal/conf"
	"gitlab.calendaria.team/services/utils/v1/dialer"
)

type ChatsRemote struct {
	dialer *dialer.Dialer
	conf   *conf.Bootstrap
}

func NewChatsRemote(d *dialer.Dialer, conf *conf.Bootstrap) (*ChatsRemote, error) {
	return &ChatsRemote{
		dialer: d,
		conf:   conf,
	}, nil
}

func (r *ChatsRemote) GetChatsClient(ctx context.Context) (chats_v1.ChatsClient, error) {
	return dialer.NewDialerBuilder(r.dialer, chats_v1.NewChatsClient).
		SetEndpoint(r.conf.Discovery.Chats).
		SetTimeout(r.conf.Discovery.ChatsTimeout.AsDuration()).
		Conn(ctx, nil)
}

func (r *ChatsRemote) GetMembersClient(ctx context.Context) (chats_v1.MembersClient, error) {
	return dialer.NewDialerBuilder(r.dialer, chats_v1.NewMembersClient).
		SetEndpoint(r.conf.Discovery.Chats).
		SetTimeout(r.conf.Discovery.ChatsTimeout.AsDuration()).
		Conn(ctx, nil)
}

func (r *ChatsRemote) GetDirectChatMembership(ctx context.Context, req *chats_v1.DirectChatMembershipRequest) (*chats_v1.MembershipReply, error) {
	membersClient, err := r.GetMembersClient(ctx)
	if err != nil {
		return nil, iam_v1.ErrorGrpcConnection("chats: %s", err.Error())
	}

	replyMembers, err := membersClient.GetDirectChatMembership(ctx, req)
	if err != nil {
		return nil, err
	}

	return replyMembers, nil
}
