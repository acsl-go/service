package service

import (
	"context"
	"time"
)

func Timer(interval time.Duration, task func()) ServiceTask {
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
