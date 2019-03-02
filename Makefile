all:
	go get gopkg.in/yaml.v2
	rm -f bin/super-proxy
	cd src && go build -o ../bin/super-proxy