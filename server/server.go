package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"runtime"
	"strings"
	"time"
)

var (
	localPort  int
	remotePort int
)

func init() {
	flag.IntVar(&localPort, "l", 5200, "the user link port")
	flag.IntVar(&remotePort, "r", 3333, "client listen port")
}

type client struct {
	conn net.Conn
	// 数据传输通道
	read  chan []byte
	write chan []byte
	// 异常退出通道
	exit chan error
}

// 从Client端读取数据
func (c *client) Read() {
	for {
		if c.conn == nil {
			continue
		}
		// 如果10秒钟内没有消息传输，则Read函数会返回一个timeout的错误
		_ = c.conn.SetReadDeadline(time.Now().Add(time.Second * 10))
		data := make([]byte, 10240)
		n, err := c.conn.Read(data)
		if err != nil && err != io.EOF {
			if strings.Contains(err.Error(), "timeout") {
				// 设置读取时间为3秒，3秒后若读取不到, 则err会抛出timeout,然后发送心跳
				_ = c.conn.SetReadDeadline(time.Now().Add(time.Second * 3))
				c.conn.Write([]byte("pi"))
				continue
			}
			fmt.Println("读取出现错误...")
			c.exit <- err
		}

		// 收到心跳包,则跳过
		if data[0] == 'p' && data[1] == 'i' {
			fmt.Println("server收到心跳包")
			continue
		}
		c.read <- data[:n]
	}
}

// 将数据写入到Client端
func (c *client) Write() {
	for {
		if c.conn == nil {
			continue
		}
		select {
		case data := <-c.write:
			_, err := c.conn.Write(data)
			if err != nil && err != io.EOF {
				c.exit <- err
			}
		}
	}
}

type user struct {
	conn net.Conn
	// 数据传输通道
	read  chan []byte
	write chan []byte
	// 异常退出通道
	exit chan error
}

// 重连通道
var userReConn chan bool

// 从User端读取数据
func (u *user) Read() {
	for {
		if u.conn == nil {
			break
		}
		_ = u.conn.SetReadDeadline(time.Now().Add(time.Second * 200))
		data := make([]byte, 10240)
		n, err := u.conn.Read(data)
		if err != nil && err != io.EOF {
			u.exit <- err
		}
		u.read <- data[:n]

	}
}

// 将数据写给User端
func (u *user) Write() {
	for {
		if u.conn == nil {
			break
		}
		select {
		case data := <-u.write:
			_, err := u.conn.Write(data)
			if err != nil && err != io.EOF {
				u.exit <- err
			}
		}
	}
}

func main() {
	flag.Parse()

	defer func() {
		err := recover()
		if err != nil {
			fmt.Println(err)
		}
	}()

	clientListener, err := net.Listen("tcp", fmt.Sprintf(":%d", remotePort))
	if err != nil {
		panic(err)
	}
	fmt.Printf("监听:%d端口,服务已开启... \n", remotePort)
	// 监听User来连接
	userListener, err := net.Listen("tcp", fmt.Sprintf(":%d", localPort))
	if err != nil {
		panic(err)
	}
	fmt.Printf("监听:%d端口, 服务已开启... \n", localPort)

	client := &client{
		conn:  nil,
		read:  make(chan []byte),
		write: make(chan []byte),
		exit:  make(chan error),
	}
	//接收被透传的client请求
	go client.Read()
	go client.Write()
	//接收客户端请求
	fmt.Println("等待client连接..")
	go AcceptClientConn(clientListener, client)
	fmt.Println("等待user连接..")
	//接收用户请求
	userConnChan := make(chan net.Conn)
	go AcceptUserConn(userListener, userConnChan) //由用户连接后，将用户连接存在userConnChan
	go HandleClient(client, userConnChan)

	re := make(chan int64)
	<-re
}

func HandleClient(client *client, userConnChan chan net.Conn) {
	for {
		if client.conn == nil {
			continue
		}
		select {
		case userConn := <-userConnChan:
			user := &user{
				conn:  userConn,
				read:  make(chan []byte),
				write: make(chan []byte),
				exit:  make(chan error),
			}
			go user.Read()
			go user.Write()
			go handle(client, user)
		}
	}
}

// 将两个Socket通道链接
// 1. 将从user收到的信息发给client
// 2. 将从client收到信息发给user
func handle(client *client, user *user) {
	for {
		if user.conn == nil {
			break
		}
		if client.conn == nil {
			fmt.Println("client失去连接，等待连接")
			user.write <- []byte("client 失去连接，5s后重试")
			time.Sleep(time.Duration(5) * time.Second)
			continue
		}
		select {
		case userRecv := <-user.read:
			// 收到从user发来的信息
			client.write <- userRecv
		case clientRecv := <-client.read:
			// 收到从client发来的信息
			user.write <- clientRecv

		case err := <-client.exit:
			fmt.Println("client出现错误, 关闭连接", err.Error())
			if client.conn != nil {
				_ = client.conn.Close()
				client.conn = nil
			}
		case err := <-user.exit:
			fmt.Println("user出现错误，关闭连接", err.Error())
			if user.conn != nil {
				_ = user.conn.Close()
				user.conn = nil
			}
			runtime.Goexit()
		}
	}
}

// 等待user连接
func AcceptClientConn(clientListener net.Listener, client *client) {
	for {
		if clientListener == nil {
			continue
		}
		// 接收client 连接
		clientConn, err := clientListener.Accept()
		if err != nil {
			panic(err)
		}
		client.conn = clientConn
		fmt.Printf("有Client连接: %s \n", clientConn.RemoteAddr())
	}

}

// 等待user连接
func AcceptUserConn(userListener net.Listener, connChan chan net.Conn) {
	for {
		if userListener == nil {
			continue
		}
		userConn, err := userListener.Accept()
		if err != nil {
			panic(err)
		}
		fmt.Printf("user connect: %s \n", userConn.RemoteAddr())
		connChan <- userConn
	}

}
