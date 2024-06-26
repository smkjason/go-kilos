package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"

	"golang.org/x/sys/unix"
)

/* --- data --- */

type key int32

const (
	ioctlReadTermios  = unix.TIOCGETA
	ioctlWriteTermios = unix.TIOCSETA
	ioctlGetWin       = unix.TIOCGWINSZ

	welcomeMessage = "Welcome to Kilos - built in Golang"

	keyArrowLeft key = iota + 1000
	keyArrowRight
	keyArrowUp
	keyArrowDown
)

var (
	stdinfd  = int(os.Stdin.Fd())
	stdoutfd = int(os.Stdout.Fd())

	e *editor
)

type abuf struct {
}

type editor struct {
	winSizeRow int
	winSizeCol int

	cx int
	cy int

	reader *bufio.Reader

	termios *unix.Termios
}

func newEditor() *editor {
	return &editor{
		reader: bufio.NewReader(os.Stdin),
	}
}

/* --- terminal --- */

// die kills and exits the program.
func die(msg string) {
	os.Stdout.WriteString("\x1b[2J")
	os.Stderr.WriteString("\x1b[H")

	log.Fatal(msg + "\r")
	os.Exit(1)
}

// disableRawMode sets the Termios back to original.
func disableRawMode() {
	unix.IoctlSetTermios(stdinfd, uint(ioctlWriteTermios), e.termios)
}

// enableRawMode enables rawMode.
func enableRawMode() error {
	t, err := unix.IoctlGetTermios(unix.Stdin, uint(ioctlReadTermios))
	if err != nil {
		return err
	}
	e.termios = t

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

func getWindowSize() (int, int, error) {
	if w, err := unix.IoctlGetWinsize(unix.Stdin, ioctlGetWin); err == nil {
		return int(w.Row), int(w.Col), nil
	}

	// Fallback: Move the cursor to the bottom-right and read the position
	if _, err := os.Stdout.Write([]byte("\x1b[999C\x1b[999B")); err != nil {
		return 0, 0, err
	}

	r, c, err := getCursorPosition()
	if err != nil {
		return 0, 0, err
	}

	return r, c, nil
}

func getCursorPosition() (row, col int, err error) {
	if _, err = os.Stdout.Write([]byte("\x1b[6n")); err != nil {
		return
	}
	if _, err = fmt.Fscanf(os.Stdin, "\x1b[%d;%d", &row, &col); err != nil {
		return
	}
	return
}

/*  --- outputs --- */

func drawRows() {
	for i := 0; i < e.winSizeRow; i++ {
		if i == e.winSizeRow/2 {
			padding := (e.winSizeCol - len(welcomeMessage)) / 2
			for ; padding > 0; padding-- {
				os.Stdout.WriteString(" ")
			}
			os.Stdout.WriteString("Kilo editor -- version v1")
		} else {
			os.Stdout.Write([]byte("~"))
		}

		os.Stdout.Write([]byte("\x1b[K"))

		if i < e.winSizeRow-1 {
			os.Stdout.WriteString("\r\n")
		}
	}
}

func refreshScreen() {
	os.Stdout.Write([]byte("\x1b[?25l"))
	os.Stdout.Write([]byte("\x1b[H"))
	drawRows()

	// Position cursor.
	os.Stdout.Write([]byte(fmt.Sprintf("\x1b[%d;%dH", e.cy+10, e.cx+10)))
	os.Stdout.Write([]byte("\x1b[?25h"))
}

/*  --- inputs --- */

func moveCursor(k key) {
	switch k {
	case keyArrowUp:
		if e.cy > 0 {
			e.cy--
		}
	case keyArrowDown:
		e.cy++
	case keyArrowLeft:
		if e.cx > 0 {
			e.cx--
		}
	case keyArrowRight:
		e.cx++
	default:
	}
}

// ctrl returns a byte resulting from pressing the given ASCII character with the ctrl-key.
func ctrl(char byte) key {
	return key(char & 0x1f)
}

func readKey() (key, error) {
	buf := make([]byte, 4)
	for {
		nread, err := e.reader.Read(buf)
		if err != nil {
			log.Fatal("Error reading")
		}

		if nread > 0 {
			switch {
			case bytes.Equal(buf, []byte("\x1b[A")):
				return keyArrowUp, nil
			case bytes.Equal(buf, []byte("\x1b[B")):
				return keyArrowDown, nil
			case bytes.Equal(buf, []byte("\x1b[C")):
				return keyArrowRight, nil
			case bytes.Equal(buf, []byte("\x1b[D")):
				return keyArrowLeft, nil
			default:
				return key(buf[0]), nil
			}
		}
	}
}

func processKey() {
	k, err := readKey()
	if err != nil {
		log.Fatal("failed to read key")
	}

	switch k {
	case ctrl('q'):
		die("goodbye")
	case keyArrowDown, keyArrowLeft, keyArrowRight, keyArrowUp:
		moveCursor(k)
	default:
		os.Stdout.WriteString(string(k))
	}
}

/* --- main --- */

func initEditor() {
	e.cx = 0
	e.cy = 0

	winRow, winCol, err := getWindowSize()
	if err != nil {
		die("setWindowSize")
	}

	e.winSizeRow, e.winSizeCol = int(winRow), int(winCol)
}

func main() {
	e = newEditor()
	err := enableRawMode()
	if err != nil {
		log.Fatal("Failed enabling Raw")
	}
	defer disableRawMode()
	initEditor()

	for {
		refreshScreen()
		processKey()
	}
}
