package boxtray

import (
	"context"
	"fmt"
	qt "github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/mainthread"
	"github.com/woshikedayaa/boxtray/common"
	"github.com/woshikedayaa/boxtray/common/capi"
	"github.com/woshikedayaa/boxtray/log"
	"log/slog"
)

func (b *Box) initInfoGui(menu *qt.QMenu) {
	const (
		infoGuiSubscriberName = "information"
		defaultVersionText    = "Version(offline Now)"
		defaultTrafficText    = "Traffic"
		defaultMemoryText     = "Memory"
	)
	logger := log.Get(infoGuiSubscriberName)

	versionAction := qt.NewQAction2(defaultVersionText)
	trafficAction := qt.NewQAction2(defaultTrafficText)
	memoryAction := qt.NewQAction2(defaultMemoryText)
	versionAction.SetCheckable(false)
	trafficAction.SetCheckable(false)
	memoryAction.SetCheckable(false)
	versionAction.SetToolTip("Version")
	trafficAction.SetToolTip("Traffic")
	memoryAction.SetToolTip("Memory")

	menu.AddActions([]*qt.QAction{versionAction, trafficAction, memoryAction})
	ch := b.Subscribe(infoGuiSubscriberName)

	go func() {
		cur := false
		for no := range ch {
			if no.Type != NotificationTypeStatus {
				continue
			}
			started := no.Message.(bool)
			if started && !cur {
				version, err := b.api.GetVersion()
				if err != nil {
					logger.Error("get version failed", slog.String("error", err.Error()))
					continue
				}
				mainthread.Wait(func() {
					versionAction.SetText(version.Version)
				})
				ctx, cancel := context.WithCancel(b.ctx)
				go func() {
					if err := b.api.GetTraffic(ctx, func(traffic capi.Traffic, stop context.CancelFunc) {
						mainthread.Wait(func() {
							trafficAction.SetText(fmt.Sprintf("↑ %s↓ %s", common.TrafficText(traffic.Up), common.TrafficText(traffic.Down)))
						})
						if !b.currentStatus.Load() {
							stop()
						}
					}); err != nil {
						logger.Error("get traffic failed", slog.String("error", err.Error()))
					}
					cancel()
				}()
				go func() {
					if err := b.api.GetMemory(ctx, func(memory capi.Memory, stop context.CancelFunc) {
						mainthread.Wait(func() {
							memoryAction.SetText(fmt.Sprintf("%s", common.MemoryText(memory.Inuse)))
						})
						if !b.currentStatus.Load() {
							stop()
						}
					}); err != nil {
						logger.Error("get memory failed", slog.String("error", err.Error()))
					}
					cancel()
				}()

				cur = true
			} else if !no.Message.(bool) {
				mainthread.Wait(func() {
					versionAction.SetText(defaultVersionText)
					trafficAction.SetText(defaultTrafficText)
					memoryAction.SetText(defaultMemoryText)
				})
				cur = false
			}
		}
		b.Unsubscribe(infoGuiSubscriberName)
	}()
}
