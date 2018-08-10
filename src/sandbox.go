package main

import (
	"context"

	"github.com/jmalloc/dissolve/src/dissolve/dnssd"
	"github.com/jmalloc/dissolve/src/dissolve/mdns"
	"github.com/jmalloc/twelf/src/twelf"
)

func main() {
	h := &dnssd.Handler{
		Resolver: mdns.NewLocalResolver(nil),
	}

	server := &mdns.Server{
		Handler: h,
		Logger:  twelf.DebugLogger,
	}

	i, err := dnssd.NewInstance(
		"svc7", "_dissolve._tcp", "local.",
		"test7.dissolve.local.", 8080,
	)
	if err != nil {
		panic(err)
	}

	h.AddInstance(i)

	if err := server.Run(context.Background()); err != nil {
		panic(err)
	}
}
