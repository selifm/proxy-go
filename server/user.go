package main

import (
	"fmt"
	"net"
)

func main() {

	reconn := make(chan bool)
	defer func() {
		err := recover()
		if err != nil {
			fmt.Println(err)
		}
	}()
	for {
		userClient, err := net.Dial("tcp", "127.0.0.1:5200")
		if err != nil {
			panic(err)
		}

		var name string
		for {
			fmt.Scanln(&name)
			if name != "" {
				if _, err := userClient.Write([]byte(name)); err != nil {
					go func() {
						reconn <- true
					}()
					panic(err)
				}
				data := make([]byte, 10240)
				if n, err := userClient.Read(data); err != nil {
					go func() {
						reconn <- true
					}()
					panic(err)
				} else {
					fmt.Println(string(data[:n]))
				}
			}
		}
		<-reconn
	}

}
