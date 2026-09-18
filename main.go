package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"hookflow/core/cache"
	"hookflow/core/config"
	"hookflow/core/database"
	"hookflow/core/delivery"
	"hookflow/core/subscription"
	"hookflow/core/webhook"
)

func main() {

	cfg := config.Load()

	fmt.Println("Server configured on port:", cfg.AppPort)

	// 1. Connect to PostgreSQL

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("Connected to PostgreSQL")

	// 2. Connect to Redis

	if err := cache.Connect(
		cfg.RedisAddr,
		cfg.RedisPassword,
		cfg.RedisDB,
	); err != nil {
		log.Fatal(err)
	}

	// ============================================
	// 3. Run database migrations
	// ============================================
	if err := database.RunMigrations(
		context.Background(),
		db,
	); err != nil {
		log.Fatal(err)
	}

	// 4. Subscription setup

	// Repository → Service → Handler
	subscriptionRepository := subscription.NewPostgresRepository(db)
	subscriptionService := subscription.NewService(subscriptionRepository)
	subscriptionHandler := subscription.NewHandler(subscriptionService)

	// 5. Webhook setup

	// Repository → Service → Handler
	// This handles incoming webhooks.
	webhookRepository := webhook.NewPostgresRepository(db)
	webhookService := webhook.NewService(webhookRepository, cache.Client)
	webhookHandler := webhook.NewHandler(webhookService)

	// 6. Delivery setup (worker background me chalega)

	deliveryRepository := delivery.NewPostgresRepository(db)
	deliveryWorker := delivery.NewWorker(
		cache.Client,
		subscriptionRepository,
		webhookRepository,
		deliveryRepository,
	)
	go deliveryWorker.Start(context.Background())
	fmt.Println("Delivery worker started")

	// 7. Create HTTP router

	mux := http.NewServeMux()

	// 8. Subscription routes
	// POST   /subscriptions  → Create subscription
	// GET    /subscriptions  → Get all subscriptions
	mux.HandleFunc("/subscriptions", func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {

		case http.MethodPost:
			subscriptionHandler.Create(w, r)

		case http.MethodGet:
			subscriptionHandler.GetAll(w, r)

		default:
			http.Error(
				w,
				"method not allowed",
				http.StatusMethodNotAllowed,
			)
		}
	})

	// GET    /subscriptions/{id} → Get subscription
	// PUT    /subscriptions/{id} → Update subscription
	// DELETE /subscriptions/{id} → Delete subscription
	mux.HandleFunc("/subscriptions/", func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {

		case http.MethodGet:
			subscriptionHandler.GetByID(w, r)

		case http.MethodPut:
			subscriptionHandler.Update(w, r)

		case http.MethodDelete:
			subscriptionHandler.Delete(w, r)

		default:
			http.Error(
				w,
				"method not allowed",
				http.StatusMethodNotAllowed,
			)
		}
	})

	// ============================================
	// 9. Webhook ingestion route
	// ============================================
	//
	// POST /ingest/{subscription_id}
	//
	// Example:
	// POST http://localhost:8080/ingest/6b7cf656-ba6f-4132-a84c-c794753c42af
	//
	// This sends the incoming webhook to webhookHandler.
	//
	mux.HandleFunc("/ingest/", webhookHandler.Ingest)

	// ============================================
	// 10. Start HTTP server
	// ============================================
	fmt.Println("HookFlow server running on port:", cfg.AppPort)

	if err := http.ListenAndServe(":"+cfg.AppPort, mux); err != nil {
		log.Fatal(err)
	}
}