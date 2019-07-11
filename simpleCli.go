package main

import (
	"fmt"
	"net"
	"reflect"
	"log"
)

type Client struct {
	conn net.Conn
}
// fPtr 指向函数原型
func (c *Client) callRPC(rpcName string, fnAdder interface{}) {
	fn := reflect.ValueOf(fnAdder).Elem()

	// 完成与 Server 的交互
	f := func(args []reflect.Value) []reflect.Value {
		// 处理输入参数
		inArgs := make([]interface{}, 0, len(args))
		for _, arg := range args {
			inArgs = append(inArgs, arg.Interface())
		}

		// 编码 RPC 数据并请求
		cliSession := NewSession(c.conn)

		reqRPC := RPCData{Name: rpcName, Args: inArgs}
		b, err := encode(reqRPC)

		if err!=nil{
			fmt.Println(err)
			return nil
		}

		err=cliSession.Write(b)

		if err!=nil{
			fmt.Println(err)
			return nil
		}
		// 解码响应数据，得到返回参数
		respBytes, err := cliSession.Read()

		if err!=nil{
			fmt.Println(err)
			return nil
		}

		respRPC, err := decode(respBytes)

		if err!=nil{
			fmt.Println(err)
			return nil
		}
		//忽略函数的名字，只是读取函数的返回值
		outArgs := make([]reflect.Value, 0, len(respRPC.Args))
		for i, arg := range respRPC.Args {
			// 必须进行 nil 转换
			if arg == nil {
				outArgs = append(outArgs, reflect.Zero(fn.Type().Out(i)))
				continue
			}
			outArgs = append(outArgs, reflect.ValueOf(arg))
		}
		return outArgs
	}

	v := reflect.MakeFunc(fn.Type(), f)

	//func (v Value) Set(x Value)
	//将 x 赋值给 v 。
	fn.Set(v)
}

func main() {

	conn, err := net.Dial("tcp", "localhost:9999")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ClientTest:=Client{conn:conn}

	var sum func(int, int) (int,error)       //自己定义一个加法计算器 定义参数和返回值

	ClientTest.callRPC("sum",&sum)  //把自定义的加法函数绑定到rpc里面

	fmt.Println(sum(1,2))

	//var incr func(int) (int,error)       //自己定义一个加法计算器 定义参数和返回值
	//
	//ClientTest.callRPC("incr",&incr)  //把自定义的加法函数绑定到rpc里面
	//
	//fmt.Println(incr(1))
}


/**
	以下是测试函数
 */
func test2(){

	swap := func(args []reflect.Value) []reflect.Value {
		return []reflect.Value{args[1], args[0]}
	}

	var intSwap func(int, int) (int, int)
	fn := reflect.ValueOf(&intSwap).Elem()  // 获取 intSwap 未初始化的函数原型
	v := reflect.MakeFunc(fn.Type(), swap)  // MakeFunc 使用传入的函数原型创建一个绑定 swap 的新函数
	fn.Set(v)                               // 为函数 intSwap 赋值

	fmt.Println(intSwap(1, 2)) // 2 1
}