package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// DeliveryQueueKey is the Redis Sorted Set used as our delivery queue.
const DeliveryQueueKey = "delivery_tasks"

// DeliveryTask represents one webhook delivery job.
type DeliveryTask struct {
	WebhookID      string `json:"webhook_id"`
	SubscriptionID string `json:"subscription_id"`
	AttemptNumber  int    `json:"attempt_number"`
}

// Enqueue adds a task to the queue immediately.
func Enqueue(
	ctx context.Context,
	client *redis.Client,
	task DeliveryTask,
) error {
	return EnqueueAt(ctx, client, task, time.Now())
}

// EnqueueDelayed adds a task to the queue for future processing.
// We will use this later for webhook retries.
func EnqueueDelayed(
	ctx context.Context,
	client *redis.Client,
	task DeliveryTask,
	delay time.Duration,
) error {
	return EnqueueAt(ctx, client, task, time.Now().Add(delay))
}

// EnqueueAt adds a task to Redis with a specific execution time.
//
// Redis Sorted Set:
//   score  = time when task should run
//   member = actual delivery task
func EnqueueAt(
	ctx context.Context,
	client *redis.Client,
	task DeliveryTask,
	runAt time.Time,
) error {

	// Convert DeliveryTask into JSON.
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	// The Redis score represents when this task should be processed.
	score := float64(runAt.UnixMilli())

	// Add the task to the Redis Sorted Set.
	err = client.ZAdd(
		ctx,
		DeliveryQueueKey,
		redis.Z{
			Score:  score,
			Member: string(data),
		},
	).Err()

	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}

	return nil
}

// Dequeue returns tasks whose scheduled time has arrived.
//
// Only tasks with:
// score <= current time
//
// are returned.
func Dequeue(
	ctx context.Context,
	client *redis.Client,
	limit int64,
) ([]DeliveryTask, error) {

	// Current time in milliseconds.
	now := fmt.Sprintf("%d", time.Now().UnixMilli())

	// Get tasks whose score is between 0 and now.
	results, err := client.ZRangeByScore(
		ctx,
		DeliveryQueueKey,
		&redis.ZRangeBy{
			Min:   "0",
			Max:   now,
			Count: limit,
		},
	).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to fetch tasks: %w", err)
	}

	var tasks []DeliveryTask

	for _, raw := range results {

		// Convert JSON back into DeliveryTask.
		var task DeliveryTask

		if err := json.Unmarshal([]byte(raw), &task); err != nil {
			// If Redis contains corrupted data,
			// skip that particular task.
			continue
		}

		tasks = append(tasks, task)

		// Remove the task from the queue after fetching it.
		//
		// This prevents the same task from being picked again
		// by the next Dequeue call.
		if err := client.ZRem(
			ctx,
			DeliveryQueueKey,
			raw,
		).Err(); err != nil {
			return nil, fmt.Errorf(
				"failed to remove task from queue: %w",
				err,
			)
		}
	}

	return tasks, nil
}