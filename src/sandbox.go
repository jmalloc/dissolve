package main

import (
	"context"

	"github.com/jmalloc/dissolve/src/dissolve/bonjour"
	"github.com/jmalloc/dissolve/src/dissolve/dnssd"
	"github.com/jmalloc/dissolve/src/dissolve/server"
	"github.com/jmalloc/twelf/src/twelf"
)

func main() {
	a := &bonjour.Answerer{}

	server := server.MulticastServer{
		Answerer: a,
		Logger:   twelf.DefaultLogger,
	}

	i, err := dnssd.NewInstance(
		"test", "_foo._tcp", "local.",
		"test.foobar.local.", 8080,
	)
	if err != nil {
		panic(err)
	}

	a.AddInstance(i)

	if err := server.Run(context.Background()); err != nil {
		panic(err)
	}
}
