package main

import (
	"bufio"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"log"
)

type Dialer func() (net.Conn, error)
type Method func(r *http.Request, clientConn net.Conn, serverDialer Dialer) error
type MethodConstructor func(config Config) Method


func checkWhitelistBeforeRunningMethod(config Config, originalMethod Method) Method {

	return func(r *http.Request, clientConn net.Conn, serverDialer Dialer) error {

		log.Print("Checking UA Whitelist")

		shouldAllow := false
		userAgent := r.UserAgent()

		if len(config.WhitelistedUseragents) == 0 {

			log.Print("No Whitelisted Useragents")

			shouldAllow = true

		} else {

			for _, pattern := range config.WhitelistedUseragents {

				log.Print("Checking " + pattern + " and " + userAgent)

				if pattern == userAgent {

					shouldAllow = true
					break
				}
			}
		}

		if shouldAllow {
			log.Print("Whitelisted")

			return originalMethod(r, clientConn, serverDialer)
		} else {

			log.Print("Blacklisted")

			log.Print(clientConn)
			clientConn.Write([]byte("HTTP/1.0 403 Forbidden\r\n\r\n")) 

			log.Print("Telling Client it's forbidden")

			return nil
		}
	}
}

func forwardTransparent(config Config) Method {

	return func(r *http.Request, clientConn net.Conn, serverDialer Dialer) error {

		log.Print("Forwarding Transparently")

		serverConn, err := serverDialer()

		if err != nil {
			return err
		}

		go io.Copy(serverConn, clientConn)
		io.Copy(clientConn, serverConn)

		return err
	}
}

func rewritePlain(config Config) Method {

	// TODO: Load these values relative to the config not the actual binary
	//
	cert, err := tls.LoadX509KeyPair(config.Cert, config.Key)

	if err != nil {
		log.Fatalf("Error Loading Credentials: " + err.Error())
	}
	
	return func(r *http.Request, clientConn net.Conn, serverDialer Dialer) error {

		log.Print("Rewriting Plain")

		serverConn, err := serverDialer()

		if err != nil {
			return err
		}

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

		 	log.Print(request)

			request.Write(serverConn)

			return err
		}

		go io.Copy(serverConn, connection)
		
		serverReader := bufio.NewReader(serverConn)
		response, err := http.ReadResponse(serverReader, r)

		log.Print(response)
		response.Write(connection)

		return err
	}
}