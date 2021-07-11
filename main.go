package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type Rule struct {
	prefix  string
	handler http.HandlerFunc
}

type Registry struct {
	rules []*Rule
}

func matchUrlPrefix(path string, prefix string) bool {
	if strings.HasPrefix(path, prefix) {
		return true
	}
	return false
}

func NewRegistry() *Registry {
	return &Registry{
		rules: make([]*Rule, 0),
	}
}

func (registry *Registry) Register(prefix string, target string) error {
	remote, err := url.Parse(target)
	if err != nil {
		return err
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Forwarded-Host", r.Host)
		r.Host = remote.Host
		proxy.ServeHTTP(w, r)
	}

	rule := &Rule{
		prefix:  prefix,
		handler: handler,
	}
	registry.rules = append(registry.rules, rule)
	return nil
}

func (registry *Registry) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, rule := range registry.rules {
		if matchUrlPrefix(r.URL.Path, rule.prefix) {
			// remove the prefix
			r.URL.Path = strings.Replace(r.URL.Path, rule.prefix, "", 1)
			rule.handler(w, r)
			return
		}
	}
	http.Error(w, "no rule matches", http.StatusNotFound)
}

func main() {
	registry := NewRegistry()

	// some examples:
	registry.Register("/dogs", "https://random.dog/")
	registry.Register("/cats", "https://aws.random.cat/")

	err := http.ListenAndServe(":8080", registry)
	if err != nil {
		panic(err)
	}
}
