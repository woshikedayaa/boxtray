package main

import (
	"github.com/woshikedayaa/boxtray/cmd/boxtray"
	"github.com/woshikedayaa/boxtray/config"
)

func main() {
	boxtray.Main(config.Config{Log: config.LogConfig{Level: "INFO"}, Api: config.ApiConfig{
		Scheme: "http",
		Host:   "localhost:9090",
		Path:   "",
		Label:  "Main",
		Secret: "20002000",
		Control: config.ControlConfig{
			Start:  []string{"systemctl", "start", "sing-box"},
			Stop:   []string{"systemctl", "stop", "sing-box"},
			Update: nil,
		},
	}})
}
