package remote

import "context"

type IChatsRemote interface {
	DeleteUsersDataInChats(ctx context.Context, usersIDs []int64) error
}

type chatsRemoteStub struct{}

func NewChatsRemote() IChatsRemote {
	return &chatsRemoteStub{}
}

func (r *chatsRemoteStub) DeleteUsersDataInChats(_ context.Context, _ []int64) error {
	return nil
}
