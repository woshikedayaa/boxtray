package boxtray

import (
	"context"
	"fmt"
	qt "github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/mainthread"
	"github.com/woshikedayaa/boxtray/cmd/boxtray/gui"
	"github.com/woshikedayaa/boxtray/common"
	"github.com/woshikedayaa/boxtray/common/capi"
	"github.com/woshikedayaa/boxtray/log"
	"log/slog"
	"os"
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
		for no := range ch {
			if no.Type != NotificationTypeStatus {
				continue
			}
			status := no.GetStatus()
			if status.Up && status.UpFromDown {
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
			} else if !status.Up {
				mainthread.Wait(func() {
					versionAction.SetText(defaultVersionText)
					trafficAction.SetText(defaultTrafficText)
					memoryAction.SetText(defaultMemoryText)
				})
			}
		}
		b.Unsubscribe(infoGuiSubscriberName)
	}()
}

func (b *Box) initControlGui(menu *qt.QMenu) {
	startAction := qt.NewQAction2("Started")
	startAction.SetCheckable(true)
	updateAction := qt.NewQAction2("Update")
	updateAction.SetCheckable(true)

	if len(b.config.Api.Control.Start) == 0 || len(b.config.Api.Control.Stop) == 0 {
		b.logger.Warn("start or stop command not configured, disable start action")
		startAction.SetDisabled(true)
	}
	if len(b.config.Api.Control.Update) == 0 {
		b.logger.Warn("update command not configured, disable update action")
		updateAction.SetDisabled(true)
	}

	startAction.OnTriggered(func() {
		if !startAction.IsEnabled() {
			return
		}
		if b.currentStatus.Load() {
			err := b.CloseManually()
			if err != nil {
				b.logger.Error("stop failed", slog.String("error", err.Error()))
			}
		} else {
			err := b.StartManually()
			if err != nil {
				b.logger.Error("start failed", slog.String("error", err.Error()))
			}
		}
	})
	updateAction.OnTriggered(func() {
		if !startAction.IsEnabled() {
			return
		}
		err := b.UpdateManually()
		if err != nil {
			b.logger.Error("update failed", slog.String("error", err.Error()))
		}
	})

	menu.AddAction(startAction)
	menu.AddAction(updateAction)
	const controlGuiSubscriberName = "control"
	ch := b.Subscribe(controlGuiSubscriberName)
	go func() {
		for no := range ch {
			if no.Type == NotificationTypeStatus {
				mainthread.Wait(
					func() {
						startAction.SetChecked(no.GetStatus().Up)
					})
			}
		}
		b.Unsubscribe(controlGuiSubscriberName)
	}()
}

func (b *Box) initBoxGui(menu *qt.QMenu) {
	quitAction := qt.NewQAction2("Quit")
	quitAction.OnTriggered(func() {
		b.logger.Info("Quit triggered,exit now !")
		b.cancel()
		os.Exit(0)
	})
	menu.AddAction(quitAction)
}

func (b *Box) initProxiesGui(menu *qt.QMenu) {
	var (
		proxiesMenus []*qt.QMenu
	)
	const proxiesNodeSubscribeName = "proxies-nodes"
	ch := b.Subscribe(proxiesNodeSubscribeName)
	go func() {
		for no := range ch {
			if no.Type != NotificationTypeStatus {
				return
			}
			status := no.GetStatus()
			if status.Up && status.UpFromDown {
				proxies, err := b.api.GetProxies()
				if err != nil {
					b.logger.Error("get proxies failed", slog.String("error", err.Error()))
					return
				}
				err = b.proxies.Parse(proxies)
				if err != nil {
					b.logger.Error("parse proxies failed", slog.String("error", err.Error()))
					return
				}
				mainthread.Wait(func() {
					selector := b.proxies.LoadSelector()
					for pair := selector.Oldest(); pair != nil; pair = pair.Next() {
						name, nodes := pair.Key, pair.Value
						subMenu := qt.NewQMenu3(name)
						b.addProxiesSelector(subMenu, common.MapSlice[*capi.Proxy, string, []*capi.Proxy, []string](nodes, func(idx int, source *capi.Proxy) string {
							return source.Name
						}), proxies.Proxies.Value(name).Now)
						// sync
						subMenu.SetTitle(gui.LatencyText(name,
							b.proxies.GetDelay(proxies.Proxies.Value(name).Now)))
						menu.AddMenu(subMenu)
						proxiesMenus = append(proxiesMenus, subMenu)
					}
				})
			} else if !status.Up {
				mainthread.Wait(func() {
					for _, v := range proxiesMenus {
						v.Hide()
						v.Destroy()
					}
				})
				proxiesMenus = nil
				b.logger.Info("service has down, remove all the proxies")
			}
		}
		b.Unsubscribe(proxiesNodeSubscribeName)
	}()
}
func (b *Box) addProxiesSelector(menu *qt.QMenu, nodes []string, now string) {
	var actions = make(map[string]*qt.QAction)

	refreshButton := qt.NewQAction2("Refresh")
	refreshButton.SetCheckable(false)
	refreshButton.SetEnabled(true)
	// todo : switch to goroutine pool
	refreshButton.OnTriggered(func() {
		for _, n := range nodes {
			go func(n string) {
				delay, err := b.api.GetDelay(n, "https://google.com/generate_204", 3000)
				_ = err
				mainthread.Wait(func() {
					actions[n].SetText(gui.LatencyText(n, delay.Delay))
				})
			}(n)
		}
		b.logger.Info("refresh delay finished")
	})
	menu.AddAction(refreshButton)
	menu.AddSeparator()
	actionGroup := qt.NewQActionGroup(nil)
	actionGroup.SetExclusive(true)
	for _, v := range nodes {
		act := qt.NewQAction2(v)
		act.SetCheckable(true)
		act.SetChecked(false)
		//mainthread.Wait(func() {
		//	act.SetText(gui.LatencyText(v, b.proxies.GetDelay(v)))
		//	if v == now {
		//		act.SetChecked(true)
		//	}
		//	actionGroup.AddAction(act)
		//	menu.AddAction(act)
		//})
		act.SetText(gui.LatencyText(v, b.proxies.GetDelay(v)))
		if v == now {
			act.SetChecked(true)
		}
		actionGroup.AddAction(act)
		menu.AddAction(act)
		actions[v] = act
	}
}
