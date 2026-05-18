package remote

import "context"

type IWebsocketsRemote interface {
	// GetUserStatus returns whether the user is online.
	GetUserStatus(ctx context.Context, userID int64) (isOnline bool, err error)
	// ListUsersStatuses returns map[userID]isOnline.
	ListUsersStatuses(ctx context.Context, usersIDs []int64) (map[int64]bool, error)
}

type websocketsRemoteStub struct{}

func NewWebsocketsRemote() IWebsocketsRemote {
	return &websocketsRemoteStub{}
}

func (r *websocketsRemoteStub) GetUserStatus(_ context.Context, _ int64) (bool, error) {
	return false, nil
}

func (r *websocketsRemoteStub) ListUsersStatuses(_ context.Context, _ []int64) (map[int64]bool, error) {
	return nil, nil
}
