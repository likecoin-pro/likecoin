package main

import (
	"github.com/likecoin-pro/likecoin/blockchain"
	"github.com/likecoin-pro/likecoin/blockchain/db"
	"github.com/likecoin-pro/likecoin/config"
	"github.com/likecoin-pro/likecoin/services/client"
	"github.com/likecoin-pro/likecoin/services/replication"
	"github.com/likecoin-pro/likecoin/services/webapi"
)

func main() {
	// config
	apiCfg := webapi.NewConfig()
	bcCfg := blockchain.NewConfig()
	config.ParseArgs()

	// init blockchain
	bc := db.NewBlockchainStorage(bcCfg)

	// start web-server
	go webapi.StartServer(apiCfg, bc)

	// init client and start blockchain-replication
	cl := client.NewClient("https://likecoin.pro/api/v0")
	go replication.NewService(cl, bc).StartReplication()

	select {}
}
