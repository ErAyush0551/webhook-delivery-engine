package delivery

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"hookflow/core/queue"
	"hookflow/core/subscription"
	"hookflow/core/webhook"

	"github.com/redis/go-redis/v9"
)

type Worker struct {
	redisClient     *redis.Client 
	subscriptionSvc subscription.Repository
	webhookRepo     webhook.Repository
	deliveryRepo    Repository
	httpClient      *http.Client
}

func NewWorker(
	redisClient *redis.Client,
	subscriptionRepo subscription.Repository,
	webhookRepo webhook.Repository,
	deliveryRepo Repository,
) *Worker {
	return &Worker{
		redisClient:     redisClient,
		subscriptionSvc: subscriptionRepo,
		webhookRepo:     webhookRepo,
		deliveryRepo:    deliveryRepo,
		httpClient: &http.Client{
			Timeout: 8 * time.Second, // assignment: 5-10 sec timeout
		},
	}
}

// Start: har 2 second me queue check karta rehta hai, jitne bhi
// tasks "ready" hain unko process karta hai. Ye ek infinite loop hai,
// isiliye main.go se isko ek goroutine (go worker.Start(ctx)) me chalana hai.
func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

func (w *Worker) processBatch(ctx context.Context) {
	tasks, err := queue.Dequeue(ctx, w.redisClient, 10)
	if err != nil {
		log.Println("worker: dequeue error:", err)
		return
	}

	for _, task := range tasks {
		w.handleTask(ctx, task)
	}
}

func (w *Worker) handleTask(ctx context.Context, task queue.DeliveryTask) {
	// 1. Webhook ka payload fetch karo
	wh, err := w.webhookRepo.GetByID(ctx, task.WebhookID)
	if err != nil {
		log.Println("worker: webhook not found:", task.WebhookID, err)
		return
	}

	// 2. Subscription details fetch karo (target URL) — cache-first
	sub, err := w.getSubscriptionCached(ctx, task.SubscriptionID)
	if err != nil {
		log.Println("worker: subscription not found:", task.SubscriptionID, err)
		return
	}

	// 3. Actual HTTP POST attempt karo
	statusCode, attemptErr := w.attemptDelivery(ctx, sub.TargetURL, wh.Payload)

	// 4. Result ko log karo
	dLog := &Log{
		WebhookID:      wh.ID,
		SubscriptionID: sub.ID,
		TargetURL:      sub.TargetURL,
		AttemptNumber:  task.AttemptNumber,
	}

	success := attemptErr == nil && statusCode >= 200 && statusCode < 300

	if success {
		dLog.Status = "success"
		code := statusCode
		dLog.HTTPStatus = &code
		w.deliveryRepo.Create(ctx, dLog)
		w.webhookRepo.UpdateStatus(ctx, wh.ID, "delivered")
		log.Printf("delivered webhook %s (attempt %d)\n", wh.ID, task.AttemptNumber)
		return
	}

	// Fail hua — retry karna hai ya permanently fail mark karna hai
	if attemptErr != nil {
		errMsg := attemptErr.Error()
		dLog.Error = &errMsg
	} else {
		code := statusCode
		dLog.HTTPStatus = &code
	}

	if task.AttemptNumber >= MaxAttempts {
		dLog.Status = "failed"
		w.deliveryRepo.Create(ctx, dLog)
		w.webhookRepo.UpdateStatus(ctx, wh.ID, "failed")
		log.Printf("permanently failed webhook %s after %d attempts\n", wh.ID, task.AttemptNumber)
		return
	}

	dLog.Status = "failed_attempt"
	w.deliveryRepo.Create(ctx, dLog)

	// Retry ko queue me wapas daalo, exponential backoff delay ke saath
	delay := RetrySchedule[task.AttemptNumber-1]
	nextTask := queue.DeliveryTask{
		WebhookID:      task.WebhookID,
		SubscriptionID: task.SubscriptionID,
		AttemptNumber:  task.AttemptNumber + 1,
	}
	if err := queue.EnqueueDelayed(ctx, w.redisClient, nextTask, delay); err != nil {
		log.Println("worker: failed to schedule retry:", err)
	}
	log.Printf("scheduled retry for webhook %s in %s (next attempt %d)\n", wh.ID, delay, task.AttemptNumber+1)
}

func (w *Worker) attemptDelivery(ctx context.Context, targetURL string, payload []byte) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(payload))
	if err != nil {
		return 0, fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body) // body poori padh lo, connection reuse ke liye

	return resp.StatusCode, nil
}

func (w *Worker) getSubscriptionCached(ctx context.Context, id string) (*subscription.Subscription, error) {
	cacheKey := "subscription:" + id

	// 1. Pehle Redis cache check karo
	cached, err := w.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		var sub subscription.Subscription
		if jsonErr := json.Unmarshal([] byte(cached), &sub); jsonErr == nil {
			return &sub, nil 
		}
	}

	// 2. Cache miss — Postgres Database se data lao
	sub, err := w.subscriptionSvc.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 3. Next time ke liye Redis cache me store kar do (5 min TTL ke saath)
	if data, jsonErr := json.Marshal(sub); jsonErr == nil {
		w.redisClient.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	return sub, nil
}