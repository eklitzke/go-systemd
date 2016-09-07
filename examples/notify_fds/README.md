# Notify FDs Example

This file demonstrates a simple HTTP server that supports graceful restarts. A
graceful restart in this context means that the HTTP server can be restarted
without any window where the listen socket is unavailable for new TCP
connections.

## Running

The best way to understand how this example works, and how to apply it to your
own programs, is to actually get it running under a real, bona fide systemd
instance.

To first set it up:

```bash
mkdir -p ~/.config/systemd/user
install -m 644 notify_fds.{service,socket} ~/.config/systemd/user
vi ~/.config/systemd/user/notify_fds.service  # make sure ExecStart has the correct executable
systemctl --user daemon-reload
systemctl --user start notify_fds.socket
```

Now verify that the socket activation feature works:

```bash
curl -v http://127.0.0.1:8076/
```

Now you should be able to verify that when executing `systemctl --user restart
notify_fds.service` at no point do you get `ECONNREFUSED` when establishing a
TCP connection to the server. Following is a very simple Go program that you can
use to validate this:

```go
package main

import (
	"flag"
	"fmt"
	"net"
)

var port = flag.Int("port", 8076, "the port to use")

func main() {
	flag.Parse()
	addr := fmt.Sprintf("127.0.0.1:%d", *port)
	fmt.Println(addr)
	errors := 0
	for {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			errors++
			fmt.Println(errors)
		} else {
			conn.Close()
		}
	}
}
```

When the service is *not* doing graceful restarts you'll see numbers printed to
the console as the service is restarted. If the graceful restart is working
correctly then you'll see no output from this program.

## Advanced Usage

This example only passes the listen file descriptor back to systemd. For true
graceful restarts you'll need to pass back client connections as well, and
serialize client state as appropriate. A full example is outside of the scope of
this documentation, but briefly what you'd do is:

 * before exiting, pass the listen socket + any client sockets back to systemd
   using `activation.SdNotifyWithFds()` and use the `FDNAMES` parameter in the
   state string to name the file descriptors
   * the format of `FDNAMES` is
     documented
     [here](https://www.freedesktop.org/software/systemd/man/sd_pid_notify_with_fds.html).
 * when starting, use `daemon.FilesWithNames()` to get back the listen socket +
   client sockets and associate the client sockets with names as appropriate
