package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sourcegraph/conc"
)

const (
	userStatusKeyPrefix = "user:status:"
	statusTTL           = 24 * time.Hour
)

type statusRepository struct {
	client *redis.Client
}

func NewStatusRepository(client *redis.Client) *statusRepository {
	return &statusRepository{client: client}
}

func (r *statusRepository) UpdateStatus(ctx context.Context, userID uuid.UUID, status string) error {
	key := fmt.Sprintf("%s%s", userStatusKeyPrefix, userID.String())
	return r.client.Set(ctx, key, status, statusTTL).Err()
}

func (r *statusRepository) GetStatus(ctx context.Context, userID uuid.UUID) (string, error) {
	key := fmt.Sprintf("%s%s", userStatusKeyPrefix, userID.String())
	status, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "offline", nil
	}
	if err != nil {
		return "", err
	}
	return status, nil
}

func (r *statusRepository) GetMultiStatus(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]string, error) {
	if len(userIDs) == 0 {
		return make(map[uuid.UUID]string), nil
	}

	// Use conc.WaitGroup for concurrent status retrieval
	var wg conc.WaitGroup
	results := make(map[uuid.UUID]string, len(userIDs))

	// Create a channel to collect errors
	errCh := make(chan error, len(userIDs))

	for _, userID := range userIDs {
		userID := userID // Create a new variable for the goroutine
		wg.Go(func() {
			status, err := r.GetStatus(ctx, userID)
			if err != nil {
				errCh <- fmt.Errorf("failed to get status for user %s: %w", userID, err)
				return
			}

			// Use a mutex to safely update the results map
			results[userID] = status
		})
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errCh)

	// Check if there were any errors
	var errs []error
	for err := range errCh {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("errors occurred while fetching statuses: %v", errs)
	}

	return results, nil
}
