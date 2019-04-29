package main 

import (
	"bufio"
	"fmt"
	"flag"
	"io"
	"io/ioutil"
	"log"
	
	"net"
	"net/http"
	"gopkg.in/yaml.v2"
)

const version = "0.4"

func handleConnection(config Config, conn net.Conn) {

	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		request, err := http.ReadRequest(reader)

		log.Print(request)

		if err != nil {
			log.Print("error: %v", err)
			conn.Write([]byte("HTTP/1.0 500 Server Error\r\n\r\n"))
			return
		}

		route := config.routeForTarget(request.URL)
		log.Print(request.URL.String() + " -> "  + route.Dest)

		serverConn, err := net.Dial("tcp", route.Dest)

		if err != nil {
			log.Print("error: %v", err)
			conn.Write([]byte("HTTP/1.0 500 Server Error\r\n\r\n"))
			return
		}

		in, out := net.Pipe()

		if request.Method == http.MethodConnect {

			log.Print("CONNECT")

			conn.Write([]byte("HTTP/1.0 200 OK\r\n\r\n"))
			go route.runMethod(request, out, serverConn)
			go io.Copy(in, conn)

		} else {

			log.Print("DIRECT")

			go route.runMethod(request, out, serverConn)
			request.Write(in)
		}

		io.Copy(conn, in)
	}
}

func main() {

	var configPath string
	var key string
	var cert string

	flag.StringVar(&configPath, "config", "", "path to the config file")
	flag.StringVar(&key, "key", "", "path to the key to use for TLS connections")
	flag.StringVar(&cert, "cert", "", "path to the cert to use for TLS connections")

	should_show_version := flag.Bool("version", false, "prints the version number")

	flag.Parse()

	if *should_show_version {
		fmt.Printf("%v\n", version)
		return
	}

	yamlFile, err := ioutil.ReadFile(configPath)

    if err != nil {
        log.Fatal(err)
	}
	
	config := NewConfig()
	err = yaml.Unmarshal(yamlFile, &config)

	config.Key = key
	config.Cert = cert

	config.registerMethod("ForwardTransparent", forwardTransparent)
	config.registerMethod("RewritePlain", rewritePlain)

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	log.Print("Super Proxy v", version)
	log.Print("")

	log.Print("Rules:")
	log.Print("")

	for _, rule := range config.Rules {
		log.Print(rule.Target, " -> ", rule.Dest)
	}

	host := ":" + config.Port

	log.Print("")
	log.Print("Listening On Port ", host)

	ln, err := net.Listen("tcp", host)

	ln = ln

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	for {
		conn, err := ln.Accept()

		if err != nil {
			log.Fatalf("error: %v", err)
		}

		go handleConnection(config, conn)
	}
}