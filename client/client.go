package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

var (
	host       string
	localPort  int
	remotePort int
)

func init() {
	flag.StringVar(&host, "h", "127.0.0.1", "remote server ip")
	flag.IntVar(&localPort, "l", 8080, "the local port")
	flag.IntVar(&remotePort, "r", 3333, "remote server port")
}

type server struct {
	conn net.Conn
	// 数据传输通道
	read  chan []byte
	write chan []byte
	// 异常退出通道
	exit chan error
	// 重新获取连接
	reconn chan bool
}

// 从Server端读取数据
func (s *server) Read() {
	for {
		if s.conn == nil {
			continue
		}
		// 如果10秒钟内没有消息传输，则Read函数会返回一个timeout的错误
		_ = s.conn.SetReadDeadline(time.Now().Add(time.Second * 10))
		data := make([]byte, 10240)
		n, err := s.conn.Read(data)
		if err != nil && err != io.EOF {
			// 读取超时，发送一个心跳包过去
			if strings.Contains(err.Error(), "timeout") {
				// 3秒发一次心跳
				_ = s.conn.SetReadDeadline(time.Now().Add(time.Second * 3))
				s.conn.Write([]byte("pi"))
				continue
			}
			fmt.Println("从server读取数据失败, ", err.Error())
			s.exit <- err
		}

		// 如果收到心跳包, 则跳过
		if data[0] == 'p' && data[1] == 'i' {
			fmt.Println("client收到心跳包")
			continue
		}
		s.read <- data[:n]
	}
}

// 将数据写入到Server端
func (s *server) Write() {
	for {
		if s.conn == nil {
			continue
		}
		select {
		case data := <-s.write:
			_, err := s.conn.Write(data)
			if err != nil && err != io.EOF {
				s.exit <- err
			}
		}
	}
}

type local struct {
	conn net.Conn
	// 数据传输通道
	read  chan []byte
	write chan []byte
	// 有异常退出通道
	exit chan error
	// 重新获取连接
	reconn chan bool
}

func (l *local) Read() {

	for {
		if l.conn == nil {
			continue
		}
		data := make([]byte, 10240)
		n, err := l.conn.Read(data)
		if err != nil {
			l.exit <- err
		}
		l.read <- data[:n]
	}
}

func (l *local) Write() {
	for {
		if l.conn == nil {
			continue
		}
		select {
		case data := <-l.write:
			_, err := l.conn.Write(data)
			if err != nil {
				l.exit <- err
			}
		}
	}
}

func main() {
	flag.Parse()
	//连接远程服务
	server := &server{
		conn:   nil,
		read:   make(chan []byte),
		write:  make(chan []byte),
		exit:   make(chan error),
		reconn: make(chan bool),
	}

	go func() {
		server.reconn <- true
	}()

	go getServerConn(server)
	go server.Read()
	go server.Write()

	//连接本地服务
	local := &local{
		conn:   nil,
		read:   make(chan []byte),
		write:  make(chan []byte),
		exit:   make(chan error),
		reconn: make(chan bool),
	}

	go func() {
		local.reconn <- true
	}()

	go getLocalConn(local)
	go local.Read()
	go local.Write()

	//远程服务和本地服务交互
	go handle(server, local)

	re := make(chan int64)
	<-re
}

func getServerConn(server *server) {
	for {
		v, ok := <-server.reconn
		if ok && v == true {
			serverConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, remotePort))
			if err == nil {
				server.conn = serverConn
				fmt.Printf("已连接server: %s \n", serverConn.RemoteAddr())
			} else {
				fmt.Printf("已连接server FAIL: %s \n", err)
				//连接失败，继续连接
				go func() {
					server.reconn <- true
				}()
			}
		}
	}
}

func getLocalConn(local *local) {
	for {
		v, ok := <-local.reconn
		if ok && v == true {
			localConn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", localPort))
			if err == nil {
				local.conn = localConn
				fmt.Printf("已连接本地server: %s \n", localConn.RemoteAddr())
			} else {
				fmt.Printf("已连接本地server FAIL: %s \n", err)
				//连接失败，继续连接
				go func() {
					local.reconn <- true
				}()
			}
		}

	}
}

func handle(server *server, local *local) {
	for {
		if server.conn == nil || local.conn == nil {
			time.Sleep(time.Duration(5) * time.Second)
			continue
		}
		select {
		case data := <-server.read:
			local.write <- data

		case data := <-local.read:
			server.write <- data

		case err := <-server.exit:
			if server.conn != nil {
				_ = server.conn.Close()
			}
			server.conn = nil
			server.reconn <- true
			fmt.Printf("server have err: %s", err.Error())
		case err := <-local.exit:
			if local.conn != nil {
				_ = local.conn.Close()
			}
			local.conn = nil
			local.reconn <- true
			fmt.Printf("local have err: %s", err.Error())
		}
	}
}
