package boxtray

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/woshikedayaa/boxtray/common/capi"
	"github.com/woshikedayaa/boxtray/config"
	"github.com/woshikedayaa/boxtray/log"
)

type StatusChange int

const (
	StatusStarted StatusChange = iota
	StatusStopped
)

// StatusEvent 表示状态事件
type StatusEvent struct {
	Status  StatusChange
	IsError bool
}

type Global struct {
	config  config.Config
	client  *capi.Client
	started atomic.Bool

	statusMu        sync.RWMutex
	statusChan      chan StatusEvent
	statusObservers []chan StatusEvent

	checkCtx    context.Context
	checkCancel context.CancelFunc
}

func InitGlobal(cfg config.Config, client *capi.Client) *Global {
	ctx, cancel := context.WithCancel(context.Background())
	g := &Global{
		config:          cfg,
		client:          client,
		statusChan:      make(chan StatusEvent, 10),
		statusObservers: make([]chan StatusEvent, 0),
		checkCtx:        ctx,
		checkCancel:     cancel,
	}

	go g.broadcastStatus()
	go g.runStatusChecker()

	return g
}

func (g *Global) SetStarted(started bool, isError bool) {
	g.started.Store(started)

	status := StatusStopped
	if started {
		status = StatusStarted
	}

	g.statusChan <- StatusEvent{
		Status:  status,
		IsError: isError,
	}
}

func (g *Global) IsStarted() bool {
	return g.started.Load()
}

func (g *Global) Subscribe() <-chan StatusEvent {
	g.statusMu.Lock()
	defer g.statusMu.Unlock()

	ch := make(chan StatusEvent, 5)
	g.statusObservers = append(g.statusObservers, ch)
	return ch
}

func (g *Global) Unsubscribe(ch <-chan StatusEvent) {
	g.statusMu.Lock()
	defer g.statusMu.Unlock()

	for i, observer := range g.statusObservers {
		if observer == ch {
			g.statusObservers = append(g.statusObservers[:i], g.statusObservers[i+1:]...)
			close(observer)
			break
		}
	}
}

func (g *Global) broadcastStatus() {
	for event := range g.statusChan {
		g.statusMu.RLock()
		for _, observer := range g.statusObservers {
			select {
			case observer <- event:
			default:
			}
		}
		g.statusMu.RUnlock()
	}
}

func (g *Global) runStatusChecker() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	g.checkStatus()

	for {
		select {
		case <-ticker.C:
			g.checkStatus()
		case <-g.checkCtx.Done():
			return
		}
	}
}

func (g *Global) checkStatus() {
	logger := log.Get("status-checker")

	_, err := g.client.GetVersion()
	if err != nil {
		// logger.Error("Backend service check failed", "error", err.Error())

		if g.IsStarted() {
			logger.Warn("Backend service is down, updating status", slog.String("error", err.Error()))
			g.SetStarted(false, true)
		}
	} else {
		if !g.IsStarted() {
			logger.Info("Backend service is up, updating status")
			g.SetStarted(true, false)
		}
	}
}

func (g *Global) StopStatusChecker() {
	if g.checkCancel != nil {
		g.checkCancel()
	}
}

var global *Global
