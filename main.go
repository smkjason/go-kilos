package main

import (
	"bufio"
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

const (
	ioctlReadTermios  = unix.TIOCGETA
	ioctlWriteTermios = unix.TIOCSETA
)

var (
	stdinfd  = int(os.Stdin.Fd())
	stdoutfd = int(os.Stdout.Fd())
)

type key byte

type editor struct {
	reader *bufio.Reader

	originalTermios *unix.Termios
}

// Kills the program.
func die() {
	os.Exit(1)
}

func newEditor() *editor {
	return &editor{
		reader:          bufio.NewReader(os.Stdin),
		originalTermios: nil,
	}
}

// disableRawMode sets the Termios back to original.
func (e *editor) disableRawMode() {
	unix.IoctlSetTermios(stdinfd, uint(ioctlWriteTermios), e.originalTermios)
}

// enableRawMode enables rawMode.
func (e *editor) enableRawMode() (*unix.Termios, error) {
	t, err := unix.IoctlGetTermios(unix.Stdin, uint(ioctlReadTermios))
	if err != nil {
		return nil, err
	}
	e.originalTermios = t
	raw := *t

	raw.Iflag &^= unix.BRKINT | unix.ICRNL | unix.INPCK | unix.ISTRIP | unix.IXON
	raw.Lflag &^= unix.ECHO | unix.ICANON | unix.ISIG
	raw.Oflag &^= unix.OPOST
	raw.Cc[unix.VMIN] = 1
	raw.Cc[unix.VTIME] = 1
	if err := unix.IoctlSetTermios(stdinfd, uint(ioctlWriteTermios), &raw); err != nil {
		return nil, err
	}

	return t, nil
}

func (e *editor) readKey() (key, error) {
	buf := make([]byte, 4)
	for {
		nread, err := e.reader.Read(buf)
		if err != nil {
			fmt.Println("Error reading")
		}

		if nread > 0 {
			switch buf[0] {
			case ctrl('q'):
				fmt.Println("Goodbye.\r")
				die()

			default:
				return key(buf[0]), nil
			}
		}
	}
}

// ctrl returns a byte resulting from pressing the given ASCII character with the ctrl-key.
func ctrl(char byte) byte {
	return char & 0x1f
}

func main() {
	e := newEditor()
	_, err := e.enableRawMode()
	if err != nil {
		fmt.Println("Failed enabling Raw")
	}
	defer e.disableRawMode()

	for {
		k, err := e.readKey()
		if err != nil {
			fmt.Println("There was an error reading input")
		}

		fmt.Print(string(k))
	}
}
