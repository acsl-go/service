package service

import (
	"context"
	"time"
)

// deprecated
func Timer(interval time.Duration, task func()) ServiceTaskFunc {
	return func(ctx context.Context) {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				task()
			case <-ctx.Done():
				return
			}
		}
	}
}
