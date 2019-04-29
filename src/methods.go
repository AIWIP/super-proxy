package main

import (
	"bufio"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"log"
)

type Method func(r *http.Request, clientConn net.Conn, serverConn net.Conn)
type MethodConstructor func(config Config) Method

func forwardTransparent(config Config) Method {

	return func(r *http.Request, clientConn net.Conn, serverConn net.Conn) {

		log.Print("Forwarding Transparently")

		go io.Copy(serverConn, clientConn)
		io.Copy(clientConn, serverConn)
	}
}

func rewritePlain(config Config) Method {

	// TODO: Load these values relative to the config not the actual binary
	//
	cert, err := tls.LoadX509KeyPair(config.Cert, config.Key)

	if err != nil {
		log.Fatalf("Error Loading Credentials: " + err.Error())
	}
	
	return func(r *http.Request, clientConn net.Conn, serverConn net.Conn) {

		log.Print("Rewriting Plain")

		var clientReader *bufio.Reader
		var connection net.Conn

		if r.Method == http.MethodConnect {

			TLSconfig := &tls.Config {
				Certificates: []tls.Certificate{cert},
			}
	
			tlsConn := tls.Server(clientConn, TLSconfig)
			tlsConn.Handshake()	

			clientReader = bufio.NewReader(tlsConn)
			connection = tlsConn

		} else {

			clientReader = bufio.NewReader(clientConn)
			connection = clientConn

			request, err := http.ReadRequest(clientReader)
			
			if err != nil {
				log.Fatalf("error: %v", err)
			}

		 	log.Print(request)

			request.Write(serverConn)
		}

		go io.Copy(serverConn, connection)
		
		serverReader := bufio.NewReader(serverConn)
		response, err := http.ReadResponse(serverReader, r)

		if err != nil {
			log.Fatalf("error: %v", err)
		}

		log.Print(response)
		response.Write(connection)
	}
}