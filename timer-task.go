package service

import (
	"os"
	"sync"
	"time"
)

func Timer(interval time.Duration, task func()) ServiceTask {
	return func(wg *sync.WaitGroup, quit_signal chan os.Signal) {
		defer wg.Done()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				task()
			case <-quit_signal:
				return
			}
		}
	}
}
