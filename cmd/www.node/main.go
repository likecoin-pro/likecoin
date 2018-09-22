package main

import (
	"flag"

	"github.com/likecoin-pro/likecoin/blockchain/db"
	"github.com/likecoin-pro/likecoin/config"
	"github.com/likecoin-pro/likecoin/services/client"
	"github.com/likecoin-pro/likecoin/services/replication"
	"github.com/likecoin-pro/likecoin/services/webapi"
)

func main() {
	var (
		argHTTPConn = ":8888"
		argVacuum   = false
	)
	flag.StringVar(&argHTTPConn, "http", argHTTPConn, "http-connection")
	flag.BoolVar(&argVacuum, "vacuum", argVacuum, "vacuum db")
	config.ParseArgs()

	// init blockchain
	bc := db.NewBlockchainStorage(config.ChainID, config.DataDir)

	// vacuum blockchain-db
	if argVacuum {
		bc.VacuumDB()
	}

	// start web-server
	go webapi.StartServer(argHTTPConn, bc)

	// init client and start blockchain-replication
	cl := client.NewClient("http://likecoin.pro/api/v0")
	go replication.NewService(cl, bc).StartReplication()

	select {}
}
