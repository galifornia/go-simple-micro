package main

import (
	"fmt"
	"time"
)

func main() {
	for {
		fmt.Println("Logger service")
		time.Sleep(8 * time.Second)
	}
}
