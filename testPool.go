package main

import (
	"fmt"
	"github.com/shanlongpan/learn/pools"
	"log"
	"net"
	"sync"
	"time"
)

func main() {
	pool := &pools.Pool{
		MaxIdle:         100,
		MaxActive:       2000,
		IdleTimeout:     20 * time.Second,
		MaxConnLifetime: 100 * time.Second,
		Wait:            true,
		Dial: func() (net.Conn, error) {
			//c, err := redis.Dial("tcp", "127.0.0.1:6379")
			c, err := net.Dial("tcp", "127.0.0.1:8972")
			if err != nil {
				return nil, err
			}
			return c, err
		},
	}
	defer pool.Close()

	t := time.Now()

	worklist := make(chan int)
	var wg sync.WaitGroup
	for i := 0; i < 2000; i++ {
		go func() {
			for range worklist {
				wg.Done()
				cli, err := pool.Get()
				if err != nil {
					log.Println(err)
					return
				}

				str := "test"

				err = pools.Write(cli.C, []byte(str))

				if err != nil {
					log.Println(err)
					pool.Put(cli, true)
					return
				}
				receByte, err := pools.Read(cli.C)
				if err != nil {
					log.Println(err)
				} else {
					if string(receByte) != "TEST" {
						fmt.Println("未返回预期的数据")
					}
				}
				pool.Put(cli, false)
			}
		}()
	}

	for i := 0; i < 500000; i++ {
		wg.Add(1)
		worklist <- i
	}

	fmt.Println("pool建立，连接数：", pool.Active)

	close(worklist)
	wg.Wait()
	// 调用服务
	fmt.Println(time.Since(t))
}
