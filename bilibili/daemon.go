package gbl

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	rdb "runtime/debug"
	"strings"
	"syscall"
)

// Daemon for background task
func Daemon() {
	if !(version == "" || strings.Contains(version, "dev")) {
		if os.Getppid() != 1 {
			cmd := exec.Command(os.Args[0], os.Args[1:]...)
			fmt.Println(cmd)
			cmd.Start()
			fmt.Printf("%s [PID] %d running...\n", os.Args[0], cmd.Process.Pid)
			os.Exit(0)
		}
	}

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
