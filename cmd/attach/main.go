package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	// Open input_pipe for writing
	inputPipe, err := os.OpenFile("input_pipe", os.O_WRONLY, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening input_pipe: %v\n", err)
		os.Exit(1)
	}
	defer inputPipe.Close()

	// Open output_pipe for reading
	outputPipe, err := os.OpenFile("output_pipe", os.O_RDONLY, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening output_pipe: %v\n", err)
		os.Exit(1)
	}
	defer outputPipe.Close()

	// Start a goroutine to read from output_pipe
	go func() {
		scanner := bufio.NewScanner(outputPipe)
		for scanner.Scan() {
			fmt.Println("Received:", scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading output_pipe: %v\n", err)
		}
	}()

	// Read user input from terminal and write to input_pipe
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter messages (Ctrl+D or Ctrl+C to quit):")
	for scanner.Scan() {
		message := scanner.Text()
		_, err := fmt.Fprintln(inputPipe, message)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to input_pipe: %v\n", err)
			os.Exit(1)
		}
		inputPipe.Sync() // Flush immediately
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
}

