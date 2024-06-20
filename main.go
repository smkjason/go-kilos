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

	originalTermios *unix.Termios
)

type key byte

type editor struct {
	reader *bufio.Reader
}

// Kills the program.
func die() {
	os.Exit(1)
}

func newEditor() *editor {
	return &editor{
		reader: bufio.NewReader(os.Stdin),
	}
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

// disableRawMode sets the Termios back to original.
func disableRawMode() {
	unix.IoctlSetTermios(stdinfd, uint(ioctlWriteTermios), originalTermios)
}

// enableRawMode enables rawMode.
func enableRawMode() error {
	t, err := unix.IoctlGetTermios(unix.Stdin, uint(ioctlReadTermios))
	if err != nil {
		return err
	}
	originalTermios = t

	raw := *t
	raw.Iflag &^= unix.BRKINT | unix.ICRNL | unix.INPCK | unix.ISTRIP | unix.IXON
	raw.Lflag &^= unix.ECHO | unix.ICANON | unix.ISIG
	raw.Oflag &^= unix.OPOST
	raw.Cc[unix.VMIN] = 1
	raw.Cc[unix.VTIME] = 1
	if err := unix.IoctlSetTermios(stdinfd, uint(ioctlWriteTermios), &raw); err != nil {
		return err
	}

	return nil
}

// ctrl returns a byte resulting from pressing the given ASCII character with the ctrl-key.
func ctrl(char byte) byte {
	return char & 0x1f
}

func main() {
	e := newEditor()
	err := enableRawMode()
	if err != nil {
		fmt.Println("Failed enabling Raw")
	}
	defer disableRawMode()

	for {
		k, err := e.readKey()
		if err != nil {
			fmt.Println("There was an error reading input")
		}

		fmt.Print(string(k))
	}
}
