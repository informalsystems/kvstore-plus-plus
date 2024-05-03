# KVStore++

KVStore++ is an ABCI application that can be used for testing with multiple CometBFT releases

This application runs in a separate process than CometBFT, and it connects to CometBFT via a UNIX socket.

## Branches 

This repo includes multiple `branches` that support different KV Stores running on different CometBFT releases 
- `main`
- `v1-alpha`
- `v0.38`
- `v0.37`

### CometBFT target releases

| KVStore Branch | CometBFT release                                      |
|--------------------|-------------------------------------------------------|
| main               | v1.0.0-alpha.2 (always latest major CometBFT release) |
| v1                 | v1.0.0-alpha.2                                        |   
| v0.38              | v0.38                                                 |
| v0.37              | v0.37                                                 |
 


This is useful to test and debug behavior/code changes across CometBFT releases

## Running

In a terminal window in the current directory for this project, run the CometBFT release that you want to test with
, for example, run the commands below to initialize and run CometBFT:

```
go run github.com/cometbft/cometbft/cmd/cometbft@v1.0.0-alpha.2 init --home /tmp/cometbft-kv
go run github.com/cometbft/cometbft/cmd/cometbft@v1.0.0-alpha.2 start --home /tmp/cometbft-kv --proxy_app unix://example.sock
```

On another terminal window in the same project folder run the KVStore application:

```
go build
./kvstore
```

