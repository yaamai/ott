package main

import (
	//        "io"
	"log"
	"os"
	"os/exec"
	//        "os/signal"
	//        "syscall"

	"github.com/creack/pty"
	//        "golang.org/x/crypto/ssh/terminal"
)

type Runner struct {
	ptmx *os.File
}

func test() error {
	// Create arbitrary command.
	c := exec.Command("sh")

	// Start the command with a pty.
	ptmx, err := pty.StartWithSize(c, &pty.Winsize{Rows: 10, Cols: 10})
	if err != nil {
		return err
	}
	// Make sure to close the pty at the end.
	defer func() { _ = ptmx.Close() }() // Best effort.

	// Handle pty size.
	// ch := make(chan os.Signal, 1)
	// signal.Notify(ch, syscall.SIGWINCH)
	// go func() {
	//         for range ch {
	//                 if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
	//                         log.Printf("error resizing pty: %s", err)
	//                 }
	//         }
	// }()
	// ch <- syscall.SIGWINCH // Initial resize.

	// Set stdin in raw mode.
	// oldState, err := terminal.MakeRaw(int(os.Stdin.Fd()))
	// if err != nil {
	//         panic(err)
	// }
	// defer func() { _ = terminal.Restore(int(os.Stdin.Fd()), oldState) }() // Best effort.

	// Copy stdin to the pty and the pty to stdout.
	//go func() { _, _ = io.Copy(ptmx, os.Stdin) }()
	b := make([]byte, 1024)
	ptmx.Read(b)
	log.Println("=> ", string(b))

	ptmx.Write([]byte("\r\necho aaaa\r\n"))
	ptmx.Write([]byte("\r\nsleep 10 && exit\r\n"))
	//ptmx.Write([]byte("exit\n"))
	// _, _ = io.Copy(os.Stdout, ptmx)
	ptmx.Read(b)
	log.Println(string(b))

	//ptmx.Read(b)
	//log.Println(string(b))
	func() {
		for {
			b := make([]byte, 1024)
			ptmx.Read(b)
			log.Println("=> ", string(b))
		}
	}()

	return nil
}

func main() {
	if err := test(); err != nil {
		log.Fatal(err)
	}
}
