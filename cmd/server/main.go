package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"strings"

	"daydream.theprimeagen.com/pkg/config"
)

type CLIServer struct {
	clis map[string]*CLIInterface
	id   int
}

func NewCLIServer() *CLIServer {
	slog.Info("initializing cli server")
	return &CLIServer{
		clis: map[string]*CLIInterface{},
		id:   0,
	}
}

func (c *CLIServer) HandleClient(conn net.Conn, id int) error {
	data := make([]byte, 1024)
	request := ""
	slog.Info("waiting for request from client connection", "id", id)
	msg, err := conn.Read(data)
	if err != nil {
		slog.Error("connection faield to read", "error", err, "id", id)
		conn.Close()
		return err
	}

	str := string(data[:msg])
	idx := strings.Index(str, "\n")
	slog.Info("received request", "id", id, "string", str, "index", idx)
	if idx == -1 {
		slog.Error("simple framing protocol failed: no newline on first message", "id", id, "string", str)
		return errors.New("simple framing protocol failed: no newline on first message")
	}

	request = str[:idx]

	if cli, ok := c.clis[request]; ok {
		go cli.AddConnection(conn)
	} else {
		cli, err := CreateNewOpenCodeSession()
		if err != nil {
			slog.Error("failed to create new open code session", "error", err, "id", id, "request", request)
			conn.Write([]byte("failed to create new open code session"))
			conn.Close()
			return err
		}
		c.clis[request] = cli
		go cli.Start()
		go cli.AddConnection(conn)
	}

	return nil
}

func (c *CLIServer) Start(ctx context.Context) error {
	slog.Info("starting cli server")
	listener, err := net.Listen("unix", config.SERVER_SOCKET)
	if err != nil {
		slog.Error("failed to listen on unix socket", "error", err)
		return err
	}

	for {
		conn, err := listener.Accept()
		id := c.id
		c.id++
		slog.Info("connection accepted", "error", err, "id", id)
		if err != nil {
			slog.Error("failed to accept connection", "error", err, "id", id)
			continue
		}

		slog.Info("new remote connection", "id", id)
		go c.HandleClient(conn, id)
	}
}

type CLIInterface struct {
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
	out    []io.ReadWriter
	cmdStr string
}

func NewCLIInterface(cmdStr string, args []string) (*CLIInterface, error) {
	cmd := exec.Command(cmdStr, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &CLIInterface{
		stdout: stdout,
		stderr: stderr,
		stdin:  stdin,
		out:    []io.ReadWriter{},
		cmdStr: cmdStr,
	}, nil
}

func (c *CLIInterface) Start() error {
	go func() {
		slog.Info("starting cli interface for", "command", c.cmdStr)
		data := make([]byte, 1024)
		for {
			msg, err := c.stdout.Read(data)
			if err != nil {
				slog.Error("failed to read message", "error", err)
				break
			}

			out := make([]byte, msg)
			copy(out, data[:msg])
			for _, conn := range c.out {
				conn.Write(out)
			}
		}
	}()
	return nil
}

func (c *CLIInterface) AddConnection(conn net.Conn) {
	c.out = append(c.out, conn)
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			slog.Error("failed to read from stdin", "error", err)
			return
		}
		if n == 0 {
			continue
		}
		c.stdin.Write(buf[:n])
	}
}

func CreateNewOpenCodeSession() (*CLIInterface, error) {
	slog.Info("creating new open code session")
	return NewCLIInterface("/home/theprimeagen/personal/daydream/long_running_process", []string{})
}

func main() {
	_ = os.Remove(config.SERVER_SOCKET)
	server := NewCLIServer()
	if err := server.Start(context.Background()); err != nil {
		slog.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}
