package service

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// deprecated, use TaskManager instead
type ServiceTaskFunc func(context.Context)

var (
	_started      bool              = false
	_wg           *sync.WaitGroup   = &sync.WaitGroup{}
	_ctx          context.Context   = nil
	_padding_list []ServiceTaskFunc = []ServiceTaskFunc{}
)

// deprecated, use TaskManager instead
func taskWrapper(task ServiceTaskFunc) {
	defer _wg.Done()
	task(_ctx)
}

// deprecated, use TaskManager instead
func Run(tasks ...ServiceTaskFunc) {
	if _started {
		for _, task := range tasks {
			_wg.Add(1)
			go taskWrapper(task)
		}
	} else {
		_padding_list = append(_padding_list, tasks...)
	}
}

// deprecated, use TaskManager instead
func Start() {
	_started = true
	ctx, cancelFunc := context.WithCancel(context.Background())
	_ctx = ctx
	Run(_padding_list...)

	fmt.Println("System Started")

	quit_signal := make(chan os.Signal, 1)
	signal.Notify(quit_signal, syscall.SIGTERM, syscall.SIGINT)
	<-quit_signal
	fmt.Println("System Stopping ...")

	cancelFunc()
	_wg.Wait()
}
