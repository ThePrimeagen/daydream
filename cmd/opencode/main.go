package main

import (
	"context"
	"io"
	"log/slog"
	"os"

	"daydream.theprimeagen.com/pkg/program"
	"golang.org/x/term"
)

type FnWriter struct {
	w func(b []byte) (int, error)
}

func (f *FnWriter) Write(b []byte) (int, error) {
	return f.w(b)
}

func asWriter(w func(b []byte) (int, error)) io.Writer {
	return &FnWriter{w: w}
}

func handleInput(prg *program.Program) {
    oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
    if err != nil {
        slog.Error("failed to make raw terminal", "error", err)
        return
    }
    defer term.Restore(int(os.Stdin.Fd()), oldState) // Restore terminal state on exit

    b := make([]byte, 1)

//	program.EchoOff(os.Stdin)

    for {
        n, err := os.Stdin.Read(b)
        if err != nil {
            slog.Error("failed to read from stdin", "error", err)
            return
        }
        if n == 0 {
            continue
        }

		prg.SendKey(string(b[:n]))
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

func main() {
	ctx := context.Background()
	debug, err := os.Create("/tmp/test.log")
	if err != nil {
		slog.Error("failed to create debug log", "error", err)
		os.Exit(1)
	}

	prg := program.NewProgram("opencode").WithWriter(asWriter(func(b []byte) (int, error) {
		os.Stdout.Write(b)
		debug.Write(b)
		return len(b), nil
	}))

	go func() {
		err = prg.Run(ctx)
		if err != nil {
			slog.Error("failed to run program", "error", err)
			os.Exit(1)
		}
		<-ctx.Done()
	}()

	go handleInput(prg)
	select {}
}
