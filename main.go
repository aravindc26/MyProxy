package main

import (
	"github.com/go-mysql-org/go-mysql/server"
	"github.com/pkg/errors"
	"log"
	"net"
	"os"
)

func main() {
	args := os.Args[1:]
	l, err := net.Listen("tcp", args[0])
	if err != nil {
		log.Println(errors.Wrapf(err, "error establishing connection to %s", args[0]))
		return
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			log.Println(errors.Wrapf(err, "error accepting new connections"))
			continue
		}
		go handleConnection(c)
	}
}

func handleConnection(c net.Conn) {
	_, err := server.NewConn(c, "root", "", server.EmptyHandler{})
	if err != nil {
		log.Println(errors.Wrap(err, "error creating connection handler"))
	}
}
