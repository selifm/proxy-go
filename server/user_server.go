package main

import (
	"fmt"
	"io"
	"net"
	"time"
)

type myUser struct {
	conn net.Conn
	// 数据传输通道
	read  chan []byte
	write chan []byte
	// 异常退出通道
	exit chan error
}

// 从User端读取数据
func (u *myUser) Read() {
	_ = u.conn.SetReadDeadline(time.Now().Add(time.Second * 200))
	for {
		data := make([]byte, 10240)
		n, err := u.conn.Read(data)
		if err != nil && err != io.EOF {
			u.exit <- err
		}
		u.read <- data[:n]
	}
}

// 将数据写给User端
func (u *myUser) Write(repData []byte) {
	for {
		select {
		case data := <-u.read:
			var str string = fmt.Sprintf("收到你的信息：%s,我给你返回响应%s", data, time.Now().String())
			var repData []byte = []byte(str)
			_, err := u.conn.Write(repData)
			if err != nil && err != io.EOF {
				u.exit <- err
			}
		}
	}
}

func main() {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Println(err)
		}
	}()

	userClient, err := net.Dial("tcp", "127.0.0.1:5200")
	if err != nil {
		panic(err)
	}

	var name string
	for {
		fmt.Scanln(&name)
		if name != "" {
			userClient.Write([]byte(name))
			data := make([]byte, 10240)
			n, _ := userClient.Read(data)
			fmt.Println(string(data[:n]))
			continue
		}
	}

}
