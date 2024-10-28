package main

import (
	"bytes"
	"context"
	"errors"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	cmtlog "github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/version"
	db "kvstore/database"
)

type KVStoreApplication struct {
	logger cmtlog.Logger
	db     *db.PebbleDB
	batch  db.Batch
}

var _ abcitypes.Application = (*KVStoreApplication)(nil)

func NewKVStoreApplication(db *db.PebbleDB, logger cmtlog.Logger) *KVStoreApplication {
	return &KVStoreApplication{db: db, logger: logger}
}

func (app *KVStoreApplication) Info(_ context.Context, info *abcitypes.RequestInfo) (*abcitypes.ResponseInfo, error) {
	return &abcitypes.ResponseInfo{
		Data:       "kvstore++",
		Version:    version.ABCIVersion,
		AppVersion: version.BlockProtocol,
	}, nil
}

func (app *KVStoreApplication) Query(_ context.Context, req *abcitypes.RequestQuery) (*abcitypes.ResponseQuery, error) {
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

	return &resp, nil
}

func (app *KVStoreApplication) CheckTx(_ context.Context, check *abcitypes.RequestCheckTx) (*abcitypes.ResponseCheckTx, error) {
	code := app.isValid(check.Tx)
	return &abcitypes.ResponseCheckTx{Code: code}, nil
}

func (app *KVStoreApplication) InitChain(_ context.Context, chain *abcitypes.RequestInitChain) (*abcitypes.ResponseInitChain, error) {
	return &abcitypes.ResponseInitChain{}, nil
}

func (app *KVStoreApplication) PrepareProposal(_ context.Context, proposal *abcitypes.RequestPrepareProposal) (*abcitypes.ResponsePrepareProposal, error) {
	return &abcitypes.ResponsePrepareProposal{Txs: proposal.Txs}, nil
}

func (app *KVStoreApplication) ProcessProposal(_ context.Context, proposal *abcitypes.RequestProcessProposal) (*abcitypes.ResponseProcessProposal, error) {
	return &abcitypes.ResponseProcessProposal{Status: abcitypes.ResponseProcessProposal_ACCEPT}, nil
}

func (app *KVStoreApplication) FinalizeBlock(_ context.Context, req *abcitypes.RequestFinalizeBlock) (*abcitypes.ResponseFinalizeBlock, error) {
	var txsResults = make([]*abcitypes.ExecTxResult, len(req.Txs))
	app.batch = app.db.NewBatch()
	for i, tx := range req.Txs {
		if code := app.isValid(tx); code != 0 {
			app.logger.Error("abci", "method", "FinalizeBlock", "msg", "invalid tx", "code", code)
			txsResults[i] = &abcitypes.ExecTxResult{Code: code}
		} else {
			parts := bytes.SplitN(tx, []byte("="), 2)
			key, value := parts[0], parts[1]
			err := app.batch.Set(key, value)
			if err != nil {
				app.logger.Error("abci", "method", "FinalizeBlock", "msg", "error setting batch", "code", code)
				return nil, err
			}
			txsResults[i] = &abcitypes.ExecTxResult{
				Code: 0,
				Events: []abcitypes.Event{
					{
						Type: "event",
						Attributes: []abcitypes.EventAttribute{
							{Key: "key", Value: string(key), Index: true},
							{Key: "value", Value: string(value), Index: true},
						},
					},
				},
			}
		}
	}
	return &abcitypes.ResponseFinalizeBlock{
		TxResults: txsResults,
	}, nil
}

func (app *KVStoreApplication) Commit(_ context.Context, commit *abcitypes.RequestCommit) (*abcitypes.ResponseCommit, error) {
	err := app.batch.Write()
	if err != nil {
		app.logger.Error("abci", "method", "Commit", "msg", "error writing batch", "err", err)
		return nil, errors.New("error during commit")
	}
	return &abcitypes.ResponseCommit{}, nil
}

func (app *KVStoreApplication) ListSnapshots(_ context.Context, snapshots *abcitypes.RequestListSnapshots) (*abcitypes.ResponseListSnapshots, error) {
	return &abcitypes.ResponseListSnapshots{}, nil
}

func (app *KVStoreApplication) OfferSnapshot(_ context.Context, snapshot *abcitypes.RequestOfferSnapshot) (*abcitypes.ResponseOfferSnapshot, error) {
	return &abcitypes.ResponseOfferSnapshot{}, nil
}

func (app *KVStoreApplication) LoadSnapshotChunk(_ context.Context, chunk *abcitypes.RequestLoadSnapshotChunk) (*abcitypes.ResponseLoadSnapshotChunk, error) {
	return &abcitypes.ResponseLoadSnapshotChunk{}, nil
}

func (app *KVStoreApplication) ApplySnapshotChunk(_ context.Context, chunk *abcitypes.RequestApplySnapshotChunk) (*abcitypes.ResponseApplySnapshotChunk, error) {
	return &abcitypes.ResponseApplySnapshotChunk{Result: abcitypes.ResponseApplySnapshotChunk_ACCEPT}, nil
}

func (app *KVStoreApplication) ExtendVote(_ context.Context, extend *abcitypes.RequestExtendVote) (*abcitypes.ResponseExtendVote, error) {
	return &abcitypes.ResponseExtendVote{}, nil
}

func (app *KVStoreApplication) VerifyVoteExtension(_ context.Context, verify *abcitypes.RequestVerifyVoteExtension) (*abcitypes.ResponseVerifyVoteExtension, error) {
	return &abcitypes.ResponseVerifyVoteExtension{}, nil
}

func (app *KVStoreApplication) isValid(tx []byte) uint32 {
	// check format
	parts := bytes.Split(tx, []byte("="))
	if len(parts) != 2 {
		return 1
	}
	return 0
}
