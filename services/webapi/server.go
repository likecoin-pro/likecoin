package webapi

import (
	"net/http"
	"time"

	"github.com/likecoin-pro/likecoin/blockchain/db"
	"github.com/likecoin-pro/likecoin/commons/log"
)

type WebServer struct {
	cfg *Config
	bc  *db.BlockchainStorage
}

func StartServer(cfg *Config, bc *db.BlockchainStorage) error {
	s := &WebServer{
		cfg: cfg,
		bc:  bc,
	}
	return s.Start()
}

func (s *WebServer) Start() error {

	log.Info.Printf("webapi> Start Server: %s", s.cfg.HTTPConnStr)

	server := &http.Server{
		Addr:           s.cfg.HTTPConnStr,
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
