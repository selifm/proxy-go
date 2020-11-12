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
