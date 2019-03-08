# Super Proxy

Super Proxy is a HTTP proxy which supports proxying HTTP, HTTPS and other TCP based protocols,
this project has three goals:

- Production Ready:  Written in high-performance Go code
- Easily Configurable: Easy configured via a YAML configuration file
- Flexible: You can redirect to different servers and modify all traffic even if it isn't HTTP based

**Note: This codebase is still under-development and should be considered a beta.**

## Usage

### Running the proxy

Run the followin command to see instructions on how to run the proxy

```
super-proxy --help
```

```
./super-proxy -config  ~/Projects/printt/printt-cloud-print/src/proxy/conf/config.yaml -key ~/Projects/printt/printt-cloud-print/src/proxy/conf/rootCA.key -cert ~/Projects/printt/printt-cloud-print/src/proxy/conf/rootCA.crt
```

### Creating a configuration

Create a config.yaml file and use the example configuration below to configure
your proxy.

```

# Specifies the port the proxy will listen on
#
port: 8080

# Rules describe what the proxy should do when one of the requests matches. If the proxy
# can't match a request to a rule in this list then it will transparently forward it
#
rules:

  # If a hostname matches the target field then it will execute that rule,
  # if a request matches multiple rules the proxy will execute the first one it finds in this
  # list.
  #
  # target: bbc.com # Target any request for bbc.co.uk
  # target: bbc.com:8080 # Target requests over port 8080 for bbc.co.uk
  #
  # The destination field declares the hostname and port number the traffic
  # should be redirected to.
  #
  # dest: fake.bbc.com:80
  #
  # The method field is used to describe what the proxy should do with the traffic,
  # we have two options:
  #
  # - ForwardTransparent: Transparently forwards the traffic to the destination without
  #                       touching it
  #
  # - RewritePlain: Strips HTTP traffic of any TLS encyption, rewrites the HTTP request
  #                 so that it's for the destination and then forwards it
  #
  # If no method is provided then the proxy defaults to ForwardTransparent
  #
  - target: google.com
    dest: fake.google.com:80
    method: RewritePlain
  - target: youtube.com
    dest: fake.youtube.com:80
    method: ForwardTransparent
  - target: gmail.com
    dest: fake.gmail.com:80
  - target: github.com
    dest: fake.github.com:80
```

## Building From Source

Ensure you have Go 1.11+ installed on your machine and the GOPATH set.
Then to build simply run `make` inside of the directory you cloned this repository into.
