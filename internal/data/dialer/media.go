package dialer

import (
	"context"

	iam_v1 "gitlab.calendaria.team/services/iam/api/iam/v1"
	"gitlab.calendaria.team/services/iam/internal/conf"
	media_v1 "gitlab.calendaria.team/services/media/api/media/v1"
	u_dialer "gitlab.calendaria.team/services/utils/v2/dialer"

	"github.com/go-kratos/kratos/v2/log"
)

type IMediaRemote interface {
	DeleteAvatar(ctx context.Context, urls []string) error
}

type MediaRemote struct {
	log    *log.Helper
	dialer u_dialer.IDialer
}

func NewMediaRemote(
	logger log.Logger,
	conf *conf.Bootstrap,
	dm u_dialer.IDialerManager,
) (IMediaRemote, func(), error) {
	dialer, err := dm.NewServiceDialer("media", conf.GetDiscovery().GetMedia())
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		dialer.Close()
	}

	return &MediaRemote{
		log:    log.NewHelper(log.With(logger, "module", "data/media")),
		dialer: dialer,
	}, cleanup, nil
}

func (r *MediaRemote) getMediaClient(ctx context.Context) (media_v1.MediaServiceClient, error) {
	conn, err := r.dialer.Connect(ctx)
	if err != nil {
		return nil, iam_v1.ErrorGrpcConnection("can't connect to media: %s", err.Error())
	}

	return media_v1.NewMediaServiceClient(conn), nil
}

func (r *MediaRemote) DeleteAvatar(ctx context.Context, urls []string) error {
	client, err := r.getMediaClient(ctx)
	if err != nil {
		return err
	}

	_, err = client.DeleteAvatar(ctx, &media_v1.DeleteAvatarRequest{Urls: urls})
	if err != nil {
		return err
	}

	return nil
}
