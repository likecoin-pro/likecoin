package webapi

import (
	"net/http"
	"time"

	"github.com/likecoin-pro/likecoin/blockchain/db"
	"github.com/likecoin-pro/likecoin/commons/log"
)

type WebServer struct {
	bc *db.BlockchainStorage
}

func StartServer(connStr string, bc *db.BlockchainStorage) error {
	s := &WebServer{bc: bc}
	return s.Start(connStr)
}

func (s *WebServer) Start(connStr string) error {

	log.Info.Printf("HTTP: Start Server: %s", connStr)

	server := &http.Server{
		Addr:           connStr,
		Handler:        s,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Panic(err)
	}
	return nil
}

func (s *WebServer) ServeHTTP(rw http.ResponseWriter, rq *http.Request) {
	NewContext(rq, rw, s.bc, "").Exec()
}
