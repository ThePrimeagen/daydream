package program

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"time"

	"daydream.theprimeagen.com/pkg/assert"
	"github.com/creack/pty"
	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

type Program struct {
	*os.File
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	path   string
	rows   int
	cols   int
	writer io.Writer
	args   []string
}

func NewProgram(path string) *Program {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		slog.Error("failed to get terminal size", "error", err)
		os.Exit(1)
	}

	return &Program{
		path:   path,
		rows:   height,
		cols:   width,
		writer: nil,
		cmd:    nil,
		File:   nil,
	}
}

func (a *Program) SendKey(key string) {
    for _, k := range key {
        a.Write([]byte{byte(k)})
        <-time.After(time.Millisecond * 40)
    }
}

func (a *Program) WithArgs(args []string) *Program {
	a.args = args
	return a
}

func (a *Program) WithWriter(writer io.Writer) *Program {
	if a.writer != nil {
		a.writer = io.MultiWriter(a.writer, writer)
	} else {
		a.writer = writer
	}
	return a
}

func setRawMode(f *os.File) error {
    fd := int(f.Fd())
    const ioctlReadTermios = unix.TCGETS  // Linux
    const ioctlWriteTermios = unix.TCSETS // Linux

    termios, err := unix.IoctlGetTermios(fd, ioctlReadTermios)
    if err != nil {
        return err
    }

    // Set raw mode but preserve output processing for ANSI colors
    termios.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON
    // Keep OPOST enabled to preserve ANSI color processing
    termios.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN
    termios.Cflag &^= unix.CSIZE | unix.PARENB
    termios.Cflag |= unix.CS8

    // Set VMIN and VTIME for non-blocking reads
    termios.Cc[unix.VMIN] = 1
    termios.Cc[unix.VTIME] = 0

    return unix.IoctlSetTermios(fd, ioctlWriteTermios, termios)
}

func EchoOff(f *os.File) {
	fd := int(f.Fd())
	//      const ioctlReadTermios = unix.TIOCGETA // OSX.
	const ioctlReadTermios = unix.TCGETS // Linux
	//      const ioctlWriterTermios =  unix.TIOCSETA // OSX.
	const ioctlWriteTermios = unix.TCSETS // Linux

	termios, err := unix.IoctlGetTermios(fd, ioctlReadTermios)
	if err != nil {
		panic(err)
	}

	newState := *termios
	newState.Lflag &^= unix.ECHO
	newState.Lflag |= unix.ICANON | unix.ISIG
	newState.Iflag |= unix.ICRNL
	if err := unix.IoctlSetTermios(fd, ioctlWriteTermios, &newState); err != nil {
		panic(err)
	}
}

func (a *Program) Run(ctx context.Context) error {
	assert.Assert(a.writer != nil, "you must provide a writer before you call run")
	assert.Assert(a.File == nil, "you have already started the program")

	cmd := exec.Command(a.path, a.args...)

	// Set proper terminal environment for color support
	cmd.Env = append(os.Environ(),
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
	)

	a.cmd = cmd

	ptmx, err := pty.Start(cmd)

	err = pty.Setsize(ptmx, &pty.Winsize{
        Rows: uint16(a.rows),
        Cols: uint16(a.cols),
    })

	if err != nil {
		return err
	}

	if err := setRawMode(ptmx); err != nil {
		ptmx.Close()
		return err
	}

	a.File = ptmx

	_, err = io.Copy(a.writer, ptmx)
	return err
}

func (a *Program) Close() error {
	err := a.File.Close()
	a.File = nil
	return err
}


