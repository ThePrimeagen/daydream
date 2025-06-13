package main

import (
	"log/slog"
	"net"
	"os"

	"daydream.theprimeagen.com/pkg/config"
	"golang.org/x/term"
)

func readInput(conn net.Conn) {
	// Store the original terminal state to restore later
    oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
    if err != nil {
        slog.Error("failed to make raw terminal", "error", err)
        return
    }
    defer term.Restore(int(os.Stdin.Fd()), oldState) // Restore terminal state on exit

    slog.Info("starting to read input")

    // Buffer to read one byte at a time
    b := make([]byte, 1)

    for {
        // Read a single byte from stdin
        n, err := os.Stdin.Read(b)
        if err != nil {
            slog.Error("failed to read from stdin", "error", err)
            return
        }
        if n == 0 {
            continue
        }

        // Get the character
		conn.Write(b[:n])

		char := b[0]
		if char == 3 { // ASCII for Ctrl+C
			slog.Info("ctrl+c detected, exiting...")
			break
		}
    }
}

func writeOutput(conn net.Conn) {
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		slog.Info("read from server", "error", err, "message", string(buf[:n]))
		if err != nil {
			slog.Error("failed to read from server", "error", err)
			os.Exit(1)
		}
		log := string(buf[:n])
		slog.Info("received from server", "message", log)
	}
}

func main() {
	conn, err := net.Dial("unix", config.SERVER_SOCKET)
	if err != nil {
		slog.Error("failed to connect to server", "error", err)
		os.Exit(1)
	}

	defer conn.Close()
	wd, err := os.Getwd()
	if err != nil {
		slog.Error("failed to get working directory", "error", err)
		os.Exit(1)
	}

	slog.Info("connected to server", "working directory", wd)
	_, err = conn.Write([]byte(wd + "\n"))
	if err != nil {
		slog.Error("failed to write to server", "error", err)
		os.Exit(1)
	}

	go readInput(conn)
	writeOutput(conn)
}

