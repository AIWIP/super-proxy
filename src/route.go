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

func (route Route) runMethod(r *http.Request, clientConn net.Conn, serverConn net.Conn) {

	route.Method(r, clientConn, serverConn)

	clientConn.Close()
	serverConn.Close()

	log.Print("Connection Closed")
}