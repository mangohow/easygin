//go:build linux || darwin
// +build linux darwin

package easygin

import (
	"os"
	"os/signal"
	"syscall"
)

func SetupSignal(fn func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGUSR1, syscall.SIGUSR2)
	
	var stopper Stopper
	select {
	case sig := <-c:
		switch sig {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			fn()
		case syscall.SIGUSR1:
			if stopper == nil {
				stopper = StartProfile()
			} else {
				stopper.Stop()
				stopper = nil
			}
		case syscall.SIGUSR2:
			dumpGoroutines()
		}
	}
}
