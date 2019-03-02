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

		TLSconfig := &tls.Config {
			Certificates: []tls.Certificate{cert}, 
			CipherSuites: []uint16{
				tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
			},
			ClientAuth: tls.NoClientCert,  
			ServerName: r.URL.Hostname(),
			InsecureSkipVerify: true,
		}

		tlsConn := tls.Server(clientConn, TLSconfig)
		tlsConn.Handshake()	

		reader := bufio.NewReader(tlsConn)
		request, err := http.ReadRequest(reader)

		if err != nil {
			log.Fatalf("error: %v", err)
		}

		request.Write(serverConn)
		
		go io.Copy(serverConn, tlsConn)
		io.Copy(tlsConn, serverConn)
	}
}