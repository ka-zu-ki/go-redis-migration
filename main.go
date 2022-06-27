package main

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
)

// 最大同時並列実行数
const concurrency = 600

func main() {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	ctx := context.Background()

	if _, err := client.Ping(ctx).Result(); err != nil {
		fmt.Println(err)
	}

	fmt.Println("redis connection started")

	// セマフォとしてのゴルーチン
	sem := make(chan struct{}, concurrency)
	defer close(sem)

	var wg sync.WaitGroup

	for i := 0; i < 6100805; i++ {
		// 空のstructを送信
		sem <- struct{}{}

		wg.Add(1)

		i := i
		go func() {
			defer wg.Done()
			// 処理が終わったらチャネルを開放
			defer func() { <-sem }()

			client.SetNX(ctx, "vh:8ZS9JS9KzEESoR:189-30_s1_p"+strconv.Itoa(i), []byte{}, time.Second*10)
			fmt.Println(i)
		}()
	}

	wg.Wait()

	fmt.Println("redis connection stopped")
}
