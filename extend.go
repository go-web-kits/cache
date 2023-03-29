package cache

import (
	"github.com/go-redis/redis/v7"
	"github.com/pkg/errors"
)

func Increase(key string, value ...int) error {
	var err error
	by := 1
	if len(value) > 0 {
		by = value[0]
	}
	_, err = Cli.IncrBy(key, int64(by)).Result()
	if err == redis.Nil {
		err = Cli.Set(key, by, 0).Err()
		if err != nil {
			return errors.Wrap(err, "redis.Cli.Increase#InitSet")
		}
	} else if err != nil {
		return errors.Wrap(err, "redis.Cli.Increase#IncrBy")
	}
	return nil
}

func IncreaseUnderSpinLock(key string, value ...int) error {
	err := spinning([]string{key}, Opt{UnderLocking: true})
	if err != nil {
		return errors.Wrap(err, "cache.IncreaseUnderSpinLock")
	}
	return Increase(key, value...)
}

func Decrease(key string, value ...int) error {
	var err error
	by := 1
	if len(value) > 0 {
		by = value[0]
	}
	_, err = Cli.DecrBy(key, int64(by)).Result()
	if err == redis.Nil {
		err = Cli.Set(key, -by, 0).Err()
		if err != nil {
			return errors.Wrap(err, "redis.Cli.Decrease#InitSet")
		}
	} else if err != nil {
		return errors.Wrap(err, "redis.Cli.Decrease#DecrBy")
	}
	return nil
}

func DecreaseUnderSpinLock(key string, value ...int) error {
	err := spinning([]string{key}, Opt{UnderLocking: true})
	if err != nil {
		return errors.Wrap(err, "cache.DecreaseUnderSpinLock")
	}
	return Decrease(key, value...)
}
