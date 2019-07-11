package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

// 定义客户端请求的数据格式
type Test struct {
	Name  string `json:"name"`
	ID    int    `json:"id"`
	Age   int    `json:"age"`
	Title []string  `json:"title"`
}

func main() {
	t := time.Now()
	host := "http://127.0.0.1:8091/ping"
	tr := &http.Transport{
		MaxIdleConns:       200,
		MaxIdleConnsPerHost:  200,
	}
	netClient:= &http.Client{Transport: tr}
	// 压测会报错
	// read: connection reset by peer

	// SYN队列和ACCEPT队列的长度是有限制的，SYN队列长度由内核参数tcp_max_syn_backlog决定，
	// ACCEPT队列长度可以在调用listen(backlog)通过backlog，但总最大值受到内核参数somaxconn(/proc/sys/net/core/somaxconn)限制。
	// 若SYN队列满了，新的SYN包会被直接丢弃。若ACCEPT队列满了，建立成功的连接不会从SYN队列中移除，同时也不会拒绝新的连接，这会加剧SYN队列的增长，
	// 最终会导致SYN队列的溢出。当ACCEPT队列溢出之后，只要打开tcp_abort_on_flow内核参数(默认为0，关闭)，建立连接后直接回RST，
	// 拒绝连接(可以通过/proc/net/netstat中ListenOverflows和ListenDrops查看拒绝的数目)。

	//$ sysctl -a | grep somaxconn
	//kern.ipc.somaxconn: 2048
	//$ sudo sysctl -w kern.ipc.somaxconn=2048

	worklist := make(chan int)
	var wg sync.WaitGroup
	for i := 0; i < 2000; i++ {
		go func() {
			for range worklist {
				wg.Done()
				u := Test{Name: "xiao"}
				b := new(bytes.Buffer)
				err := json.NewEncoder(b).Encode(u)
				if err != nil {
					fmt.Println(err)
				}

				res, err := netClient.Post(host, "application/json; charset=utf-8", b)

				if err != nil {
					fmt.Println(err)
				}
				body, err := ioutil.ReadAll(res.Body)
				//fmt.Println(string(body))
				if err != nil {
					fmt.Println("请求1错误", err)
					return
				}
				var rtn Test
				err = json.Unmarshal(body, &rtn)

				if err != nil {
					fmt.Println("请求错误", err)
					return
				} else {
					if rtn.Name != "xiao" {
						fmt.Println(rtn)
					}
				}
				err = res.Body.Close()
				if err != nil {
					fmt.Println("请求错误", err)
					return
				}
			}
		}()
	}

	for i := 0; i < 50000; i++ {
		wg.Add(1)
		worklist <- i
	}

	close(worklist)
	wg.Wait()
	// 调用服务
	fmt.Println(time.Since(t))
}
