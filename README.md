To build:

```
go build socket-activated-http-server.go
```

This is an example of running a http go app with systemd socket activation. I also included a reference for how to do with [a minimal go container][1] (6.5M for this example). 


Drop the included `.service` and `.socket` files into `/etc/systemd/system/`. 


```
core@localhost ~ $ sudo systemctl status go-http-server
go-http-server.service - Socket Activated Go HTTP Server Example
   Loaded: loaded (/etc/systemd/system/go-http-server.service; static)
   Active: inactive (dead) since Sat 2013-06-15 21:22:29 UTC; 1min 7s ago

core@localhost $ sudo systemctl start go-http-server.socket 

core@localhost ~ $ sudo systemctl status go-http-server.socket
go-http-server.socket - Socket for Go HTTP Server Example
       Loaded: loaded (/etc/systemd/system/go-http-server.socket; static)
       Active: active (listening) since Sat 2013-06-15 21:21:29 UTC; 2min 42s ago
       Listen: [::]:8000 (Stream)

Jun 15 21:21:29 localhost systemd[1]: Starting Socket for Go HTTP Server Example.
Jun 15 21:21:29 localhost systemd[1]: Listening on Socket for Go HTTP Server Example.
```

This is where the actual socket activation occurs. Note that go-http-server is not started above:

```
core@localhost $ curl localhost:8000
Hello World!
Have interface: lo
Have interface: enp0s3
```

Now go-http-server is started!

```
core@localhost ~ $ sudo systemctl status go-http-server        
go-http-server.service - Socket Activated Go HTTP Server Example
   Loaded: loaded (/etc/systemd/system/go-http-server.service; static)
   Active: active (running) since Sat 2013-06-15 21:24:27 UTC; 13s ago
 Main PID: 3501 (socket-activate)
   CGroup: name=systemd:/system/go-http-server.service
           └─3501 /tmp/go-socket-activated-http-server-container/bin/socket-activated-http-server

Jun 15 21:24:27 localhost systemd[1]: Starting Socket Activated Go HTTP Server Example...
Jun 15 21:24:27 localhost systemd[1]: Started Socket Activated Go HTTP Server Example.
Jun 15 21:24:27 localhost socket-activated-http-server[3501]: served /
```

Now in a container. Edit go-http-server.service to read (Note the change of ExecStart):

```
[Unit]
Description=Socket Activated Go HTTP Server Example

[Service]
# To run directly
#ExecStart=/tmp/go-socket-activated-http-server-container/bin/socket-activated-http-server

# Works in a container with no networking either. Inspired by:
#   http://blog.oddbit.com/post/systemd-and-the-case-of-the-missing-network
#

ExecStart=/usr/bin/systemd-nspawn --private-network -D /tmp/go-socket-activated-http-server-container/ /bin/socket-activated-http-server
```

Reload systemd's configs and give it a shot

```
core@localhost ~ $ sudo systemctl --system daemon-reload
core@localhost ~ $ sudo systemctl stop go-http-server
Warning: Stopping go-http-server, but it can still be activated by:
  go-http-server.socket
core@localhost ~ $ curl localhost:8000
Hello World!
Have interface: lo
```

Note that it worked with a private interaface as well! This means the application was completely network isolated ([inspired by this post][2]), but was able to serve from a socket systemd gave it. 

[1]: https://github.com/polvi/go-socket-activated-http-server-container-amd64
[2]: http://blog.oddbit.com/post/systemd-and-the-case-of-the-missing-network
