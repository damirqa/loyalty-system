package domain

import "time"

type Withdrawal struct {
	UserID      int64
	Order       string
	Sum         float64
	ProcessedAt time.Time
}
