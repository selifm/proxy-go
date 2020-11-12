package main

import (
	"fmt"
	"runtime"
	"time"
)

var g1Num = 10
var g2Num = 10

func g1() {

	for {
		fmt.Println("111111111111111111111111111111111111")
		g1Num--
		if g1Num < 0 {
			runtime.Goexit()
		}
		time.Sleep(time.Duration(1) * time.Millisecond)
	}
}

func g2() {
	go g1()
	for {
		fmt.Println("222222222222")
		g2Num--
		if g2Num < 5 {
			runtime.Goexit()
		}
		time.Sleep(time.Duration(1) * time.Millisecond)
	}

}

func main() {
	go g2()

	time.Sleep(time.Duration(1) * time.Minute)
}
