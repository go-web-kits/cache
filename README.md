# cache

Cache DSL for Automatic Serialization & Deserialization  
一套缓存 DSL，能够自动地序列化 & 反序列化，支持分布式锁，受 ActiveSupport::Cache 启发

Maintainers: @will.huang

## Features

1. 简洁清晰，封装了常用的缓存操作逻辑
2. 自动序列化存储 & 反序列化读取，轻松应对各种类型值的缓存存储
3. 支持分布式锁以及自旋锁（阻塞监听）
4. 能够打印缓存操作以及响应时间日志，帮你分析缓存问题

## Setup

```go
// init a redis client
client := redis.NewClient(&redis.Options{
    // ...
})

// init this lib
cache.Init(client, log.New(os.Stdout, "\r\n", 0))
```

## Usage

### Set

```go
cache.Set("key1", 1.2)
cache.Set("key2", "string")
cache.Set("key3", struct{ Abc string }{Abc: "abc"})
cache.Set("key4", map[string]interface{}{"hello": 1})
cache.Set("key5", []string{"hello", "world"})

// ExpiresIn
cache.Set("key1", 1.2, cache.Opt{ExpiresIn: 3 * time.Second})
```

### Get

```go
cache.Get("key1") // => 1.2
cache.Get("key2") // => "string"

// 未给定反序列化标的时的默认行为：
//     1. 所有数字类型会被反序列化成 `float64`
//     2. struct & map => map[string]interface{}
//     3. slice => []interface{}
cache.Get("key3") // => map[string]interface{}{"Abc": "abc"}
cache.Get("key4") // => map[string]interface{}{"hello": float64(1)}
cache.Get("key5") // => []interface{}{"hello", "world"}

// 给定反序列化标的：
var val3 struct{ Abc string }
cache.Get("key3", cache.Opt{To: &val3})
var val4 struct{ Hello int `json:"hello"` }
cache.Get("key4", cache.Opt{To: &val4})
var val5 []string
cache.Get("key5", cache.Opt{To: &val5})

// key 不存在时会报错
cache.Get("notexist") // => error
```

注：获取到非序列化的结果，返回 string 类型的该值

### Fetch

功能：首先调用 `Get`，若 key 不存在，则将给定默认值（或者 lambda 的执行结果）进行 `Set`。

```go
// result == true
result, err = cache.Fetch("k1", cache.Opt{Default: true})
// result == true
result, err = cache.Fetch("k1", cache.Opt{Default: false})

// default: lambda
//     1. 如果 lambda return error，则 `Set` 不会被执行，该 error 作为 `Fetch` 结果返回
//     2. 如果 lambda return 带有 error 的 dbx.Result，同上
//     3. 如果 lambda return dbx.Result，最终被 `Set` 以及返回的是 result.Data
//     4. 整个过程中发生 key not found 之外的 error 将同样被返回
result, err = cache.Fetch("k2", cache.Opt{
	Default: func() interface{} {
		if something {
			return "hello"
		} else {
			return errors.New("")
		}
	},
})

// 给定反序列化标的
var v3 struct{ Hello int `json:"hello"` }
_, err = cache.Fetch("k3", cache.Opt{
	Default: map[string]interface{}{"hello": 1},
	To: &v3,
})

result, err = cache.Fetch("k5", cache.Opt{Force: true}) // error!
result, err = cache.Fetch("k5") // result == err == nil
```

### Delete & DeleteMatched

```go
cache.Delete("key1", "key2")
cache.DeleteMatched("key*")
```

### Increase & Decrease

默认 step 为 1

```go
// 如果 key 不存在，会被初始化为 0 + step
cache.Increase("key1")     // 1
cache.Increase("key2", 10) // 10
cache.Increase("key1", 4)  // 5

// 如果 key 不存在，会被初始化为 0 - step
cache.Decrease("key1")     // -1
cache.Decrease("key2", 10) // -10
cache.Decrease("key1", 4)  // -5
```

## DistributedLock

分布式锁作用域为单个 key。

在一个 Lock 中执行语句（执行完成后自动释放锁）：
```go
cache.Lock("my-key", 1 * time.Second, func() error {
	// do something
})
// `1 * time.Second` means maxTTL
```

同时，如果有使用相同 key 的其他操作，给定 `UnderLocking` 选项。
注意此时默认行为为监听阻塞，直到锁过期或被释放（自旋锁，SpinLock）：
```go
cache.Get("my-key", cache.Opt{UnderLocking: true})
cache.Set("my-key", 123, cache.Opt{UnderLocking: true})
// 以下三个功能有特定方法来表示 UnderLocking
cache.DeleteUnderSpinLock("my-key")
cache.IncreaseUnderSpinLock("my-key")
cache.DecreaseUnderSpinLock("my-key")
```

如果希望遇锁直接返回 error，可以给定 `FailIfLocked` 选项：
```go
cache.Get("my-key", cache.Opt{UnderLocking: true, FailIfLocked: true})
```

## How It Works

1. 序列化和反序列化  
    参考 `ActiveSupport::Cache`，写入 cache 前序列化为 string，输出时进行反序列化  
    [source code](entry.go) & [test](test/entry.go)

2. 分布式锁：很简单，Exist key-name 则为锁定状态

