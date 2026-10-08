package delivery

import "time"

type Log struct {
	ID             string
	WebhookID      string
	SubscriptionID string
	TargetURL      string
	AttemptNumber  int
	Status         string // "success", "failed_attempt", "failed"
	HTTPStatus     *int
	Error          *string
	CreatedAt      time.Time
}

// RetrySchedule — attempt number ke hisaab se kitni der baad retry karna hai
// Attempt 1 fail ho to 10s baad Attempt 2, phir 30s baad Attempt 3
var RetrySchedule = []time.Duration{
	10 * time.Second,
	30 * time.Second,
	1 * time.Minute,
	5 * time.Minute,
	15 * time.Minute,
}

const MaxAttempts = 5

//we make this file to define the structure of the log and the retry schedule.
//with the help of this file we can keep track of the logs and retry the failed attempts.
