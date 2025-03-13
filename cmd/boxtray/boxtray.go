package boxtray

import (
	"context"
	"fmt"
	qt "github.com/mappu/miqt/qt6"
	"github.com/woshikedayaa/boxtray/common"
	"github.com/woshikedayaa/boxtray/common/capi"
	"github.com/woshikedayaa/boxtray/config"
	"github.com/woshikedayaa/boxtray/log"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

func Main(cfg config.Config) {
	err := log.Init(cfg.Log)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "level: ", err.Error())
		return
	}
	logger := log.Get("init")
	client, err := capi.NewClient(cfg.Api.Endpoint(), &capi.ClientConfig{
		Timeout: 10 * time.Second,
		Secret:  cfg.Api.Secret,
	})
	logger.Info("Set endpoint", slog.String("endpoint", cfg.Api.Endpoint()))
	if err != nil {
		logger.Error("Create Client", slog.String("error", err.Error()))
		return
	}
	NewBox(client, cfg).RunLoop(context.Background())
}

type BoxNotificationType uint16

const (
	NotificationTypeError BoxNotificationType = iota
	NotificationTypeStatus
)

type BoxNotification struct {
	Type    BoxNotificationType
	Message any
}

func (no *BoxNotification) GetStatus() BoxStatus {
	if no.Type == NotificationTypeStatus {
		return no.Message.(BoxStatus)
	}
	panic("incorrect notification type")
}

func (no *BoxNotification) GetError() error {
	if no.Type == NotificationTypeError {
		return no.Message.(error)
	}
	panic("incorrect notification type")
}

type BoxStatus struct {
	Up         bool
	UpFromDown bool
}
type Box struct {
	ctx context.Context

	currentStatus atomic.Bool
	api           *capi.Client
	subscribers   map[string]chan BoxNotification
	logger        *log.Logger

	config config.Config
	cancel context.CancelFunc
	mu     sync.RWMutex
}

func NewBox(client *capi.Client, cfg config.Config) *Box {
	return &Box{
		api:         client,
		subscribers: make(map[string]chan BoxNotification),
		config:      cfg,
	}
}

func (b *Box) RunLoop(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}

	b.ctx, b.cancel = context.WithCancel(ctx)
	defer b.cancel()
	b.logger = log.Get("main")
	_ = qt.NewQApplication(nil)
	tray := qt.NewQSystemTrayIcon2(qt.QApplication_Style().StandardIcon(qt.QStyle__SP_ComputerIcon, nil, nil))
	rootMenu := qt.NewQMenu2()
	tray.SetContextMenu(rootMenu)
	tray.Show()

	b.initInfoGui(rootMenu)
	rootMenu.AddSeparator()
	b.initControlGui(rootMenu)
	rootMenu.AddSeparator()
	b.initBoxGui(rootMenu)
	rootMenu.AddSeparator()
	b.initProxiesGui(rootMenu)
	go b.notificationPublisher(b.ctx)
	os.Exit(qt.QApplication_Exec())
}

func (b *Box) boardCast(notification BoxNotification) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for name, sub := range b.subscribers {
		switch notification.Type {
		// triangle , lmao
		case NotificationTypeError:
			go func() {
				select {
				case sub <- notification:
				case <-time.After(300 * time.Millisecond):
					if notification.Message != nil {
						if e, ok := notification.Message.(error); ok {
							b.logger.Error("time out when sending a error notification", slog.String("error", e.Error()), slog.String("name", name))
							return
						}
					}
					panic(fmt.Sprintf("a error occurred while handling a error: not a standing error: %v", notification.Message))
				}
			}()
		case NotificationTypeStatus:
			go func() {
				select {
				case sub <- notification:
				case <-time.After(5 * time.Second):
					b.logger.Warn("notification to channel spend too much time!", slog.String("name", name), slog.String("Status", fmt.Sprintf("%s", notification.Message)))
				}
			}()
		}
	}
	if notification.Type == NotificationTypeError {
		b.currentStatus.Store(false)
	}
}

func (b *Box) CloseManually() error {
	if len(b.config.Api.Control.Start) == 0 {
		return fmt.Errorf("start command not configured")
	}
	return common.RunOneShot(b.ctx, b.config.Api.Control.Start[0], b.config.Api.Control.Start[1:])
}
func (b *Box) StartManually() error {
	if len(b.config.Api.Control.Stop) == 0 {
		return fmt.Errorf("stop command not configured")
	}
	return common.RunOneShot(b.ctx, b.config.Api.Control.Stop[0], b.config.Api.Control.Stop[1:])
}

func (b *Box) UpdateManually() error {
	if len(b.config.Api.Control.Update) == 0 {
		return fmt.Errorf("update command not configured")
	}
	return common.RunOneShot(b.ctx, b.config.Api.Control.Update[0], b.config.Api.Control.Update[1:])
}

func (b *Box) Subscribe(name string) <-chan BoxNotification {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan BoxNotification)
	if _, exist := b.subscribers[name]; exist {
		panic("duplicated subscriber")
	}
	b.subscribers[name] = ch
	return ch
}

func (b *Box) Unsubscribe(name string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if ch, exist := b.subscribers[name]; exist {
		close(ch)
		delete(b.subscribers, name)
	}
}

func (b *Box) notificationPublisher(ctx context.Context) {
	ret := make(chan error)
	next := make(chan struct{})
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	go func() {
		for range next {
			_, err := b.api.GetVersion()
			if err != nil {
				ret <- err
			}
		}
	}()
	for range ticker.C {
		select {
		case err := <-ret:
			if err == nil {
				continue
			}
			if b.currentStatus.Load() {
				b.logger.Error("status check failed", slog.String("error", err.Error()))
				b.logger.Warn("detect service down")
				b.boardCast(BoxNotification{
					Type:    NotificationTypeError,
					Message: err,
				})
				b.boardCast(BoxNotification{
					Type: NotificationTypeStatus,
					Message: BoxStatus{
						Up:         false,
						UpFromDown: false,
					},
				})
				b.currentStatus.Store(false)
			}
			next <- struct{}{}
		case <-ctx.Done():
			close(next)
		default:
			if !b.currentStatus.Load() {
				b.logger.Warn("detect service available now")
			}
			// no error
			b.boardCast(BoxNotification{
				Type: NotificationTypeStatus,
				Message: BoxStatus{
					Up:         true,
					UpFromDown: !b.currentStatus.Load(),
				},
			})
			b.currentStatus.Store(true)
			next <- struct{}{}
		}
	}
}
