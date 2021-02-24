package main

import (
	"github.com/go-courier/courier"
	"github.com/go-courier/httptransport/__examples__/server/cmd/app/routes"

	"github.com/go-courier/httptransport"
)

func main() {
	ht := &httptransport.HttpTransport{
		Port: 8080,
	}
	ht.SetDefaults()

	courier.Run(routes.RootRouter, ht)
}
