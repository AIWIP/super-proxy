package main 

import (
	"fmt"
	"flag"
	"io"
	"io/ioutil"
	"log"
	
	"net"
	"net/http"

	"gopkg.in/yaml.v2"
)

const version = "0.1"

func handleConnection(config Config, w http.ResponseWriter, r *http.Request) {

	hijacker, ok := w.(http.Hijacker)

	if !ok {
		log.Fatalf("error: failed to hijack connection")
	}

	if r.Method == http.MethodConnect {
		w.WriteHeader(http.StatusOK)
	}

	clientConn, _, err := hijacker.Hijack()

	if err != nil {
        log.Fatalf("error: %v", err)
	}

	route := config.routeForTarget(r.URL)
	log.Print(r.URL.String() + " -> "  + route.Dest)

	serverConn, err := net.Dial("tcp", route.Dest)

	if err != nil {
        log.Fatalf("error: %v", err)
	}

	if r.Method == http.MethodConnect {

		go route.runMethod(r, clientConn, serverConn)

	} else {

		in, out := net.Pipe()
		
		go route.runMethod(r, out, serverConn)

		r.Write(in)
		io.Copy(clientConn, in)

		in.Close()
		out.Close()
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
		fmt.Printf(version)
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

	server := &http.Server{
        Addr: host,
        Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handleConnection(config, w, r)
        }),
    }

	err = server.ListenAndServe()

	if err != nil {
		log.Fatalf("error: %v", err)
	}
}