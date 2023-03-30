package main

import (
	"github.com/bnb-chain/greenfield-go-sdk/client"
)

func main() {
	cli, _ := client.New()
	cli.Send()

	cli.Send()
	cli.QueryStorageProviders()
}
