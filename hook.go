package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v7"
)

type Hook struct{}

func (h Hook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	return context.WithValue(ctx, "start", time.Now()), nil
}

func (h Hook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	start := ctx.Value("start").(time.Time)
	duration := time.Since(start)
	vals := []string{}
	for _, v := range cmd.Args()[1:] {
		vals = append(vals, fmt.Sprint(v))
	}

	var content string
	switch cmd.Name() {
	case "get", "mget", "del", "exists":
		content = strings.Join(vals, " ")
	case "set", "setnx", "incrby", "decrby":
		content = cmd.Args()[1].(string) + ":: " + strings.Join(vals[1:], ", ")
	default:
		content = strings.Join(vals, " ")
	}

	log(strings.ToUpper(cmd.Name()), content, duration)
	return nil
}

func (h Hook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	return ctx, nil
}

func (h Hook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	return nil
}
