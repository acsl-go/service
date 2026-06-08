package service

import (
	"context"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
)

type ITask interface {
	Main(ctx context.Context)
}

type TaskManager struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

type TaskWrapper struct {
	body func(ctx context.Context)
}

func (tw *TaskWrapper) Main(ctx context.Context) {
	tw.body(ctx)
}

func NewTask(body func(ctx context.Context)) *TaskWrapper {
	return &TaskWrapper{body: body}
}

func NewTaskManager() *TaskManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &TaskManager{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (tm *TaskManager) Context() context.Context {
	return tm.ctx
}

func (tm *TaskManager) Run(task ITask) ITask {
	if task == nil {
		log.Printf("【Warning】Attempting to run nil task, skipping.\nStack:\n%s", debug.Stack())
		return nil
	}
	tm.wg.Add(1)
	go func() {
		defer tm.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				log.Printf("【Task Panic】Recover: %v\nStack:\n%s", r, debug.Stack())
			}
		}()
		task.Main(tm.ctx)
	}()
	return task
}

func (tm *TaskManager) RunMultiple(tasks ...ITask) {
	for _, task := range tasks {
		tm.Run(task)
	}
}

func (tm *TaskManager) Stop() {
	tm.cancel()
}

func (tm *TaskManager) Wait() {
	if tm.ctx.Err() == nil {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

		select {
		case <-sigChan:
			tm.Stop()
		case <-tm.ctx.Done():
		}

		signal.Stop(sigChan)
	}
	tm.wg.Wait()
}
