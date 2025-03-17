package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/woshikedayaa/boxtray/cmd/boxtray"
	"github.com/woshikedayaa/boxtray/cmd/boxtray/metadata"
	"github.com/woshikedayaa/boxtray/common"
	"github.com/woshikedayaa/boxtray/config"
	"github.com/woshikedayaa/boxtray/log"
	"os"
)

var (
	configFile string
	version    bool
)

func init() {
	flag.StringVar(&configFile, "c", "~/.config/boxtray/config.json", "Set the config file")
	flag.BoolVar(&version, "version", false, "display version")
	flag.Parse()
}

func main() {
	if version {
		fmt.Println(metadata.Version)
		return
	}
	configFile, err := common.ExpandHomePath(configFile)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	file, err := os.Open(configFile)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	cfg := config.Config{}
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cfg); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	_ = file.Close()

	if err := log.Init(cfg.Log); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "level: ", err.Error())
		os.Exit(1)
	}
	os.Exit(boxtray.Main(cfg))
}
