package main

import (
	"fmt"
	"io"
	"net"
	"time"
)

type myLocalServer struct {
	conn net.Conn
	// 数据传输通道
	read  chan []byte
	write chan []byte
	// 异常退出通道
	exit chan error
}

// 从User端读取数据
func (u *myLocalServer) Read() {
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
func (u *myLocalServer) Write() {
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

	clientListener, err := net.Listen("tcp", fmt.Sprintf(":%d", 8080))
	if err != nil {
		panic(err)
	}

	for {
		fmt.Printf("监听:%d端口, 作为需要透传的服务... \n", 8080)
		clientConn, err := clientListener.Accept()
		if err != nil {
			panic(err)
		}
		my := &myLocalServer{
			conn:  clientConn,
			read:  make(chan []byte),
			write: make(chan []byte),
			exit:  make(chan error),
		}
		go my.Read()
		go my.Write()

		<-my.exit
	}

}
