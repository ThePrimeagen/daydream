package main

import (
	"context"
	"log/slog"
	"net"
	"os"

	"daydream.theprimeagen.com/pkg/config"
)

func handleServer(ctx context.Context, listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Error("failed to accept connection", "error", err)
			continue
		}

		slog.Info("new remote connection")
		go handleClient(ctx, conn)
	}
}

func handleClient(outer context.Context, conn net.Conn) {
	defer conn.Close()
	ctx, cancel := context.WithCancel(outer)

	chn := make(chan []byte, 2)
	go func() {
		buf := make([]byte, 1024)
		for {
			msg, err := conn.Read(buf)
			if err != nil {
				slog.Error("failed to read message", "error", err)
				break
			}

			out := make([]byte, msg)
			copy(out, buf[:msg])
			chn <- out
		}

		cancel()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-chn:
			slog.Info("sending message", "message", string(msg))
		default:
		}
	}
}

func main() {
	_ = os.Remove(config.SERVER_SOCKET)
	server, err := net.Listen("unix", config.SERVER_SOCKET)
	if err != nil {
		slog.Error("failed to listen on unix socket", "error", err)
		os.Exit(1)
	}

	/// TODO: i should listen for SIGINT and SIGTERM and cancel the context
	ctx, _ := context.WithCancel(context.Background())

	go handleServer(ctx, server)
	<-ctx.Done()
}

