// 演示 oa.98ent.com/p9/common/modules/cache：GetC 命中/回源、批量 MGet。
//
//	go run ./modules/cache/example
package main

import (
	"fmt"
	"sync/atomic"

	"oa.98ent.com/p9/common/modules/cache"
)

func main() {
	var loads int32
	loadUser := func(params map[string]any) (any, error) {
		n := atomic.AddInt32(&loads, 1)
		id := params["id"].(string)
		fmt.Println("callback load user", id, "times=", n)
		return map[string]any{"id": id, "name": "alice"}, nil
	}

	params := map[string]any{"id": "u1"}
	v1, err := cache.TwoMinuteCache.GetC("user:u1", loadUser, params, false)
	must(err)
	v2, err := cache.TwoMinuteCache.GetC("user:u1", loadUser, params, false)
	must(err)
	fmt.Println("== GetC 第一次回源、第二次命中")
	fmt.Println("first:", v1)
	fmt.Println("second:", v2)
	fmt.Println("loads:", atomic.LoadInt32(&loads))

	_, err = cache.TwoMinuteCache.GetC("user:u1", loadUser, params, true)
	must(err)
	fmt.Println("== noCache=true 强制回源")
	fmt.Println("loads:", atomic.LoadInt32(&loads))

	raw, ok := cache.TwoMinuteCache.GetCache("user:u1")
	fmt.Println("== GetCache", ok, raw)

	var batchLoads int32
	batch, err := cache.TwoMinuteCache.MGetCKeyString(
		[]string{"en-US", "zh-CN"},
		func(keys []string, _ map[string]any) (map[string]any, error) {
			n := atomic.AddInt32(&batchLoads, 1)
			fmt.Println("callback batch keys", keys, "times=", n)
			out := make(map[string]any, len(keys))
			for _, k := range keys {
				out[k] = "dict:" + k
			}
			return out, nil
		},
		nil,
		"i18n:",
		false,
	)
	must(err)
	fmt.Println("== MGetCKeyString")
	fmt.Println("batch:", batch)
	fmt.Println("batchLoads:", atomic.LoadInt32(&batchLoads))

	batch2, err := cache.TwoMinuteCache.MGetCKeyString(
		[]string{"en-US", "zh-CN"},
		func(keys []string, _ map[string]any) (map[string]any, error) {
			atomic.AddInt32(&batchLoads, 1)
			fmt.Println("callback should not run", keys)
			return nil, nil
		},
		nil,
		"i18n:",
		false,
	)
	must(err)
	fmt.Println("second batch (cache hit):", batch2)
	fmt.Println("batchLoads:", atomic.LoadInt32(&batchLoads))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
