package boxtray

import (
	"fmt"
	"fyne.io/systray"
	"github.com/woshikedayaa/boxtray/common/capi"
	"github.com/woshikedayaa/boxtray/config"
	"github.com/woshikedayaa/boxtray/log"
	"log/slog"
	"os"
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
	global = InitGlobal(cfg, client)
	global.started.Store(true)
	systray.Run(onReady, onExit)
}
