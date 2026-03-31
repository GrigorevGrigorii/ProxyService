package background

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	leaderLockKey = "background-scheduler-leader-lock-key"
)

type LeaderElection struct {
	redisClient      *redis.UniversalClient
	nodeID           string
	lockTTL          time.Duration
	renewalInterval  time.Duration
	electionInterval time.Duration
	managerOpts      asynq.PeriodicTaskManagerOpts
	isLeader         atomic.Bool
}

func NewLeaderElection(redisClient *redis.UniversalClient, opts asynq.PeriodicTaskManagerOpts) *LeaderElection {
	return &LeaderElection{
		redisClient:      redisClient,
		nodeID:           uuid.New().String(),
		lockTTL:          10 * time.Second,
		renewalInterval:  3 * time.Second,
		electionInterval: 5 * time.Second,
		managerOpts:      opts,
	}
}

func (le *LeaderElection) RunWithElection(ctx context.Context) {
	le.tryBecomeLeader(ctx)

	electionTicker := time.NewTicker(le.electionInterval)
	defer electionTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Context cancelled, stopping leader election")
			return
		case <-electionTicker.C:
			le.tryBecomeLeader(ctx)
		}
	}
}

func (le *LeaderElection) tryBecomeLeader(ctx context.Context) {
	if le.isLeader.Load() {
		return
	}

	// Try to acquire leadership
	acquired, err := (*le.redisClient).SetNX(ctx, leaderLockKey, le.nodeID, le.lockTTL).Result()
	if err != nil {
		log.Error().Err(err).Msg("Failed to acquire leadership lock")
		return
	}

	if acquired {
		log.Info().Str("nodeID", le.nodeID).Msg("Acquired leadership, starting scheduler")

		manager, err := asynq.NewPeriodicTaskManager(le.managerOpts)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create periodic task manager")
			// Release the lock since we can't start the manager
			(*le.redisClient).Del(ctx, leaderLockKey)
			return
		}
		le.runAsLeader(ctx, manager)
	}
}

func (le *LeaderElection) runAsLeader(ctx context.Context, manager *asynq.PeriodicTaskManager) {
	le.isLeader.Store(true)
	defer le.isLeader.Store(false)

	// Start the manager
	if err := manager.Start(); err != nil {
		log.Error().Err(err).Msg("Failed to start periodic task manager")
		return
	}
	defer func() {
		manager.Shutdown()
		log.Info().Msg("Stopped periodic task manager")
	}()

	log.Info().Msg("Scheduler started successfully, beginning leadership maintenance")

	// Renew leadership periodically
	renewTicker := time.NewTicker(le.renewalInterval)
	defer renewTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Context cancelled, relinquishing leadership")
			return

		case <-renewTicker.C:
			// Check that we STILL hold the lock (by verifying the value)
			currentLeader, err := (*le.redisClient).Get(ctx, leaderLockKey).Result()
			if err != nil {
				if err == redis.Nil {
					log.Warn().Msg("Leadership lock expired")
				} else {
					log.Error().Err(err).Msg("Failed to check leadership status")
				}
				return // Lost leadership, exit
			}

			if currentLeader != le.nodeID {
				log.Warn().Str("currentLeader", currentLeader).Msg("Another node holds the leadership lock")
				return // Lost leadership to another instance
			}

			// Still the leader - renew the lock
			renewed, err := (*le.redisClient).Expire(ctx, leaderLockKey, le.lockTTL).Result()
			if err != nil || !renewed {
				log.Warn().Err(err).Msg("Failed to renew leadership lock")
				return // Lost leadership
			}

			log.Debug().Msg("Leadership lock renewed successfully")
		}
	}
}
