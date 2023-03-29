package cache

import (
	l "log"
	"os"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/go-web-kits/dbx"
	"github.com/pkg/errors"
)

var Cli *redis.Client
var Logger interface{ Println(args ...interface{}) } = l.New(os.Stdout, "\r\n", 0)
var UnLog = false

type Opt struct {
	UnderLocking bool
	FailIfLocked bool
	ExpiresIn    time.Duration
	Default      interface{} // could be `func() interface{}` or other value type
	Force        bool
	To           interface{} // Unmarshal to the object (pointer)
	// ZeroValue interface{}
}

func Init(cli *redis.Client, logger interface{ Println(args ...interface{}) }) {
	Cli = cli
	Cli.AddHook(Hook{})
	if logger != nil {
		Logger = logger
	}
}

func Get(key string, opts ...Opt) (interface{}, error) {
	var value string
	var err error
	opt := optGet(opts)

	err = spinning([]string{key}, opt)
	if err != nil {
		return nil, errors.Wrap(err, "cache.Get")
	}

	value, err = Cli.Get(key).Result()
	if err != nil {
		return nil, errors.Wrap(err, "cache.Get")
	}

	decoded, err := decode(value)
	if err != nil {
		// returns the un-decoded value
		return value, nil
		// return nil, errors.Wrap(err, "redis.Cli.Get")
	}

	if len(opts) > 0 && opts[0].To != nil {
		return UnCompress(decoded, opts[0].To)
	}
	return UnCompress(decoded)
}

func Set(key string, value interface{}, opts ...Opt) error {
	opt := optGet(opts)
	compressed, err := Compress(value)
	if err != nil {
		return errors.Wrap(err, "cache.Set")
	}

	err = spinning([]string{key}, opt)
	if err != nil {
		return errors.Wrap(err, "cache.Set")
	}

	err = Cli.Set(key, encode(compressed), opt.ExpiresIn).Err()
	if err == nil && opt.To != nil {
		_, err = UnCompress(compressed, opt.To)
	}
	return errors.Wrap(err, "cache.Set")
}

// val, err := Fetch("key")                    => Key Not Found is not error
// val, err := Fetch("key", Opt{Force: true})  => Key Not Found is error
// val, err := Fetch("key", Opt{Default: ...}) => Err when the Set() call fails
func Fetch(key string, opts ...Opt) (val interface{}, err error) {
	val, err = Get(key, opts...)
	if err == nil {
		return val, nil // Cache Matched
	}

	if len(opts) == 0 || filtered(err) != nil {
		return val, filtered(err) // Key Not Found is not error
	}

	opt := opts[0]
	if opt.Force {
		return val, errors.Wrap(err, "cache.Fetch") // Key Not Found is error
	}

	if opt.Default != nil {

		if f, ok := opt.Default.(func() interface{}); ok {
			val = f()
			if e, ok := val.(error); ok {
				return nil, e
			}
			if result, ok := val.(dbx.Result); ok {
				if result.Err != nil {
					return nil, result.Err
				}
				val = result.Data
			}
		} else {
			val = opt.Default
		}

		return val, Set(key, val, opt)
	}

	return val, err // Key Not Found is not error
}

func Delete(keys ...string) error {
	return Cli.Del(keys...).Err()
}

func DeleteUnderSpinLock(keys ...string) error {
	err := spinning(keys, Opt{UnderLocking: true})
	if err != nil {
		return errors.Wrap(err, "cache.DeleteUnderSpinLock")
	}
	return Delete(keys...)
}

func DeleteMatched(pattern string, opts ...Opt) error {
	opt := optGet(opts)
	keys, err := Cli.Keys(pattern).Result()
	if err != nil {
		return errors.Wrap(err, "cache.DeleteMatched")
	}
	if len(keys) == 0 {
		return nil
	}

	err = spinning(keys, opt)
	if err != nil {
		return errors.Wrap(err, "cache.DeleteMatched")
	}
	return Delete(keys...)
}
