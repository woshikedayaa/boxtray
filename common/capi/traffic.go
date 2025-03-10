package capi

import (
	"context"
	"encoding/json"
)

type Traffic struct {
	Up   int `json:"up"`
	Down int `json:"down"`
}

func (c *Client) GetTraffic(parentCtx context.Context, handler func(traffic Traffic, stop context.CancelFunc)) error {
	ctx, cancelFunc := context.WithCancel(parentCtx)
	defer cancelFunc()
	stream, errorC, err := c.doGetStream(ctx, "/traffic", nil)
	if err != nil {
		return err
	}
	for bs := range stream {
		tf := Traffic{}
		err := json.Unmarshal(bs, &tf)
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case e := <-errorC:
			return e
		default:
			handler(tf, cancelFunc)
		}
	}
	select {
	case e, ok := <-errorC:
		if ok && e != nil {
			return e
		}
		return nil
	default:
		return nil
	}
}
