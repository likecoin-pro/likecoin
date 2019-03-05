package main

import (
	"github.com/likecoin-pro/likecoin/blockchain"
	"github.com/likecoin-pro/likecoin/blockchain/db"
	"github.com/likecoin-pro/likecoin/config"
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

	// start blockchain-replication
	go replication.NewService(nil, bc).StartReplication()

	select {}
}
