package main

import (
	"MyProxy/config"
	"fmt"
	"github.com/go-mysql-org/go-mysql/server"
	"github.com/pkg/errors"
	"log"
	"net"
)

func main() {
	settings, err := config.NewConfigFromTOML("config.toml")
	if err != nil {
		log.Println(errors.Wrap(err, "error unmarshalling config.toml"))
		return
	}

	log.Printf("read config: %+v", settings)

	credProvider := server.NewInMemoryProvider()
	for _, v := range settings.Credentials {
		credProvider.AddUser(v.User, v.Password)
	}

	host := settings.Connection.Host
	port := settings.Connection.Port
	port = fmt.Sprintf("%s:%s", host, port)

	l, err := net.Listen("tcp", port)
	if err != nil {
		log.Println(errors.Wrapf(err, "error establishing connection to %s", port))
		return
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			log.Println(errors.Wrapf(err, "error accepting new connections"))
			continue
		}
		go handleConnection(c, credProvider)
	}
}

func handleConnection(c net.Conn, credProvider *server.InMemoryProvider) {
	defaultServer := server.NewDefaultServer()
	conn, err := server.NewCustomizedConn(c, defaultServer, credProvider, server.EmptyHandler{})
	if err != nil {
		log.Println(errors.Wrap(err, "error creating connection handler"))
		return
	}

	for {
		if err := conn.HandleCommand(); err != nil {
			log.Println(errors.WithStack(err))
			return
		}
	}
}
