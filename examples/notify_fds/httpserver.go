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

// +build ignore

package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/coreos/go-systemd/activation"
	"github.com/coreos/go-systemd/daemon"
)

func helloServer(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "welcome to the brave new systemd future\n")
}

func main() {
	listeners, err := activation.Listeners(true)
	if err != nil {
		log.Fatalf("failed to activate listener: %v\n", err)
	}

	if len(listeners) != 1 {
		panic("Unexpected number of socket activation fds")
	}
	listenSock := listeners[0]

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

	http.HandleFunc("/", helloServer)
	http.Serve(listenSock, nil)
}
