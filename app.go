package main

import (
	"bytes"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	cmtlog "github.com/cometbft/cometbft/libs/log"
	tm "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cometbft/cometbft/version"
	db "kvstorepp/database"
	"time"
)

type KVStoreApplication struct {
	logger cmtlog.Logger
	db     *db.PebbleDB
	batch  db.Batch
}

func (app *KVStoreApplication) Info(_ abcitypes.RequestInfo) abcitypes.ResponseInfo {
	return abcitypes.ResponseInfo{
		Data:       "kvstore++",
		Version:    version.ABCIVersion,
		AppVersion: version.BlockProtocol,
	}
}

func (app *KVStoreApplication) Query(req abcitypes.RequestQuery) abcitypes.ResponseQuery {
	resp := abcitypes.ResponseQuery{Key: req.Data}

	item, err := app.db.Get(req.Data)
	if err != nil {
		resp.Log = "error getting value from application"
	} else {
		if item == nil {
			resp.Log = "item not found"
			resp.Value = []byte{}
		} else {
			resp.Log = "found item"
			resp.Value = item
		}
	}

	return resp
}

func (app *KVStoreApplication) CheckTx(req abcitypes.RequestCheckTx) abcitypes.ResponseCheckTx {
	code := app.isValid(req.Tx)
	return abcitypes.ResponseCheckTx{Code: code}
}

func (app *KVStoreApplication) InitChain(_ abcitypes.RequestInitChain) abcitypes.ResponseInitChain {
	return abcitypes.ResponseInitChain{}
}

func (app *KVStoreApplication) PrepareProposal(req abcitypes.RequestPrepareProposal) abcitypes.ResponsePrepareProposal {
	return abcitypes.ResponsePrepareProposal{Txs: req.Txs}
}

func (app *KVStoreApplication) ProcessProposal(_ abcitypes.RequestProcessProposal) abcitypes.ResponseProcessProposal {
	return abcitypes.ResponseProcessProposal{Status: abcitypes.ResponseProcessProposal_ACCEPT}
}

func (app *KVStoreApplication) BeginBlock(_ abcitypes.RequestBeginBlock) abcitypes.ResponseBeginBlock {
	app.batch = app.db.NewBatch()
	event := []abcitypes.Event{{
		Type: "begin_block_event",
		Attributes: []abcitypes.EventAttribute{
			{Key: "key", Value: "key", Index: true},
			{Key: "value", Value: "value", Index: true},
		},
	}}
	return abcitypes.ResponseBeginBlock{Events: event}
}

func (app *KVStoreApplication) DeliverTx(req abcitypes.RequestDeliverTx) abcitypes.ResponseDeliverTx {

	if code := app.isValid(req.Tx); code != 0 {
		app.logger.Error("abci", "method", "DeliverTx", "msg", "invalid tx", "code", code)
		return abcitypes.ResponseDeliverTx{
			Code: abcitypes.CodeTypeOK,
			Data: req.Tx,
			Log:  "tx invalid",
			Info: "tx processing info",
			Events: []abcitypes.Event{{
				Type: "deliver_tx_event",
				Attributes: []abcitypes.EventAttribute{{
					Key:   "key",
					Value: "value",
				}},
			}},
		}
	} else {
		parts := bytes.SplitN(req.Tx, []byte("="), 2)
		key, value := parts[0], parts[1]
		err := app.batch.Set(key, value)
		if err != nil {
			app.logger.Error("abci", "method", "FinalizeBlock", "msg", "error setting batch", "code", code)
			panic(err)
		}

		respDeliverTx := abcitypes.ResponseDeliverTx{
			Code: abcitypes.CodeTypeOK,
			Data: req.Tx,
			Log:  "tx committed successfully",
			Info: "tx processing info",
			Events: []abcitypes.Event{{
				Type: "deliver_tx_event",
				Attributes: []abcitypes.EventAttribute{{
					Key:   "key",
					Value: "value",
				}},
			}},
		}

		return respDeliverTx
	}
}

func (app *KVStoreApplication) EndBlock(_ abcitypes.RequestEndBlock) abcitypes.ResponseEndBlock {
	return abcitypes.ResponseEndBlock{
		ValidatorUpdates: nil,
		ConsensusParamUpdates: &tm.ConsensusParams{
			Block: &tm.BlockParams{
				MaxBytes: int64(10000000),
				MaxGas:   int64(100000),
			},
			Evidence: &tm.EvidenceParams{
				MaxAgeNumBlocks: int64(10),
				MaxAgeDuration:  time.Duration(10000),
				MaxBytes:        int64(10000000),
			},
			Validator: &tm.ValidatorParams{
				PubKeyTypes: []string{"ed25519"},
			},
			Version: &tm.VersionParams{
				App: uint64(1),
			},
		},
		Events: []abcitypes.Event{{
			Type: "deliver_tx_event",
			Attributes: []abcitypes.EventAttribute{{
				Key:   "key",
				Value: "value",
			}},
		}},
	}
}

func (app *KVStoreApplication) Commit() abcitypes.ResponseCommit {
	err := app.batch.Write()
	if err != nil {
		app.logger.Error("abci", "method", "Commit", "msg", "error writing commit", "err", err)
		panic(err)
	}
	return abcitypes.ResponseCommit{}
}

func (app *KVStoreApplication) ListSnapshots(snapshots abcitypes.RequestListSnapshots) abcitypes.ResponseListSnapshots {
	return abcitypes.ResponseListSnapshots{}
}

func (app *KVStoreApplication) OfferSnapshot(snapshot abcitypes.RequestOfferSnapshot) abcitypes.ResponseOfferSnapshot {
	return abcitypes.ResponseOfferSnapshot{}
}

func (app *KVStoreApplication) LoadSnapshotChunk(chunk abcitypes.RequestLoadSnapshotChunk) abcitypes.ResponseLoadSnapshotChunk {
	return abcitypes.ResponseLoadSnapshotChunk{}
}

func (app *KVStoreApplication) ApplySnapshotChunk(chunk abcitypes.RequestApplySnapshotChunk) abcitypes.ResponseApplySnapshotChunk {
	return abcitypes.ResponseApplySnapshotChunk{Result: abcitypes.ResponseApplySnapshotChunk_ACCEPT}
}

var _ abcitypes.Application = (*KVStoreApplication)(nil)

func NewKVStoreApplication(db *db.PebbleDB, logger cmtlog.Logger) *KVStoreApplication {
	return &KVStoreApplication{db: db, logger: logger}
}

func (app *KVStoreApplication) isValid(tx []byte) uint32 {
	// check format
	parts := bytes.Split(tx, []byte("="))
	if len(parts) != 2 {
		return 1
	}
	return 0
}
