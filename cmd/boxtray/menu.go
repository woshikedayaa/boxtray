package boxtray

import (
	"context"
	"fmt"
	"fyne.io/systray"
	"fyne.io/systray/example/icon"
	"github.com/woshikedayaa/boxtray/common"
	"github.com/woshikedayaa/boxtray/common/capi"
	"github.com/woshikedayaa/boxtray/common/constant"
	"github.com/woshikedayaa/boxtray/common/gui"
	"github.com/woshikedayaa/boxtray/log"
	"log/slog"
	"maps"
	"os"
	"slices"
	"strings"
	"sync"
	"time"
)

func onReady() {
	systray.SetIcon(icon.Data)
	systray.SetTitle("boxtray")
	systray.SetTooltip("Boxtray")

	infoItems := initInfoGui()
	systray.AddSeparator()
	controlItems := initControlGui()
	systray.AddSeparator()
	initMainGui()
	systray.AddSeparator()
	proxiesItems := initProxiesGui()
	_, _, _ = infoItems, controlItems, proxiesItems
}

type InfoItems struct {
	backend     *systray.MenuItem
	memoryItem  *systray.MenuItem
	trafficItem *systray.MenuItem
}

func initInfoGui() *InfoItems {
	backend := systray.AddMenuItem(global.config.Api.DisplayName(), "Name")
	memoryItem := systray.AddMenuItem("Memory", "Memory")
	trafficItem := systray.AddMenuItem("Traffic", "Traffic")

	backend.Disable()
	memoryItem.Disable()
	trafficItem.Disable()

	items := &InfoItems{
		backend:     backend,
		memoryItem:  memoryItem,
		trafficItem: trafficItem,
	}

	statusCh := global.Subscribe()

	go func() {
		for event := range statusCh {
			if event.Status == StatusStarted {
				backend.Enable()
				memoryItem.Enable()
				trafficItem.Enable()
				go fetchInfo(items)
			} else {
				backend.SetTitle(global.config.Api.DisplayName() + " (Offline)")
				memoryItem.SetTitle("Memory")
				trafficItem.SetTitle("Traffic")
				backend.Disable()
				memoryItem.Disable()
				trafficItem.Disable()
			}
		}
	}()

	return items
}

func fetchInfo(items *InfoItems) {
	logger := log.Get("info")

	version, err := global.client.GetVersion()
	if err != nil {
		logger.Error("Get Version Failed", slog.String("error", err.Error()))
		return
	}

	items.backend.SetTitle(version.Version)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		if err := global.client.GetMemory(ctx, func(memory capi.Memory, stop context.CancelFunc) {
			items.memoryItem.SetTitle(fmt.Sprintf("%s", common.MemoryText(memory.Inuse)))
		}); err != nil {
			logger.Error("Get Memory Failed", slog.String("error", err.Error()))
		}
	}()

	go func() {
		defer wg.Done()
		if err := global.client.GetTraffic(ctx, func(traffic capi.Traffic, stop context.CancelFunc) {
			upText, downText := common.TrafficText(traffic.Up), common.TrafficText(traffic.Down)
			placeHold := ""
			const maxLength = 44
			if len(upText)+len(downText) < maxLength {
				placeHold = strings.Repeat(" ", maxLength-(len(upText)+len(downText)))
			}
			items.trafficItem.SetTitle(fmt.Sprintf("↑ %s ↓ %s%s", upText, downText, placeHold))
		}); err != nil {
			logger.Error("Get Traffic Failed", slog.String("error", err.Error()))
		}
	}()

	wg.Wait()
}

type ProxiesItems struct {
	// len(radio) == len(nodes)
	radio []*gui.RadioGroup
	nodes []map[string]*systray.MenuItem
}

func initProxiesGui() *ProxiesItems {
	logger := log.Get("proxies")
	items := &ProxiesItems{}
	subCh := global.Subscribe()
	go func() {
		currentStatus := StatusUnknown
		for event := range subCh {
			if event.Status == StatusStopped {
				currentStatus = StatusStopped
				continue
			}
			if currentStatus == StatusStarted && currentStatus == event.Status {
				continue
			}
			currentStatus = StatusStarted

			proxies, err := global.client.GetProxies()
			if err != nil {
				logger.Error("Get proxies", slog.String("error", err.Error()))
				continue
			}
			for _, v := range items.radio {
				v.Remove()
			}
			items.radio = nil
			items.nodes = nil
			keys := slices.Sorted(maps.Keys(proxies.Proxies))
			for _, k := range keys {
				node := proxies.Proxies[k]
				if node.Type != constant.ProxyDisplayName(constant.TypeURLTest) && node.Type != constant.ProxyDisplayName(constant.TypeSelector) {
					continue
				}
				radio := gui.NewRadioGroup(node.Name, node.Name, func(i int) {
					innerErr := global.client.SwitchProxy(node.Name, node.All[i])
					if innerErr != nil {
						logger.Error("Switch proxy", slog.String("error", innerErr.Error()))
						return
					}
					logger.Info("Switch proxy", slog.String("selector", node.Name), slog.String("node", node.All[i]))
				})
				nodeSet := make(map[string]*systray.MenuItem, len(node.All))
				for _, proxyName := range node.All {
					item := radio.AddItem(proxyName, "")
					nodeSet[proxyName] = item
				}
				radio.Select(slices.Index(node.All, node.Now))

				items.radio = append(items.radio, radio)
				items.nodes = append(items.nodes, nodeSet)
			}
		}
	}()

	return items
}

type ControlItems struct {
	startButton  *systray.MenuItem
	updateButton *systray.MenuItem
}

func initControlGui() *ControlItems {
	logger := log.Get("control")
	startButton := systray.AddMenuItemCheckbox("Started", "", false)
	updateButton := systray.AddMenuItem("Update", "Update config file")

	items := &ControlItems{
		startButton:  startButton,
		updateButton: updateButton,
	}

	statusCh := global.Subscribe()

	go func() {
		for event := range statusCh {
			if event.Status == StatusStarted {
				startButton.Check()
			} else {
				startButton.Uncheck()
			}
		}
	}()
	if len(global.config.Api.Control.Start) == 0 || len(global.config.Api.Control.Stop) == 0 {
		logger.Warn("Start or Stop command not configured")
		startButton.Disable()
		updateButton.Disable()
		return items
	}

	go func() {
		for range startButton.ClickedCh {
			if startButton.Checked() {
				logger.Info("Stopping Singbox")
				timeout, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
				if err := common.RunOneShot(timeout, global.config.Api.Control.Stop[0], global.config.Api.Control.Stop[1:]); err != nil {
					logger.Error("Stopping Singbox", slog.String("error", err.Error()))
					cancelFunc()
					continue
				}
				cancelFunc()
				global.SetStarted(false, false)
				logger.Info("Singbox Stopped")
			} else {
				logger.Info("Starting Singbox")
				timeout, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
				if err := common.RunOneShot(timeout, global.config.Api.Control.Start[0], global.config.Api.Control.Start[1:]); err != nil {
					logger.Error("Starting Singbox", slog.String("error", err.Error()))
					cancelFunc()
					continue
				}
				cancelFunc()
				global.SetStarted(true, false)
				logger.Info("Singbox Started")
			}
		}
	}()

	go func() {
		if len(global.config.Api.Control.Update) == 0 {
			updateButton.Disable()
			logger.Warn("Update command not configured")
			return
		}

		for range updateButton.ClickedCh {
			timeout, cancelFunc := context.WithTimeout(context.Background(), 30*time.Second)
			if err := common.RunOneShot(timeout, global.config.Api.Control.Update[0], global.config.Api.Control.Update[1:]); err != nil {
				logger.Error("Update Singbox", slog.String("error", err.Error()))
				cancelFunc()
				continue
			}
			cancelFunc()
		}
	}()

	return items
}

func initMainGui() {
	mQuit := systray.AddMenuItem("Quit", "")
	go func() {
		<-mQuit.ClickedCh
		os.Exit(0)
	}()
}

func onExit() {
	if global != nil {
		global.StopStatusChecker()
	}
}
