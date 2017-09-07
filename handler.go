// +build linux darwin

package main

import (
	"os"
	"os/signal"
	rdb "runtime/debug"
	"syscall"
)

func handlerInterrupt() {
	go func() {
		signals := make(chan os.Signal)
		defer close(signals)

		signal.Notify(signals, syscall.SIGBUS, syscall.SIGFPE, syscall.SIGSEGV)
		signal.Notify(signals, syscall.SIGHUP, syscall.SIGABRT, syscall.SIGTERM)
		signal.Notify(signals, syscall.SIGQUIT, syscall.SIGILL, syscall.SIGTRAP)

		for {
			select {
			case sig := <-signals:
				logger.Println("signal:", sig.String(), "trace:", string(rdb.Stack()))
			}
		}
	}()
}
