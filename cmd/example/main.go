package main

import (
	"fmt"
	"os"
	"time"
)

func write_output() {
	for i := range 10000 {
		fmt.Printf("Hello, world!: %d\n", i)
		time.Sleep(time.Second)
	}
}

func read_input() {
	data := make([]byte, 1024)
	for {
		n, err := os.Stdin.Read(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
			return
		}
		if n == 0 {
			continue
		}
		fmt.Printf("Key pressed: %c (ASCII: %d)\n", data[0], data[0])
	}
}

func main() {
	go write_output()
	go read_input()

	select {}
}

