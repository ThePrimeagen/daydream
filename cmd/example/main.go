package main

import (
	"bufio"
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
	scanner := bufio.NewScanner(os.Stdin)
	for {
		if !scanner.Scan() {
			break
		}
		fmt.Println(scanner.Text())
	}

	fmt.Println("STDIN closed")
}

func main() {
	go write_output()
	go read_input()

	select {}
}

