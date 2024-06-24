package service

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type ServiceTask func(*sync.WaitGroup, chan os.Signal)

var (
	started      bool            = false
	wg           *sync.WaitGroup = &sync.WaitGroup{}
	quit_signal  chan os.Signal  = make(chan os.Signal, 1)
	padding_list []ServiceTask   = []ServiceTask{}
)

func Run(tasks ...ServiceTask) {
	if started {
		for _, task := range tasks {
			wg.Add(1)
			go task(wg, quit_signal)
		}
	} else {
		padding_list = append(padding_list, tasks...)
	}
}

func Start() {
	started = true
	signal.Notify(quit_signal, syscall.SIGTERM, syscall.SIGINT)

	Run(padding_list...)

	fmt.Println("System Started")

	<-quit_signal
	fmt.Println("System Stopping ...")
	quit_signal <- syscall.SIGTERM

	wg.Wait()
}
