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

// +build ignore

package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"

	"github.com/coreos/go-systemd/activation"
)

func HelloServer(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "hello socket activated world!\n")
}

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Term)

	go func() {
		s := <-c
		notify.SdNotifyWithFds()

	}()

	// Block until a signal is received.
	fmt.Println("Got signal:", s)

	listeners, err := activation.Listeners(true)
	if err != nil {
		panic(err)
	}

	if len(listeners) != 1 {
		panic("Unexpected number of socket activation fds")
	}

	http.HandleFunc("/", HelloServer)
	http.Serve(listeners[0], nil)
}
