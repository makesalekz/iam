package server

import (
	"context"
	"os"
	"time"

	"github.com/makesalekz/iam/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/robfig/cron/v3"
)

type CronServer struct {
	log  *log.Helper
	cron *cron.Cron
}

// NewCronServer.
func NewCronServer(
	logger log.Logger,
	usersUsecase *biz.UsersUsecase,
) *CronServer {
	location, err := time.LoadLocation("Asia/Almaty")
	if err != nil {
		location = time.Local
	}

	cs := &CronServer{
		log:  log.NewHelper(log.With(logger, "module", "server/cron")),
		cron: cron.New(cron.WithLocation(location)),
	}

	cs.deleteUserData(usersUsecase)

	return cs
}

func (cs *CronServer) deleteUserData(usersUsecase *biz.UsersUsecase) {
	frequency := "@midnight"

	if os.Getenv("DEBUG") != "" {
		frequency = "@every 3m"
	}

	entryID, err := cs.cron.AddFunc(frequency, func() {
		usersUsecase.DeleteUserData(context.Background())
	})
	if err != nil {
		cs.log.Errorf("failed on cron entryID: %v, err: %v", entryID, err)
		return
	}
}

func (cs *CronServer) Start(_ context.Context) error {
	cs.cron.Start()
	cs.log.Info("cron server started")

	return nil
}

func (cs *CronServer) Stop(_ context.Context) error {
	cs.cron.Stop()
	cs.log.Info("cron server stopped")

	return nil
}
