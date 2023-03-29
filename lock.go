package cache

import (
	"time"

	"github.com/go-web-kits/cache/redislock"
	"github.com/pkg/errors"
)

var DistributedLock = Lock

func Lock(key string, maxTTL time.Duration, lambda func() error) error {
	locker := redislock.New(Cli)
	lock, err := locker.Obtain("__lock:"+key, maxTTL, nil)
	if err != nil {
		return errors.Wrap(err, "cache.Lock#ObtainKey")
	}

	err = lambda()
	if err != nil {
		return err
	}

	err = lock.Release()
	if err != nil {
		return errors.Wrap(err, "cache.Lock#ReleaseLock")
	}
	return nil
}

func GetLock(key string, ttl time.Duration) (*redislock.Lock, error) {
	locker := redislock.New(Cli)
	lock, err := locker.Obtain("__lock:"+key, ttl, nil)
	return lock, errors.Wrap(err, "cache.GetLock#ObtainKey")
}

func spinning(keys []string, opt Opt) error {
	if !opt.UnderLocking {
		return nil
	}

	k := []string{}
	for _, key := range keys {
		k = append(k, "__lock:"+key)
	}
	result, err := Cli.Exists(k...).Result()
	if err != nil {
		return err
	}
	if result == int64(0) { // no lock
		return nil
	}

	if opt.FailIfLocked {
		return errors.New("cache.spinning: under locking")
	}
	time.Sleep(100 * time.Millisecond)
	return spinning(keys, opt)
}
