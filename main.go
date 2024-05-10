package main

import (
	"flag"
	"fmt"
	abciserver "github.com/cometbft/cometbft/abci/server"
	"github.com/cometbft/cometbft/libs/service"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	cmtlog "github.com/cometbft/cometbft/libs/log"
	db "kvstorepp/database"
)

const defaultHomeDir = "$HOME/.kvstore-plus-plus"
const defaultSocket = "unix:///tmp/kvstore-plus-plus.sock"

var homeDir string
var socketAddr string

func init() {
	flag.StringVar(&homeDir, "home", "", fmt.Sprintf("Path to the kvstore directory (if empty, uses %s", defaultHomeDir))
	flag.StringVar(&socketAddr, "address", defaultSocket, fmt.Sprintf("Unix domain socket address (if empty, uses %s", defaultSocket))
}

func main() {
	flag.Parse()

	logger := cmtlog.NewTMLogger(cmtlog.NewSyncWriter(os.Stdout))

	if homeDir == "" {
		homeDir = os.ExpandEnv(defaultHomeDir)
	}
	if socketAddr == "" {
		socketAddr = defaultSocket
	}

	dbPath := filepath.Join(homeDir, "data")
	db, err := db.NewPebbleDB("kvstore-plus-plus", dbPath)
	if err != nil {
		log.Fatalf("Opening database: %v", err)
	}
	logger.Info("database start", "folder", dbPath)

	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("Closing database: %v", err)
		}
	}()

	app := NewKVStoreApplication(db, logger)

	server := abciserver.NewSocketServer(socketAddr, app)
	server.SetLogger(logger)

	if err := server.Start(); err != nil {
		logger.Error("server", "error starting socket server", "err", err)
		os.Exit(1)
	}
	logger.Info("server listening", "address", socketAddr)

	if server.IsRunning() {
		logger.Info("service running", "msg", "ABCI server is running")
	} else {
		logger.Error("server", "server is not running, aborting...")
		os.Exit(1)
	}

	defer func(server service.Service) {
		err := server.Stop()
		if err != nil {
			logger.Error("failed to stop the server", "err", err)
		}
	}(server)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}
