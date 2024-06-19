package main

import (
	"bufio"
	"fmt"
	"os"
	"unicode"

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

type editor struct {
	// Original Termios.
	termios *unix.Termios
}

func (e *editor) disableRawMode() {
	unix.IoctlSetTermios(stdinfd, uint(ioctlWriteTermios), e.termios)
}

func (e *editor) enableRawMode() (*unix.Termios, error) {
	t, err := unix.IoctlGetTermios(unix.Stdin, uint(ioctlReadTermios))
	if err != nil {
		return nil, err
	}
	e.termios = t

	raw := *t // make a copy to avoid mutating the original
	raw.Iflag &^= unix.BRKINT | unix.ICRNL | unix.INPCK | unix.ISTRIP | unix.IXON
	raw.Lflag &^= unix.ECHO | unix.ICANON | unix.ISIG
	raw.Oflag &^= unix.OPOST
	// raw.Cc[unix.VMIN] = 0
	// raw.Cc[unix.VTIME] = 1
	if err := unix.IoctlSetTermios(stdinfd, uint(ioctlWriteTermios), &raw); err != nil {
		return nil, err
	}

	return t, nil
}

func main() {
	e := editor{}
	_, err := e.enableRawMode()
	if err != nil {
		fmt.Println("Failed enabling Raw")
	}
	defer e.disableRawMode()

	scanner := bufio.NewReader(os.Stdin)
	for {
		r, _, err := scanner.ReadRune()
		if err != nil {
			fmt.Println("There was an error reading...")
			break
		}

		if string(r) == "q" {
			break
		}

		if unicode.IsControl(r) {
			fmt.Println("This is a control")
		} else {
			fmt.Println("You entered: " + string(r) + "\r")
		}
	}
}
