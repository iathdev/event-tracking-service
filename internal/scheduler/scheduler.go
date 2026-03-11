package scheduler

import (
	"event-tracking-service/config"

	"github.com/go-co-op/gocron-redis-lock/v2"
	"github.com/go-co-op/gocron/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Scheduler struct {
	scheduler   gocron.Scheduler
	redisClient *redis.Client
	cfg         *config.Config
	logger      *zap.Logger
}

func NewScheduler(
	redisClient *redis.Client,
	cfg *config.Config,
	logger *zap.Logger,
) (*Scheduler, error) {
	locker, err := redislock.NewRedisLocker(redisClient, redislock.WithTries(1))
	if err != nil {
		return nil, err
	}

	s, err := gocron.NewScheduler(
		gocron.WithDistributedLocker(locker),
	)
	if err != nil {
		return nil, err
	}

	return &Scheduler{
		scheduler:   s,
		redisClient: redisClient,
		cfg:         cfg,
		logger:      logger,
	}, nil
}

func (s *Scheduler) RegisterJobs() error {
	// TODO: Register your jobs here
	// Example:
	// _, err := s.scheduler.NewJob(
	// 	gocron.DurationJob(s.cfg.Scheduler.ProcessInterval),
	// 	gocron.NewTask(s.yourJobFunction),
	// 	gocron.WithName("your_job_name"),
	// )
	// if err != nil {
	// 	return err
	// }

	return nil
}

func (s *Scheduler) Start() {
	s.scheduler.Start()
	s.logger.Info("Scheduler started")
}

func (s *Scheduler) Stop() error {
	s.logger.Info("Scheduler stopping")
	return s.scheduler.Shutdown()
}
