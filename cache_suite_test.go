package cache_test

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/go-web-kits/cache"
	_ "github.com/go-web-kits/cache/test"
	. "github.com/go-web-kits/testx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCache(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cache Suite")
}

var _ = BeforeSuite(func() {
	client := redis.NewClient(&redis.Options{
		Network:  "tcp",
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       1,  // use default DB
		// Dial timeout for establishing new connections.
		DialTimeout: time.Duration(10000) * time.Millisecond,
		// Timeout for socket reads. If reached, commands will fail
		// with a timeout instead of blocking. Use value -1 for no timeout and 0 for default.
		ReadTimeout: time.Duration(1000) * time.Millisecond,
		// Timeout for socket writes. If reached, commands will fail  with a timeout instead of blocking.
		WriteTimeout: time.Duration(1000) * time.Millisecond,
		PoolSize:     10,
		// Maximum number of socket connections.
		// Default is 10 connections per every CPU as reported by runtime.NumCPU.
		// PoolSize int
	})

	pong, err := client.Ping().Result()
	fmt.Println("    Pong: ", pong, err)

	// cache.Init(client, nil)
	cache.Init(client, log.New(os.Stdout, "\r\n", 0))
})

var _ = AfterSuite(func() {
	ShutApp()
})
