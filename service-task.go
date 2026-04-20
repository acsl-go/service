package service

import (
	"context"
	"errors"
	"log"
	"runtime/debug"
	"sync"
	"time"
)

type IAdapter interface {
	Close() error
}

type IService interface {
	Listen(ctx context.Context) (IAdapter, error)
	Serve(ctx context.Context, conn IAdapter)
	OnListenFailed(error) time.Duration
	OnStopping()
	OnStopped()
}

type DefaultService struct{}

func (*DefaultService) OnListenFailed(err error) time.Duration { return 5 * time.Second }
func (*DefaultService) OnStopping()                            {}
func (*DefaultService) OnStopped()                             {}

type emptyService struct {
	DefaultService
}

func (*emptyService) Listen(ctx context.Context) (IAdapter, error) {
	return nil, errors.New("service not implemented")
}

func (*emptyService) Serve(ctx context.Context, conn IAdapter) {
	log.Println("service not implemented")
}

func (*emptyService) OnListenFailed(err error) time.Duration { return -1 } // no retry

var _ IService = (*emptyService)(nil)

type ServiceTask struct {
	cancel   context.CancelFunc
	lock     sync.Mutex
	delegate IService
}

var (
	ErrPanic      = errors.New("service: panic occurred")
	ErrNilAdapter = errors.New("service: listen returned nil adapter without error")
	ErrRestart    = errors.New("service: restart requested")
	ErrStop       = errors.New("service: stop requested")
)

func NewServiceTask(delegate IService) *ServiceTask {
	if delegate == nil {
		log.Printf("【Warning】Creating ServiceTask with nil delegate, using emptyService instead.\nStack:\n%s", debug.Stack())
		delegate = &emptyService{}
	}
	task := &ServiceTask{delegate: delegate}
	return task
}

func (t *ServiceTask) Run(mgr *TaskManager) *ServiceTask {
	return mgr.Run(t).(*ServiceTask)
}

func (t *ServiceTask) Stop() {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.cancel != nil {
		t.cancel()
	}
}

func (t *ServiceTask) Main(parentCtx context.Context) {
	t.lock.Lock()
	if t.cancel != nil { // ensure only one loop is running
		return
	}
	serviceCtx, serviceCancel := context.WithCancel(parentCtx)
	t.cancel = serviceCancel
	t.lock.Unlock()

	defer func() {
		t.lock.Lock()
		defer t.lock.Unlock()
		t.cancel()
		t.cancel = nil
	}()

	defer t.cancel()

	retryTimer := time.NewTimer(0)
	defer retryTimer.Stop()
	<-retryTimer.C

	var delay time.Duration = 0
	var adapter IAdapter
	var err error

	// listen loop with retry
	for adapter == nil {
		if serviceCtx.Err() != nil {
			return
		}

		adapter, err = t.listen(serviceCtx)
		if err == nil && adapter == nil {
			err = ErrNilAdapter
		}
		if err != nil {
			// handle error
			delay = t.onListenFailed(err)
			if delay > 0 && delay <= time.Millisecond {
				delay = time.Millisecond
			}

			if delay < 0 {
				return
			}
			// wait for retry
			retryTimer.Reset(delay)
			select {
			case <-serviceCtx.Done():
				return
			case <-retryTimer.C:
			}
		}
	}

	// serve
	defer adapter.Close()

	wg := sync.WaitGroup{}

	if serviceCtx.Err() == nil {
		wg.Add(1)
		go t.serve(serviceCtx, adapter, &wg)
		<-serviceCtx.Done()
	}

	// stop service
	t.onStopping()
	_ = adapter.Close()

	// wait for service to stop
	wg.Wait()

	t.onStopped()
}

func (t *ServiceTask) listen(ctx context.Context) (adapter IAdapter, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("【Listen Panic】Recover: %v\nStack:\n%s", r, debug.Stack())
			adapter = nil
			err = ErrPanic
		}
	}()
	return t.delegate.Listen(ctx)
}

func (t *ServiceTask) onListenFailed(err error) (ret time.Duration) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("【OnListenFailed Panic】Recover: %v\nStack:\n%s", r, debug.Stack())
			ret = time.Second
		}
	}()
	return t.delegate.OnListenFailed(err)
}

func (t *ServiceTask) serve(ctx context.Context, adapter IAdapter, wg *sync.WaitGroup) {
	defer t.cancel()
	defer wg.Done()
	defer func() {
		if r := recover(); r != nil {
			log.Printf("【Serve Panic】Recover: %v\nStack:\n%s", r, debug.Stack())
		}
	}()
	t.delegate.Serve(ctx, adapter)
}

func (t *ServiceTask) onStopping() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("【OnStopping Panic】Recover: %v\nStack:\n%s", r, debug.Stack())
		}
	}()
	t.delegate.OnStopping()
}

func (t *ServiceTask) onStopped() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("【OnStopped Panic】Recover: %v\nStack:\n%s", r, debug.Stack())
		}
	}()
	t.delegate.OnStopped()
}
