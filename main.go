package main

import (
	"MyProxy/config"
	"crypto/tls"
	"fmt"
	"log"
	"net"

	"github.com/aravindc26/go-mysql/mysql"
	"github.com/aravindc26/go-mysql/server"
	"github.com/pkg/errors"
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

	mysql8server := getMySql8Server()

	for {
		c, err := l.Accept()
		if err != nil {
			log.Println(errors.Wrapf(err, "error accepting new connections"))
			continue
		}
		go handleConnection(c, credProvider, mysql8server)
	}
}

func handleConnection(c net.Conn, credProvider *server.InMemoryProvider, mysql8server *server.Server) {
	conn, err := server.NewCustomizedConn(c, mysql8server, credProvider, server.EmptyHandler{})
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

func getMySql8Server() *server.Server {
	caPem, caKey := server.GenerateCA()
	certPem, keyPem := server.GenerateAndSignRSACerts(caPem, caKey)
	tlsConf := server.NewServerTLSConfig(caPem, certPem, keyPem, tls.VerifyClientCertIfGiven)

	mysql8server := server.NewServer("8.0.0", mysql.DEFAULT_COLLATION_ID, mysql.AUTH_NATIVE_PASSWORD, server.GetPublicKeyFromCert(certPem), tlsConf)
	return mysql8server
}
