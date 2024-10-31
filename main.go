package main

import (
	"flag"
	"kvstorepp/database"

	abciserver "github.com/cometbft/cometbft/abci/server"
	cmtlog "github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/service"

	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

const defaultHomeDir = "$HOME/.kvstore-plus-plus"
const defaultSocket = "unix:///tmp/kvstore-plus-plus.sock"

var homeDir string
var socketAddr string
var app string

func init() {
	flag.StringVar(&homeDir, "home", "", "Path to the kvstore directory (if empty, uses $HOME/.kvstore)")
	flag.StringVar(&app, "app", "", "If you want to run the malachite app enter malachite here, otherwise it runs the regular kvstore")
	flag.StringVar(&socketAddr, "address", "unix://example.sock", "Unix domain socket address (if empty, uses \"unix://example.sock\"")
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
	db, err := database.NewPebbleDB("kvstore-plus-plus", dbPath)
	if err != nil {
		log.Fatalf("Opening database: %v", err)
	}
	logger.Info("database start", "folder", dbPath)

	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("Closing database: %v", err)
		}
	}()

	var server service.Service

	if app == "malachite" {
		app := NewKVStoreApplicationMalachite(db, logger)
		server = abciserver.NewSocketServer(socketAddr, app)

	} else {
		app := NewKVStoreApplication(db, logger)
		server = abciserver.NewSocketServer(socketAddr, app)
	}

	// server := abciserver.NewSocketServer(socketAddr, app)
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
