package main

import (
	"context"
	"fmt"
	"github.com/woshikedayaa/boxtray/common/capi"
)

func main() {
	client, err := capi.NewClient(context.Background(), "http://localhost:9090", &capi.ClientConfig{Secret: "20002000"})
	if err != nil {
		panic(err)
	}
	version, err := client.GetVersion()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v", version)
}
