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

type pingHandler struct{}

func (h *pingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ping\n"))
}

type pongHandler struct{}

func (h *pongHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ping\n"))
}

func main() {
	listeners, err := activation.ListenersWithNames(true)
	if err != nil {
		log.Fatalf("failed to activate listener: %v\n", err)
	}

	if len(listeners) != 2 {
		log.Fatalf("Unexpected number of socket-activated listeners: %v\n", listeners)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// This signal handler will return the listen socket back to systemd before
	// exiting.
	go func() {
		_ = <-c

		// Make a best effort to store the listen socket in systemd, ignoring any
		// errors.
		tcpListener, ok := listenSock.(*net.TCPListener)
		if ok {
			listenFile, err := tcpListener.File()
			if err != nil {
				daemon.SdNotifyWithFds(true, "FDSTORE=1", listenFile)
			}
		}
		os.Exit(0)
	}()

	var listenMap map[string]net.Listener
	for nl := range listeners {
		listenMap[nl.Name] = nl.Listener
	}
	pingListener, ok := listenMap["ping"]
	if !ok {
		log.Fatalf("expected to get 'ping' socket")

	}
	pongListener, ok := listenMap["pong"]
	if !ok {
		log.Fatalf("expected to get 'pong' socket")

	}
	go func() {
		srv := &http.Server{Handler: pingHandler}
		srv.Serve(pingListener, nil)
	}()
	srv := &http.Server{Handler: pongHandler}
	srv.Serve(pongListener, nil)
}
