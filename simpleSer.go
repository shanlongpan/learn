package main

import (
	"fmt"
	"net"
	"reflect"
	"time"
)

type Server struct {
	addr  string
	funcs map[string]reflect.Value
}

func NewServer(host string) *Server {
	return &Server{addr: host, funcs: make(map[string]reflect.Value)}
}

//注册函数反射  f为函数
func (s *Server) Register(rpcName string, f interface{}) {
	if _, ok := s.funcs[rpcName]; ok {
		return
	}
	// 1. 要通过反射来调用起对应的方法，必须要先通过reflect.ValueOf(interface)来获取到reflect.Value，得到“反射类型对象”后才能做下一步处理
	fVal := reflect.ValueOf(f)
	s.funcs[rpcName] = fVal
}

// 等待
func (s *Server) Run() {
	l, _ := net.Listen("tcp", s.addr)
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go handleConn(conn, s)
	}
}

func handleConn(conn net.Conn, s *Server) {
	conn.SetReadDeadline(time.Now().Add(2 * time.Minute)) // 2分钟超时时间，没有数据读入，断开连接

	defer conn.Close()

	srvSession := NewSession(conn)

	for {
		// 读取 RPC 调用数据
		b, err := srvSession.Read()

		if err != nil {
			fmt.Println(err)
			return
		}
		// 解码 RPC 调用数据
		rpcData, err := decode(b)

		if err != nil {
			fmt.Println(err)
			return
		}

		f, ok := s.funcs[rpcData.Name]

		if !ok {
			fmt.Printf("func %s not exists", rpcData.Name)
			return
		}

		// 构造函数的参数
		inArgs := make([]reflect.Value, 0, len(rpcData.Args))
		for _, arg := range rpcData.Args {
			inArgs = append(inArgs, reflect.ValueOf(arg))
		}

		// 执行调用
		out := f.Call(inArgs)
		outArgs := make([]interface{}, 0, len(out))
		for _, o := range out {
			outArgs = append(outArgs, o.Interface())
		}

		// 包装数据返回给客户端
		respRPCData := RPCData{rpcData.Name, outArgs}
		respBytes, _ := encode(respRPCData)
		srvSession.Write(respBytes)
	}

}

func main() {
	ser := NewServer("localhost:9999")

	//注册函数反射
	ser.Register("sum", sum)
	ser.Register("incr", incr)

	ser.Run()
}

func sum(n, m int) (int, error) {
	return n + m, nil
}

func incr(n int) (int, error) {
	return n + 1, nil
}

/**
  如下是自己测使用的测试代码
*/

func test() {
	funcs := make(map[string]reflect.Value) // server 端维护 funcName => func 的 map
	funcs["incr"] = reflect.ValueOf(incr)

	args := []reflect.Value{reflect.ValueOf(1)} // 构建参数（client 传递上来）

	vals := funcs["incr"].Call(args) // 调用执行

	var res []interface{}

	for _, val := range vals {
		//Interface returns v's current value as an interface{}
		res = append(res, val.Interface()) // 处理返回值
	}
	fmt.Println(res) // [2, <nil>]

	args_test2 := []reflect.Value{reflect.ValueOf(1), reflect.ValueOf(2)}

	vals = funcs["sum"].Call(args_test2)
	var res2 []interface{}
	for _, val := range vals {
		//Interface returns v's current value as an interface{}
		res2 = append(res2, val.Interface()) // 处理返回值
	}
	fmt.Println(res2) // [2, <nil>]
}
