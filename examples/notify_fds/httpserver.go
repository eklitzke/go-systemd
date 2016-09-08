// Copyright 2016 Uber Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/coreos/go-systemd/activation"
	"github.com/coreos/go-systemd/daemon"
)

const (
	pingPort = 8076
	pongPort = 8077
)

type pingHandler struct{}

func (h *pingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ping\n"))
}

type pongHandler struct{}

func (h *pongHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong\n"))
}

func toTCPListener(listener net.Listener) *net.TCPListener {
	tcpListener, ok := listener.(*net.TCPListener)
	if !ok {
		log.Fatalf("expected a TCP socket\n")
	}
	return tcpListener
}

func main() {
	listeners, err := activation.Listeners(true)
	if err != nil {
		log.Fatalf("failed to activate listener: %v\n", err)
	}

	if len(listeners) != 2 {
		log.Fatalf("Unexpected number of socket-activated listeners: %v\n", listeners)
	}

	var pingListener, pongListener *net.TCPListener
	for _, l := range listeners {
		port := l.Addr().(*net.TCPAddr).Port
		switch port {
		case pingPort:
			pingListener = toTCPListener(l)
		case pongPort:
			pongListener = toTCPListener(l)
		default:
			log.Fatalf("unexpected port: %d\n", port)
		}
	}
	if pingListener == nil {
		log.Fatalf("missing ping listener\n")
	}
	if pongListener == nil {
		log.Fatalf("missing pong listener\n")
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// This signal handler will return the listen socket back to systemd before
	// exiting.
	go func() {
		_ = <-c

		pingFile, err := pingListener.File()
		if err == nil {
			pongFile, err := pongListener.File()
			if err == nil {
				daemon.SdNotifyWithFds(true, "FDSTORE=1", pingFile, pongFile)
			}
		}
		os.Exit(0)
	}()

	go func() {
		var ph pingHandler
		srv := &http.Server{Handler: &ph}
		srv.Serve(pingListener)
	}()
	var ph pongHandler
	srv := &http.Server{Handler: &ph}
	srv.Serve(pongListener)
}
