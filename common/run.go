package common

import (
	"context"
	"os/exec"
)

func RunOneShot(ctx context.Context, c string, arg []string) error {
	res := exec.CommandContext(ctx, c, arg...)
	err := res.Run()
	return err
}
