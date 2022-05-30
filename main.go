package main

import (
	"fmt"
	"time"
)

func main() {
	result := Scan(10*time.Second, true)
	if len(result) == 0 {
		fmt.Println("no HDLC framed serial ports found")
	}
	for _, path := range result {
		fmt.Println(path)
	}
}
