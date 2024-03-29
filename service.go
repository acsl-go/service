package service

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type ServiceTask func(*sync.WaitGroup, chan os.Signal)

func Run(tasks ...ServiceTask) {

	quit_signal := make(chan os.Signal, 1)
	signal.Notify(quit_signal, syscall.SIGTERM, syscall.SIGINT)

	wg := &sync.WaitGroup{}

	for _, task := range tasks {
		wg.Add(1)
		go task(wg, quit_signal)
	}

	fmt.Println("System Started")

	<-quit_signal
	fmt.Println("System Stopping ...")
	quit_signal <- syscall.SIGTERM

	wg.Wait()
}
