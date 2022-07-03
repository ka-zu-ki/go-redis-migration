package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/go-redis/redis/v9"
)

// 最大同時並列実行数
const concurrency = 1200000

func main() {
	buf, err := ioutil.ReadFile("ssh key file path")
	if err != nil {
		panic(err)
	}
	key, err := ssh.ParsePrivateKeyWithPassphrase(buf, []byte("password"))
	if err != nil {
		panic(err)
	}

	sshConfig := &ssh.ClientConfig{
		User: "ssh key username",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}

	sshClient, err := ssh.Dial("tcp", "vm ip:22", sshConfig)
	if err != nil {
		panic(err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: net.JoinHostPort("memorystore ip", "6379"),
		Dialer: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return sshClient.Dial(network, addr)
		},
		ReadTimeout:  -1,
		WriteTimeout: -1,
	})

	ctx := context.Background()

	err = client.Ping(ctx).Err()
	if nil != err {
		log.Println(err)
	}

	fmt.Println("redis connection started")

	// データリセット
	status := client.FlushAll(ctx)
	fmt.Println(status)

	// セマフォとしてのゴルーチン
	sem := make(chan struct{}, concurrency)
	defer close(sem)

	var wg sync.WaitGroup

	for i := 0; i < 2; i++ {
		for j := 0; j < 6100805; j++ {
			// 空のstructを送信
			sem <- struct{}{}

			wg.Add(1)

			j := j
			go func() {
				defer wg.Done()
				// 処理が終わったらチャネルを開放
				defer func() { <-sem }()

				if i < 1 {
					// 500MB分のキーをセット　TTL60分
					err := client.SetNX(ctx, "vh:8ZS9JS9KzEESoR:189-30_s1_p"+strconv.Itoa(i)+strconv.Itoa(j), []byte{}, time.Minute*60).Err()
					if err != nil {
						fmt.Println(err)
					}
				} else {
					// 500MB分のキーをセット　TTL10秒
					err := client.SetNX(ctx, "vh:8ZS9JS9KzEESoR:189-30_s1_p"+strconv.Itoa(i)+strconv.Itoa(j), []byte{}, time.Second*10).Err()
					if err != nil {
						fmt.Println(err)
					}
				}
			}()
		}
	}

	wg.Wait()

	fmt.Println("redis connection stopped")

	defer client.Close()
	defer sshClient.Close()
}
