package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"daydream.theprimeagen.com/pkg/program"
	"golang.org/x/term"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	debug, err := os.Create("/tmp/test.log")
	if err != nil {
		slog.Error("failed to create debug log", "error", err)
		os.Exit(1)
	}
	defer debug.Close()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	prg := program.NewProgram("opencode").WithWriter(program.FnToWriter(func(b []byte) (int, error) {
		os.Stdout.Write(b)
		debug.Write(b)
		return len(b), nil
	}))

	// Store terminal state for restoration
	var oldState *term.State
	if term.IsTerminal(int(os.Stdin.Fd())) {
		oldState, err = term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			slog.Error("failed to make raw terminal", "error", err)
		}
	}

	// Ensure terminal is restored on exit
	defer func() {
		if oldState != nil {
			term.Restore(int(os.Stdin.Fd()), oldState)
		}
	}()

	go func() {
		err = prg.Run(ctx)
		if err != nil {
			slog.Error("failed to run program", "error", err)
		}
		cancel()
	}()

	go prg.PassThroughInput(os.Stdin)

	// Wait for signal or context cancellation
	select {
	case <-sigChan:
		cancel()
	case <-ctx.Done():
	}
}
