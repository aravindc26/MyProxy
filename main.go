package main

import (
	"MyProxy/config"
	"MyProxy/handler"
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"

	"github.com/aravindc26/go-mysql/client"
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

	connPool := client.NewPool(log.Printf, 100, 400, 5, "127.0.0.1:3306", "root", "S3cret", "v")

	conn, err := connPool.GetConn(context.Background())
	if err != nil {
		log.Printf("error fetching connection %v", err)
	}
	defer conn.Close()

	err = conn.Ping()
	if err != nil {
		log.Fatal("ping err", err)
	}

	mysqlHandler := handler.NewProxyHandler(connPool)

	for {
		c, err := l.Accept()
		if err != nil {
			log.Println(errors.Wrapf(err, "error accepting new connections"))
			continue
		}
		go handleConnection(c, credProvider, mysql8server, mysqlHandler)
	}
}

func handleConnection(c net.Conn, credProvider *server.InMemoryProvider, mysql8server *server.Server, mysqlHandler *handler.ProxyHandler) {
	conn, err := server.NewCustomizedConn(c, mysql8server, credProvider, mysqlHandler)
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
