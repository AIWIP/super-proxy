package main

import (
	"log"
	"net"
	"net/http"
)

type Route struct {
	Dest string
	Method Method
}

func (route Route) runMethod(r *http.Request, clientConn net.Conn) {

	var serverPointer *net.Conn = nil
	serverDialer := func () (net.Conn, error) {

		serverConn, err := net.Dial("tcp", route.Dest)

		if serverConn != nil {
			serverPointer = &serverConn
		}

		return serverConn, err 
	}

	err := route.Method(r, clientConn, serverDialer)
	
	if err != nil {

		log.Print("error: %v", err)
		clientConn.Write([]byte("HTTP/1.0 500 Server Error\r\n\r\n")) 
	}

	clientConn.Close()

	if serverPointer != nil {

		serverConn := *serverPointer
		log.Print(serverPointer)

		serverConn.Close()
	}
	

	log.Print("Connection Closed")
}