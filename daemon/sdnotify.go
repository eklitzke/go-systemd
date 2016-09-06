// Code forked from Docker project
package daemon

import (
	"net"
	"os"
	"syscall"
)

func SdNotifyWithFds(unsetEnvironment bool, state string, files ...*os.File) (sent bool, err error) {
	socketName := os.Getenv("NOTIFY_SOCKET")
	if socketName == "" {
		return false, nil
	}
	if unsetEnvironment {
		defer os.Unsetenv("NOTIFY_SOCKET")
	}
	socketAddr := &net.UnixAddr{
		Name: socketName,
		Net:  "unixgram",
	}

	conn, err := net.DialUnix(socketAddr.Net, nil, socketAddr)
	// Error connecting to NOTIFY_SOCKET
	if err != nil {
		return false, err
	}
	defer conn.Close()

	// Send the state message
	_, err = conn.Write([]byte(state))
	// Error sending the message
	if err != nil {
		return false, err
	}

	if len(files) == 0 {
		return true, nil
	}

	// Transfer file descriptor array
	unixFile, err := conn.File()
	if err != nil {
		return false, err
	}
	socket := int(unixFile.Fd())
	fds := make([]int, len(files))
	for i := range files {
		fds[i] = int(files[i].Fd())
	}
	rights := syscall.UnixRights(fds...)
	err = syscall.Sendmsg(socket, nil, rights, nil, 0)
	if err != nil {
		return false, err
	}
	return true, nil
}

// SdNotify sends a message to the init daemon. It is common to ignore the error.
// It returns one of the following:
// (false, nil) - notification not supported (i.e. NOTIFY_SOCKET is unset)
// (false, err) - notification supported, but failure happened (e.g. error connecting to NOTIFY_SOCKET or while sending data)
// (true, nil) - notification supported, data has been sent
func SdNotify(state string) (sent bool, err error) {
	return SdNotifyWithFds(false, state)
}
