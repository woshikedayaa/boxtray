package boxtray

import (
	"fmt"
	"github.com/woshikedayaa/boxtray/common/capi"
	"github.com/woshikedayaa/boxtray/common/constant"
	"github.com/woshikedayaa/boxtray/log"
	"log/slog"
	"strings"
	"sync/atomic"
)

type ProxiesManager struct {
	//
	// Copy on Write
	//
	selectors atomic.Pointer[map[string][]*capi.Proxy]
	delays    atomic.Pointer[map[string]uint16]

	logger *slog.Logger
}

func NewProxiesManager() *ProxiesManager {
	p := &ProxiesManager{}
	selectors, delays := make(map[string][]*capi.Proxy), make(map[string]uint16)
	p.selectors.Store(&selectors)
	p.delays.Store(&delays)
	p.logger = log.Get("proxies-manager")
	return p
}

func (p *ProxiesManager) Parse(proxies *capi.Proxies) error {
	if len(proxies.Proxies) == 0 {
		return fmt.Errorf("empty proxies")
	}

	selectors, delays := make(map[string][]*capi.Proxy), make(map[string]uint16)
	for name, proxy := range proxies.Proxies {
		switch strings.ToLower(proxy.Type) {
		case "":
			continue
		case constant.TypeSelector, constant.TypeURLTest:
			if len(proxy.All) == 0 {
				p.logger.Warn("selector is empty,skip it", slog.String("selector", name))
				continue
			}

			var nodes []*capi.Proxy
			for _, node := range proxy.All {
				if proxyNode, ok := proxies.Proxies[node]; ok {
					nodes = append(nodes, proxyNode)
				} else {
					p.logger.Warn("selector contain a non-existed outbound", slog.String("selector", name), slog.String("outbound", node))
				}
			}
			selectors[name] = nodes

		default:
			for _, his := range proxy.History {
				delays[name] = his.Delay
			}
		}
	}
	for k, _ := range selectors {
		if d, exist := delays[proxies.Proxies[k].Now]; exist {
			delays[k] = d
		}
	}

	p.selectors.Store(&selectors)
	p.delays.Store(&delays)
	return nil
}

func (p *ProxiesManager) LoadSelector() map[string][]*capi.Proxy {
	return *p.selectors.Load()
}
func (p *ProxiesManager) LoadDelay() map[string]uint16 {
	return *p.delays.Load()
}

func (p *ProxiesManager) GetDelay(name string) uint16 {
	delay := p.LoadDelay()
	return delay[name]
}
