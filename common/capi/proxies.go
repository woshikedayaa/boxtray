package capi

import "time"

type Proxies struct {
	Proxies map[string]Proxy `json:"proxies"`
}

type History struct {
	Time  time.Time `json:"time"`
	Delay uint16    `json:"delay"`
}
type Proxy struct {
	Type    string    `json:"type"`
	Name    string    `json:"name"`
	UDP     bool      `json:"udp"`
	History []History `json:"history"`
	//
	Now string   `json:"now"`
	All []string `json:"all"`
}
