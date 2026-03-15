package service

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type ServiceTask func(context.Context)

var (
	_started      bool            = false
	_wg           *sync.WaitGroup = &sync.WaitGroup{}
	_ctx          context.Context = nil
	_padding_list []ServiceTask   = []ServiceTask{}
)

func taskWrapper(task ServiceTask) {
	defer _wg.Done()
	task(_ctx)
}

func Run(tasks ...ServiceTask) {
	if _started {
		for _, task := range tasks {
			_wg.Add(1)
			go taskWrapper(task)
		}
	} else {
		_padding_list = append(_padding_list, tasks...)
	}
}

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
