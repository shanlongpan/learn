package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	pb "learn/proto"
	"log"
	"net"
)

// 定义服务端实现约定的接口
type UserInfoService struct{}

var u = UserInfoService{}

// 实现 interface
func (s *UserInfoService) GetUserInfo(ctx context.Context, req *pb.UserRequest) (resp *pb.UserResponse, err error) {
	name := req.Name

	// 模拟在数据库中查找用户信息
	// ...
	if name == "小明" {
		resp = &pb.UserResponse{
			Id:    233,
			Name:  name,
			Age:   20,
			Title: []string{"Gopher", "PHPer"}, // repeated 字段是 slice 类型
		}
	}else{
		resp=&pb.UserResponse{
			Id:    111,
			Name:  "小李",
			Age:   30,
			Title: []string{"厨师"}, // repeated 字段是 slice 类型
		}
	}
	err = nil
	return
}

func main() {
	port := ":2333"
	l, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("listen error: %v\n", err)
	}
	fmt.Printf("listen %s\n", port)
	s := grpc.NewServer()

	// 将 UserInfoService 注册到 gRPC
	// 注意第二个参数 UserInfoServiceServer 是接口类型的变量
	// 需要取地址传参
	pb.RegisterUserInfoServiceServer(s, &u)
	s.Serve(l)
}
