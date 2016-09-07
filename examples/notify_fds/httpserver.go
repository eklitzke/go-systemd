// Copyright 2015 CoreOS, Inc.
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
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/coreos/go-systemd/activation"
	"github.com/coreos/go-systemd/daemon"
)

func HelloServer(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "hello socket activated world!\n")
}

func main() {
	listeners, err := activation.Listeners(true)
	if err != nil {
		panic(err)
	}

	if len(listeners) != 1 {
		panic("Unexpected number of socket activation fds")
	}
	listenSock := listeners[0]

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		_ = <-c

		// Make a best effort to store the listen socket in systemd
		tcpListener, ok := listenSock.(*net.TCPListener)
		if ok {
			listenFile, err := tcpListener.File()
			if err != nil {
				sent, err := daemon.SdNotifyWithFds(true, "FDSTORE=1", listenFile)
				fmt.Printf("sent = %v, err = %v\n", sent, err)
			}
		}
		os.Exit(0)
	}()

	http.HandleFunc("/", HelloServer)
	http.Serve(listenSock, nil)
}
