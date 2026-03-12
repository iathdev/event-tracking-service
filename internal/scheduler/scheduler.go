package scheduler

import (
	"context"
	"event-tracking-service/config"
	"event-tracking-service/internal/services"

	redislock "github.com/go-co-op/gocron-redis-lock/v2"
	"github.com/go-co-op/gocron/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Scheduler struct {
	scheduler      gocron.Scheduler
	redisClient    *redis.Client
	cfg            *config.Config
	logger         *zap.Logger
	eventProcessor *services.EventProcessor
}

func NewScheduler(
	redisClient *redis.Client,
	cfg *config.Config,
	logger *zap.Logger,
	eventProcessor *services.EventProcessor,
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
		scheduler:      s,
		redisClient:    redisClient,
		cfg:            cfg,
		logger:         logger,
		eventProcessor: eventProcessor,
	}, nil
}

func (s *Scheduler) RegisterJobs() error {
	_, err := s.scheduler.NewJob(
		gocron.DurationJob(s.cfg.Scheduler.ProcessInterval),
		gocron.NewTask(s.processEventQueue),
		gocron.WithName("process_event_queue"),
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Scheduler) processEventQueue() {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.Scheduler.ProcessTimeout)
	defer cancel()

	s.logger.Info("scheduler: starting event queue processing")
	s.eventProcessor.ProcessQueue(ctx)
	s.logger.Info("scheduler: event queue processing finished")
}

func (s *Scheduler) Start() {
	s.scheduler.Start()
	s.logger.Info("Scheduler started")
}

func (s *Scheduler) Stop() error {
	s.logger.Info("Scheduler stopping")
	return s.scheduler.Shutdown()
}
