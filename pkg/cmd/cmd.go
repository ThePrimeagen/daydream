package cmd

import (
	"io"
	"os/exec"
)


type Command struct {
	cmd *exec.Cmd
	stdin io.Writer
	stdout io.Reader
	stderr io.Reader
}

//func NewCommand() (*Command, error) {
//	wd, err := os.Getwd()
//	if err != nil {
//		slog.Error("failed to get working directory, WTF", "error", err)
//		return nil, err
//	}
//	return &Command{
//		Name: wd,
//	}, nil
//}

type reader struct {
	cb func([]byte)
}

func (r *reader) Read(p []byte) (n int, err error) {
	r.cb(p)
	return len(p), nil
}

func toReader(cb func(data []byte)) io.Reader {
	return &reader{
		cb: cb,
	}
}

type writer struct {
	cb func([]byte)
}

func (w *writer) Write(p []byte) (n int, err error) {
	w.cb(p)
	return len(p), nil
}

func toWriter(cb func(data []byte)) io.Writer {
	return &writer{
		cb: cb,
	}
}

func NewCommand() (*Command, error) {
	cmd := exec.Command("/home/theprimeagen/personal/daydream/long_running_program")
	command := &Command{
		cmd: cmd,
	}

	cmd.Stdin = toReader(func(data []byte) {
		command.stdin.Write(data)
	})

	cmd.Stdout = toWriter(func(data []byte) {
		command.stdout.Write(data)
	})
	cmd.Stderr = toWriter(func(data []byte) {
		command.stderr.Write(data)
	})

	return command, nil
}

func (c *Command) Start() error {
	err := c.cmd.Start()
	if err != nil {
		return err
	}
	return nil
}

func (c *Command) Connect(stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	//todo: what about buffering of stdout?
	c.stdin = stdin
	c.stdout = stdout
	c.stderr = stderr
}
