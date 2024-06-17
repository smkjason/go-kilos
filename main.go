package main

import (
	"bufio"
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

var (
	stdinFd = int(os.Stdin.Fd())
)

func enableRawMode() (*unix.Termios, error) {
	t, err := unix.IoctlGetTermios(stdinFd, unix.TIOCGETA)
	if err != nil {
		fmt.Println("Failed to get Termios")
	}

	raw := *t
	raw.Cflag &^= unix.ECHO

	err = unix.IoctlSetTermios(stdinFd, unix.TIOCSETA, &raw)
	if err != nil {
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
