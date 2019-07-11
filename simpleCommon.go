package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"io"
	"net"
)

//通用cli和server都可以使用
type Session struct {
	conn net.Conn
}

func NewSession(c net.Conn) *Session {
	return &Session{conn: c}
}

// 向连接中写数据 为了解决“粘包”问题，即客户端发送的多个数据包被当做一个数据包接收。也称数据的无边界性，read()
// 函数不知道数据包的开始或结束标志（实际上也没有任何开始或结束标志），只把它们当做连续的数据流来处理。
// 我们用四个字节定义一次发包数据的大小

func (s *Session) Write(data []byte) error {
	buf := make([]byte, 4+len(data))                       // 4 字节头部 + 数据长度
	binary.BigEndian.PutUint32(buf[:4], uint32(len(data))) // 写入头部
	copy(buf[4:], data)                                    // 写入数据
	_, err := s.conn.Write(buf)
	if err != nil {
		return err
	}
	return nil
}

// 从连接中读数据
func (s *Session) Read() ([]byte, error) {
	header := make([]byte, 4)
	_, err := io.ReadFull(s.conn, header)
	if err != nil {
		return nil, err
	}
	dataLen := binary.BigEndian.Uint32(header)
	data := make([]byte, dataLen)
	_, err = io.ReadFull(s.conn, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

type RPCData struct {
	Name string
	Args []interface{}
}

//gob 编码
func encode(data RPCData) ([]byte, error) {
	var buf bytes.Buffer
	bufEnc := gob.NewEncoder(&buf)
	if err := bufEnc.Encode(data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

//gob 解码
func decode(b []byte) (RPCData, error) {
	buf := bytes.NewBuffer(b)
	bufDec := gob.NewDecoder(buf)
	var data RPCData
	if err := bufDec.Decode(&data); err != nil {
		return data, err
	}
	return data, nil
}
