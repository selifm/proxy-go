package main

import (
	"fmt"
	"net"
)

func main() {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Println(err)
		}
	}()
	reconn := make(chan error)
	for {
		userClient, err := net.Dial("tcp", "127.0.0.1:5200")
		if err != nil {
			go func() {
				reconn <- err
			}()
		}

		var name string
		for {
			fmt.Println("请输入你要传输的数据:")
			fmt.Scanln(&name)
			if name != "" {
				if _, err := userClient.Write([]byte(name)); err != nil {
					go func() {
						reconn <- err
					}()
					break
				}
				data := make([]byte, 10240)
				if n, err := userClient.Read(data); err != nil {
					go func() {
						reconn <- err
					}()
					break
				} else {
					fmt.Println(string(data[:n]))
				}
			}
		}
		<-reconn
	}

}
