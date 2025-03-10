package main

import (
	"context"
	"fmt"
	"github.com/woshikedayaa/boxtray/common/capi"
	"github.com/woshikedayaa/boxtray/common/constant"
	"time"
)

func main() {
	client, err := capi.NewClient("http://localhost:9090", &capi.ClientConfig{Secret: "20002000", Timeout: 100 * time.Minute})
	if err != nil {
		panic(err)
	}
	version, err := client.GetVersion()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", version)

	proxies, err := client.GetProxies()
	if err != nil {
		panic(err)
	}
	for k, v := range proxies.Proxies {
		if v.Type == constant.ProxyDisplayName(constant.TypeSelector) {
			fmt.Printf("%s - %d current: %s\n", v.Name, len(v.All), v.Now)
		}
		_ = k
	}

	if e := client.GetTraffic(context.Background(), func(traffic capi.Traffic, stop context.CancelFunc) {
		fmt.Printf("%#v\n", traffic)
	}); e != nil {
		panic(e)
	}
}
