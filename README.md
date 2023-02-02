# yaaf-common-net

[![Build](https://github.com/go-yaaf/yaaf-common-net/actions/workflows/build.yml/badge.svg)](https://github.com/go-yaaf/yaaf-common-net/actions/workflows/build.yml)

Collection of network facilities and web technologies: web server, REST, web sockets, SMTP etc

## About
This library includes some concrete implementations of web servers for:
* REST server
* Web Socket client and server
* Static files web server


This library depends on the `yaaf-common` interface library

#### Adding dependency

```bash
$ go get -v -t github.com/go-yaaf/yaaf-common-net ./...
```

#### Publishing module
```bash
$ GOPROXY=proxy.golang.org go list -m github.com/go-yaaf/yaaf-common-net@v1.2
```