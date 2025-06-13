package main

import (
	"log/slog"
	"net"
	"os"

	"daydream.theprimeagen.com/pkg/config"
	"golang.org/x/term"
)

func readInput(conn net.Conn) {
    oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
    if err != nil {
        slog.Error("failed to make raw terminal", "error", err)
        return
    }
    defer term.Restore(int(os.Stdin.Fd()), oldState) // Restore terminal state on exit

    b := make([]byte, 1)

    for {
        n, err := os.Stdin.Read(b)
        if err != nil {
            slog.Error("failed to read from stdin", "error", err)
            return
        }
        if n == 0 {
            continue
        }

		conn.Write(b[:n])

		char := b[0]
		if char == 3 {
			slog.Info("ctrl+c detected, exiting...")
			os.Exit(0)
		}

		if char == 4 {
			slog.Info("ctrl+d detected, exiting...")
			os.Exit(0)
		}
    }
}

func writeOutput(conn net.Conn) {
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			slog.Error("failed to read from server", "error", err)
			os.Exit(0)
		}
		os.Stdout.Write(buf[:n])
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

