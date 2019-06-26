package main 

import (
	"fmt"
	"net/url"
)

type Rule struct {
	Target string		`yaml:"target"`
	Dest string			`yaml:"dest"`
	Method string		`yaml:"method"`
}

type Config struct {
	Port string			`yaml:"port"`
	Key string
	Cert string

	Rules []Rule		`yaml:"rules"`
	Methods map[string]Method
	WhitelistedUseragents []string `yaml:"whitelisted_useragents"`
}

func NewConfig() Config {
	return Config {
		"80",
		"",
		"",
		make([]Rule, 0),
		make(map[string]Method),
		make([]string, 0),
	}
}

func (c Config) registerMethod(name string, constructor MethodConstructor) {
	c.Methods[name] = constructor(c)
}

func (c Config) methodForRule(rule *Rule) Method {

	method := forwardTransparent(c)

	if rule != nil {
		possibleMethod := c.Methods[rule.Method]

		if possibleMethod != nil {
			method = possibleMethod
		}
	}

	return checkWhitelistBeforeRunningMethod(c, method)
}

func (c Config) routeForTarget(target *url.URL) Route {

	for _, rule := range c.Rules {

		hostname := target.Hostname()
		address := fmt.Sprintf("%v:%v", hostname, target.Port())

		if rule.Target == target.Hostname() || rule.Target == address {

			return Route {
				rule.Dest,
				c.methodForRule(&rule),
			}
		}
	}

	hostname := target.Hostname()
	port := target.Port()

	if port == "" {
		port = "80"
	}

	address := hostname + ":" + port

	return Route{
		address,
		c.methodForRule(nil),
	}
}