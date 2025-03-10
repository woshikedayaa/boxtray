package capi

import (
	"context"
	"encoding/json"
)

type Memory struct {
	Inuse   int `json:"inuse"`
	Oslimit int `json:"oslimit"`
}

func (c *Client) GetMemory(parentCtx context.Context, handler func(memory Memory, stop context.CancelFunc)) error {
	const MemoryPath = "/memory"
	ctx, cancelFunc := context.WithCancel(parentCtx)
	defer cancelFunc()
	stream, errorC, err := c.doGetStream(ctx, MemoryPath, nil)
	if err != nil {
		return err
	}
	for bs := range stream {
		mem := Memory{}
		err := json.Unmarshal(bs, &mem)
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case e := <-errorC:
			return e
		default:
			handler(mem, cancelFunc)
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
