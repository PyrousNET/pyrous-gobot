package main

import (
	"context"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

type systemdNotifier struct {
	conn              *net.UnixConn
	watchdogInterval  time.Duration
	watchdogSupported bool
}

func newSystemdNotifier() *systemdNotifier {
	socket := os.Getenv("NOTIFY_SOCKET")
	if socket == "" {
		return nil
	}

	addr := &net.UnixAddr{
		Name: socket,
		Net:  "unixgram",
	}
	if socket[0] == '@' {
		addr.Name = "\x00" + socket[1:]
	}

	conn, err := net.DialUnix("unixgram", nil, addr)
	if err != nil {
		log.Printf("systemd notifier disabled: %v", err)
		return nil
	}

	interval, supported := watchdogIntervalFromEnv()

	return &systemdNotifier{
		conn:              conn,
		watchdogInterval:  interval,
		watchdogSupported: supported,
	}
}

func watchdogIntervalFromEnv() (time.Duration, bool) {
	usecStr := os.Getenv("WATCHDOG_USEC")
	if usecStr == "" {
		return 0, false
	}

	if pidStr := os.Getenv("WATCHDOG_PID"); pidStr != "" {
		pid, err := strconv.Atoi(pidStr)
		if err == nil && pid != os.Getpid() {
			return 0, false
		}
	}

	usec, err := strconv.ParseInt(usecStr, 10, 64)
	if err != nil || usec <= 0 {
		return 0, false
	}

	return time.Duration(usec) * time.Microsecond, true
}

func (n *systemdNotifier) notify(state string) {
	if n == nil {
		return
	}
	if _, err := n.conn.Write([]byte(state)); err != nil {
		log.Printf("systemd notify failed: %v", err)
	}
}

func (n *systemdNotifier) NotifyReady() {
	n.notify("READY=1")
}

func (n *systemdNotifier) NotifyWatchdog() {
	n.notify("WATCHDOG=1")
}

func (n *systemdNotifier) NotifyStopping() {
	n.notify("STOPPING=1")
}

func (n *systemdNotifier) StartWatchdog(ctx context.Context) {
	if n == nil || !n.watchdogSupported || n.watchdogInterval <= 0 {
		return
	}

	interval := n.watchdogInterval / 2
	if interval <= 0 {
		interval = n.watchdogInterval
	}

	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				n.NotifyWatchdog()
			}
		}
	}()
}

func (n *systemdNotifier) Close() {
	if n == nil {
		return
	}
	if err := n.conn.Close(); err != nil {
		log.Printf("systemd notifier close error: %v", err)
	}
}
