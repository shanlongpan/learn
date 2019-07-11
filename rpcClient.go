package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	pb "learn/proto"
	"log"
	"sync"
	"time"
)

var wg sync.WaitGroup

func main() {
	conn, err := grpc.Dial(":2333", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("dial error: %v\n", err)
	}
	defer conn.Close()

	worklist:=make(chan int)
	// 实例化 UserInfoService 微服务的客户端
	t := time.Now()
	for i:=0;i<2000;i++{
		go func() {
			for range worklist{
				wg.Done()

				client := pb.NewUserInfoServiceClient(conn)
				req := new(pb.UserRequest)

				req.Name = "小明"
				r, err:= client.GetUserInfo(context.Background(), req)
				if err != nil {
					log.Fatalf("resp error: %v\n", err)
				}
				if r.Name!="小明"{
					fmt.Println(r.Name)
				}
			}
		}()
	}

	for i := 0; i < 50000; i++ {
		wg.Add(1)
		worklist<-i
	}

	close(worklist)
	wg.Wait()

	// 调用服务
	fmt.Println(time.Since(t))
}
