package cache

import (
	"github.com/go-redis/redis/v7"
	"github.com/go-web-kits/utils/logx"
	"github.com/pkg/errors"
)

func IsKeyNotFound(err error) bool {
	return errors.Cause(err) == redis.Nil
}

func filtered(err error) error {
	if IsKeyNotFound(err) {
		return nil
	} else {
		return err
	}
}

func log(op string, val string, arg interface{}) {
	if UnLog {
		return
	}

	op = logx.Blod(logx.Magenta(op))
	if val != "" {
		val = " `" + val + "`"
	}

	if Logger != nil {
		logx.LogBy(Logger, "Redis", op+val, arg)
	} else {
		logx.Log("Redis", op+val, arg)
	}
}

func optGet(opts []Opt) Opt {
	var opt Opt
	if len(opts) > 0 {
		opt = opts[0]
	}
	return opt
}
