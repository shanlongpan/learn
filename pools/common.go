package pools

import (
	"encoding/binary"
	"io"
	"net"
)

func Write(s net.Conn,data []byte) error {
	buf := make([]byte, 4+len(data))                       // 4 字节头部 + 数据长度
	binary.BigEndian.PutUint32(buf[:4], uint32(len(data))) // 写入头部
	copy(buf[4:], data)                                    // 写入数据
	_, err := s.Write(buf)
	if err != nil {
		return err
	}
	return nil
}

// 从连接中读数据
func Read(s net.Conn)([]byte, error) {
	header := make([]byte, 4)
	_, err := io.ReadFull(s, header)
	if err != nil {
		return nil, err
	}
	dataLen := binary.BigEndian.Uint32(header)
	data := make([]byte, dataLen)
	_, err = io.ReadFull(s, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
