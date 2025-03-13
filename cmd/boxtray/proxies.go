package boxtray

import "sync"

type ProxiesManager struct {
	Selectors sync.Map // map[string]*capi.Proxy

}
