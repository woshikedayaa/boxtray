package boxtray

import (
	"context"
	_ "embed"
	"fmt"
	qt "github.com/mappu/miqt/qt6"
	"github.com/woshikedayaa/boxtray/common"
	"github.com/woshikedayaa/boxtray/common/capi"
	"github.com/woshikedayaa/boxtray/config"
	"github.com/woshikedayaa/boxtray/log"
	"log/slog"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

//go:embed resources/singbox.ico
var icoByte []byte

func Main(cfg config.Config) int {
	logger := log.Get("init")
	client, err := capi.NewClient(cfg.Api.Endpoint(), &capi.ClientConfig{
		Timeout: 10 * time.Second,
		Secret:  cfg.Api.Secret,
	})
	logger.Info("Set endpoint", slog.String("endpoint", cfg.Api.Endpoint()))
	if err != nil {
		logger.Error("Create Client", slog.String("error", err.Error()))
		return 1
	}

	box := NewBox(client, cfg)
	return box.RunLoop(context.Background())
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
	ctx    context.Context
	cancel context.CancelFunc
	config config.Config

	// Status
	currentStatus    atomic.Bool
	api              *capi.Client
	subscribers      *sync.Map //map[string]chan BoxNotification
	subscribersCount atomic.Int32
	logger           *log.Logger

	proxies *ProxiesManager
}

func NewBox(client *capi.Client, cfg config.Config) *Box {
	if cfg.Box.UrlTest == "" {
		cfg.Box.UrlTest = "https://google.com/generate_204"
	}
	if cfg.Box.MaxDelay < 1 {
		cfg.Box.MaxDelay = 3000
	}
	b := &Box{
		api:              client,
		subscribers:      &sync.Map{},
		subscribersCount: atomic.Int32{},
		config:           cfg,
		proxies:          NewProxiesManager(),
		logger:           log.Get("main"),
	}
	return b
}

func (b *Box) initGui() {
	_ = qt.NewQApplication([]string{"boxtray"})
	rootMenu := qt.NewQMenu2()
	b.initInfoGui(rootMenu)
	rootMenu.AddSeparator()
	b.initControlGui(rootMenu)
	rootMenu.AddSeparator()
	b.initBoxGui(rootMenu)
	rootMenu.AddSeparator()
	b.initProxiesGui(rootMenu)

	icoPixMap := qt.NewQPixmap()
	icoPixMap.LoadFromData2(icoByte, "")
	tray := qt.NewQSystemTrayIcon2(qt.NewQIcon2(icoPixMap))
	tray.SetContextMenu(rootMenu)
	tray.Show()
}

func (b *Box) clean() {
	if b.cancel != nil {
		b.cancel()
	}
	for b.subscribersCount.Load() != 0 {
	}
}

func (b *Box) RunLoop(ctx context.Context) int {
	if ctx == nil {
		panic("nil context")
	}
	b.initGui()
	b.ctx, b.cancel = context.WithCancel(ctx)
	defer b.clean()
	go b.notificationPublisher(b.ctx)
	return qt.QApplication_Exec()
}

func (b *Box) broadCast(notification BoxNotification) {
	b.subscribers.Range(func(key, value any) bool {
		name := key.(string)
		sub := value.(chan BoxNotification)
		switch notification.Type {
		// triangle , lmao
		case NotificationTypeError:
			go func() {
				select {
				case sub <- notification:
				case <-time.After(1 * time.Second):
					if notification.Message != nil {
						if e, ok := notification.Message.(error); ok {
							b.logger.Error("notification spend too much time!", slog.String("error", e.Error()), slog.String("type", "error"), slog.String("name", name))
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
				case <-time.After(1 * time.Second):
					b.logger.Warn("notification spend too much time!", slog.String("name", name), slog.String("type", "status"), slog.String("Status", fmt.Sprintf("%s", strconv.FormatBool(notification.GetStatus().Up))))
				}
			}()
		}
		return true
	})
	if notification.Type == NotificationTypeError {
		b.currentStatus.Store(false)
	}
}

func (b *Box) CloseManually() error {
	if len(b.config.Api.Control.Stop) == 0 {
		return fmt.Errorf("stop command not configured")
	}
	b.logger.Debug("close now", slog.String("command", fmt.Sprint(b.config.Api.Control.Stop)))
	return common.RunOneShot(b.ctx, b.config.Api.Control.Stop[0], b.config.Api.Control.Stop[1:])
}
func (b *Box) StartManually() error {
	if len(b.config.Api.Control.Start) == 0 {
		return fmt.Errorf("start command not configured")
	}
	b.logger.Debug("start now", slog.String("command", fmt.Sprint(b.config.Api.Control.Start)))
	return common.RunOneShot(b.ctx, b.config.Api.Control.Start[0], b.config.Api.Control.Start[1:])
}

func (b *Box) UpdateManually() error {
	if len(b.config.Api.Control.Update) == 0 {
		return fmt.Errorf("update command not configured")
	}
	b.logger.Debug("update now", slog.String("command", fmt.Sprint(b.config.Api.Control.Update)))
	return common.RunOneShot(b.ctx, b.config.Api.Control.Update[0], b.config.Api.Control.Update[1:])
}

func (b *Box) Subscribe(name string) <-chan BoxNotification {
	ch := make(chan BoxNotification)
	if _, exist := b.subscribers.Load(name); exist {
		panic("duplicated subscriber")
	}
	b.subscribers.Store(name, ch)
	b.subscribersCount.Add(1)
	b.logger.Debug("new subscribe", slog.String("name", name))
	return ch
}

func (b *Box) Unsubscribe(name string) {
	if ch, exist := b.subscribers.Load(name); exist {
		close(ch.(chan BoxNotification))
		b.subscribers.Delete(name)
		b.subscribersCount.Add(-1)
		b.logger.Debug("unsubscribe", slog.String("name", name))
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
	next <- struct{}{}
	for range ticker.C {
		select {
		case err := <-ret:
			if err == nil {
				continue
			}
			if b.currentStatus.Load() {
				b.logger.Warn("detect service down", slog.String("error", err.Error()))
				b.broadCast(BoxNotification{
					Type:    NotificationTypeError,
					Message: err,
				})
				b.broadCast(BoxNotification{
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
			return
		default:
			if !b.currentStatus.Load() {
				b.logger.Warn("detect service available now")
			}
			// no error
			b.broadCast(BoxNotification{
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
