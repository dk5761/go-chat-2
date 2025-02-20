package redis

import (
	"context"
	"fmt"
	"sync"
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

	// Use a short timeout context for Redis operations
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := r.client.Set(timeoutCtx, key, status, statusTTL).Err()
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}
	return nil
}

func (r *statusRepository) GetStatus(ctx context.Context, userID uuid.UUID) (string, error) {
	key := fmt.Sprintf("%s%s", userStatusKeyPrefix, userID.String())

	// Use a short timeout context for Redis operations
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	status, err := r.client.Get(timeoutCtx, key).Result()
	if err == redis.Nil {
		return "offline", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get status: %w", err)
	}
	return status, nil
}

func (r *statusRepository) GetMultiStatus(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]string, error) {
	if len(userIDs) == 0 {
		return make(map[uuid.UUID]string), nil
	}

	// Use a short timeout context for Redis operations
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use conc.WaitGroup for concurrent status retrieval
	var wg conc.WaitGroup
	results := make(map[uuid.UUID]string, len(userIDs))
	mu := &sync.Mutex{} // Add mutex for thread-safe map updates

	// Create a channel to collect errors
	errCh := make(chan error, len(userIDs))

	for _, userID := range userIDs {
		userID := userID // Create a new variable for the goroutine
		wg.Go(func() {
			status, err := r.GetStatus(timeoutCtx, userID)
			if err != nil {
				errCh <- fmt.Errorf("failed to get status for user %s: %w", userID, err)
				return
			}

			// Use mutex to safely update the results map
			mu.Lock()
			results[userID] = status
			mu.Unlock()
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
