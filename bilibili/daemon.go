package gbl

import (
	"os"
	"os/signal"
	rdb "runtime/debug"
	"syscall"
)

// Daemon for background task
func Daemon() {
	// if os.Getppid() != 1 {
	// 	args := append([]string{os.Args[0]}, os.Args[1:]...)
	// 	os.StartProcess(os.Args[0], args, &os.ProcAttr{Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}})
	// 	return
	// }

	go func() {
		signals := make(chan os.Signal, 1)
		defer close(signals)

		signal.Notify(signals, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGQUIT)

		for {
			sig := <-signals
			infoln("signal:", sig.String(), "trace:", string(rdb.Stack()))
			os.Exit(1)
		}
	}()

	loop()
}
