package background

import (
	"context"
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
	manager          *asynq.PeriodicTaskManager
}

func NewLeaderElection(redisClient *redis.UniversalClient, manager *asynq.PeriodicTaskManager) *LeaderElection {
	return &LeaderElection{
		redisClient:      redisClient,
		nodeID:           uuid.New().String(),
		lockTTL:          10 * time.Second,
		renewalInterval:  3 * time.Second,
		electionInterval: 5 * time.Second,
		manager:          manager,
	}
}

func (le *LeaderElection) RunWithElection(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Context cancelled, stopping leader election")
			return
		default:
			le.tryBecomeLeader(ctx)
			time.Sleep(le.electionInterval)
		}
	}
}

func (le *LeaderElection) tryBecomeLeader(ctx context.Context) {
	// Try to acquire leadership
	acquired, err := (*le.redisClient).SetNX(ctx, leaderLockKey, le.nodeID, le.lockTTL).Result()
	if err != nil {
		log.Error().Err(err).Msg("Failed to acquire leadership lock")
		return
	}

	if acquired {
		log.Info().Str("nodeID", le.nodeID).Msg("Acquired leadership, starting scheduler")
		le.runAsLeader(ctx)
	} else {
		log.Info().Str("nodeID", le.nodeID).Msg("Not acquired leadership, skip starting scheduler")
	}
}

func (le *LeaderElection) runAsLeader(ctx context.Context) {
	// Start the periodic task manager
	if err := le.manager.Start(); err != nil {
		log.Error().Err(err).Msg("Failed to start periodic task manager")
		return
	}
	defer func() {
		le.manager.Shutdown()
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
