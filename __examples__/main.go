package main

import (
	"github.com/go-courier/courier"

	"github.com/go-courier/httptransport"
	"github.com/go-courier/httptransport/__examples__/routes"
)

func main() {
	ht := &httptransport.HttpTransport{
		Port: 8080,
	}
	ht.SetDefaults()

	courier.Run(routes.RootRouter, ht)
}
