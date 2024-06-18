package main

import (
	"bufio"
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

var (
	stdinfd  = int(os.Stdin.Fd())
	stdoutfd = int(os.Stdout.Fd())

	ioctlReadTermios  = unix.TIOCGETA
	ioctlWriteTermios = unix.TIOCSETA
)

func enableRawMode() (*unix.Termios, error) {
	t, err := unix.IoctlGetTermios(stdinfd, uint(ioctlReadTermios))
	if err != nil {
		return nil, err
	}
	raw := *t // make a copy to avoid mutating the original
	raw.Lflag &^= unix.ECHO
	if err := unix.IoctlSetTermios(stdinfd, uint(ioctlWriteTermios), &raw); err != nil {
		return nil, err
	}
	return t, nil
}

func main() {
	_, err := enableRawMode()
	if err != nil {
		fmt.Println("Failed enabling Raw")
	}
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() && scanner.Text() != "q" {
		text := scanner.Text()
		fmt.Println("You entered: " + text)
	}
}
