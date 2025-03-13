package task

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

type Func func(ctx context.Context) error

type taskItem struct {
	name string
	err  error
	task Func
}

type taskError struct {
	err  error
	name string
}

var errTaskSuccess = &taskError{err: errors.New("success")}

func (t *taskError) Error() string {
	if t == nil {
		return "<nil>"
	}
	if t.name == "" {
		return fmt.Sprintf("task : %s", t.err.Error())
	}
	return fmt.Sprintf("task %s: %s", t.name, t.err.Error())
}

func (t *taskError) Unwrap() error {
	if t == nil {
		return nil
	}
	return t.err
}

func (t taskItem) Error() error {
	if t.err == nil {
		return nil
	}
	return &taskError{t.err, t.name}
}

func (t taskItem) Run(ctx context.Context) {
	t.err = t.task(ctx)
}

type Group struct {
	// global mutex
	*sync.RWMutex

	tasks    []taskItem
	fastFail bool
	queue    chan struct{}
	n        int
	clean    func()
}

func NewGroup() *Group {
	g := &Group{
		tasks:    make([]taskItem, 0),
		RWMutex:  &sync.RWMutex{},
		fastFail: true,
	}
	return g
}

func (g *Group) Clean(c func()) {
	g.RLock()
	defer g.RUnlock()
	g.clean = c
}

// Concurrency set up the max goroutine count
// set n < 1 to disable Concurrency;
func (g *Group) Concurrency(n int) {
	g.RLock()
	defer g.RUnlock()
	if n < 1 {
		if g.queue != nil {
			close(g.queue)
		}
		g.queue = nil
		return
	}
	g.queue = make(chan struct{}, n)
	for i := 0; i < n; i++ {
		g.queue <- struct{}{}
	}
}

func (g *Group) Run(ctx context.Context) error {
	if g == nil {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	upstreamContext := ctx
	upstreamCancelContext, taskCanceled := context.WithCancelCause(ctx)
	taskContext, taskFinish := context.WithCancelCause(context.Background())
	// wait lock
	// when groups tasks after starting the groups  can't be edited.
	g.Lock()
	errAccess := &sync.Mutex{}
	taskCount := int32(len(g.tasks))

	var ea error
	for _, tk := range g.tasks {
		go func(currentTask taskItem) {

			// queue
			if g.queue != nil {
				select {
				case <-taskContext.Done():
					nt := atomic.AddInt32(&taskCount, -1)
					if nt == 0 {
						taskCanceled(errTaskSuccess)
						taskFinish(errTaskSuccess)
					}
					return
				case <-g.queue:
				}
			}
			currentTask.Run(upstreamCancelContext)
			if err := currentTask.Error(); err != nil {
				// fastFail handle
				if g.fastFail {
					taskCanceled(err)
				}
				errAccess.Lock()
				ea = errors.Join(ea, err)
				errAccess.Unlock()
			}
			// count--
			nt := atomic.AddInt32(&taskCount, -1)
			if nt == 0 {
				taskCanceled(errTaskSuccess)
				taskFinish(errTaskSuccess)
			}

			if g.queue != nil {
				g.queue <- struct{}{}
			}

		}(tk)
	}
	g.Unlock()
	var isUpstreamError bool
	select {
	case <-upstreamCancelContext.Done():
		break
	case <-upstreamContext.Done():
		isUpstreamError = true
		taskCanceled(upstreamContext.Err())
	}
	<-taskContext.Done()
	g.clean()
	if isUpstreamError {
		return upstreamContext.Err()
	}
	return ea
}

func (g *Group) DisableFastFail() {
	g.RLock()
	g.fastFail = false
	g.RUnlock()
}

func (g *Group) Append(name string, f Func) {
	g.RLock()
	g.tasks = append(g.tasks, taskItem{name: name, task: f})
	g.RUnlock()
}
