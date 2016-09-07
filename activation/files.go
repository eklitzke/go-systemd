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

// NamedFile represents a file descriptor that systemd is storing on behalf of
// the user program, along with the name systemd has given it. Typically the
// file object will be a socket, but other file types can be stored by
// SdNotifyWithFds.
type NamedFile struct {
	File *os.File // the file object
	Name string   // the name given by systemd
}

func unsetAll() {
	os.Unsetenv(listenPid)
	os.Unsetenv(listenFds)
	os.Unsetenv(listenFdNames)
}

// Files retrieves the file objects stored by systemd.
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

// FilesWithNames retrieves the file objects stored by systemd, along with their
// names. This will typically be used by programs using the SdNotifyWithFds
// method to store client non-listen sockets.
func FilesWithNames(unsetEnv bool) []NamedFile {
	if unsetEnv {
		defer unsetAll()
	}

	// delegate to Files() to actually fetch the file descriptors
	files := Files(false)

	// parse out the file names
	names := strings.Split(os.Getenv(listenFdNames), ":")
	namedFiles := make([]NamedFile, len(files), len(files))
	for i := 0; i < len(files); i++ {
		namedFiles[i].File = files[i]
		if i < len(names) {
			namedFiles[i].Name = names[i]
		}
	}
	return namedFiles
}
