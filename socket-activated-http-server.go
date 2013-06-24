package main

import (
        "os"
        "strconv"
        "syscall"
        "fmt"
        "net"
        "net/http"
)

const (
        listenFdsStart = 3
)

// Example HTTP server that uses systemd's socket activation
// Based on code snippet:
//     https://gist.github.com/alberts/4640792
func fcntl(fd int, cmd int, arg int) (val int, errno int) {
        r0, _, e1 := syscall.Syscall(syscall.SYS_FCNTL, uintptr(fd), uintptr(cmd), uintptr(arg))
        val = int(r0)
        errno = int(e1)
        return
}

func ListenFds() []*os.File {
        pid, err := strconv.Atoi(os.Getenv("LISTEN_PID"))
        if err != nil || pid != os.Getpid() {
                return nil
        }
        nfds, err := strconv.Atoi(os.Getenv("LISTEN_FDS"))
        if err != nil || nfds == 0 {
                return nil
        }
        files := []*os.File(nil)
        for fd := listenFdsStart; fd < listenFdsStart+nfds; fd++ {
                flags, errno := fcntl(fd, syscall.F_GETFD, 0)
                if errno != 0 {
                        return nil
                }
                if flags&syscall.FD_CLOEXEC != 0 {
                        continue
                }
                syscall.CloseOnExec(fd)
                files = append(files, os.NewFile(uintptr(fd), ""))
        }
        return files
}


func handler(w http.ResponseWriter, r *http.Request) {
	// easy logging to the journal :)
	fmt.Println("served", r.URL)

	w.Header().Set("Content-Type", "text/plain")
 	fmt.Fprintf(w, "Hello World!\n")

	// We are printing the interfaces to show what happens when ran inside 
	// of a container with --private-networking
	ifaces, _ := net.Interfaces()
	for _, iface := range ifaces {
	 	fmt.Fprintf(w, "Have interface: %s\n", iface.Name)
	}
}

// TODO: Currently only supports one socket from systemd
func main() {
	listen_fds := ListenFds()

	for _, fd := range listen_fds {
		l, err := net.FileListener(fd)
		if err != nil {
			// handle error
			fmt.Println("got err", err)
		}

		http.HandleFunc("/", handler)
		http.Serve(l, nil)
	}
}
