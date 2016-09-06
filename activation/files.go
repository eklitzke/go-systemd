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

// Package activation implements primitives for systemd socket activation.
package activation

import (
	"os"
	"strconv"
	"strings"
	"syscall"
)

// based on: https://gist.github.com/alberts/4640792
const (
	listenFdsStart = 3
	listenPid      = "LISTEN_PID"
	listenFds      = "LISTEN_FDS"
	listenFdNames  = "LISTEN_FDNAMES"
)

type NamedFd struct {
	File *os.File
	Name string
}

func unsetAll() {
	os.Unsetenv(listenPid)
	os.Unsetenv(listenFds)
	os.Unsetenv(listenFdNames)
}

func Files(unsetEnv bool) []*os.File {
	if unsetEnv {
		defer unsetAll()
	}

	pid, err := strconv.Atoi(os.Getenv(listenPid))
	if err != nil || pid != os.Getpid() {
		return nil
	}

	nfds, err := strconv.Atoi(os.Getenv(listenFds))
	if err != nil || nfds == 0 {
		return nil
	}

	files := make([]*os.File, 0, nfds)
	for fd := listenFdsStart; fd < listenFdsStart+nfds; fd++ {
		syscall.CloseOnExec(fd)
		files = append(files, os.NewFile(uintptr(fd), "LISTEN_FD_"+strconv.Itoa(fd)))
	}

	return files
}

func FilesWithNames(unsetEnv bool) []NamedFd {
	if unsetEnv {
		defer unsetAll()
	}
	files := Files(false)
	names := strings.Split(os.Getenv(listenFdNames), ":")

	namedFds := make([]NamedFd, len(files), len(files))
	for i := 0; i < len(files); i++ {
		namedFds[i].File = files[i]
		if i < len(names) {
			namedFds[i].Name = names[i]
		}
	}
	return namedFds
}
