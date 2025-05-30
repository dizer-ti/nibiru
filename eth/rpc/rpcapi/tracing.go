// Copyright (c) 2023-2024 Nibi, Inc.
package rpcapi

import (
	"encoding/json"
	"fmt"
	"math"

	cmtrpcclient "github.com/cometbft/cometbft/rpc/client"
	tmrpctypes "github.com/cometbft/cometbft/rpc/core/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gethcommon "github.com/ethereum/go-ethereum/common"
	pkgerrors "github.com/pkg/errors"

	"github.com/NibiruChain/nibiru/v2/eth/rpc"
	"github.com/NibiruChain/nibiru/v2/x/evm"
)

// TraceTransaction returns the structured logs created during the execution of EVM
// and returns them as a JSON object.
func (b *Backend) TraceTransaction(
	hash gethcommon.Hash,
	config *evm.TraceConfig,
) (res json.RawMessage, err error) {
	// Get transaction by hash
	transaction, err := b.GetTxByEthHash(hash)
	if err != nil {
		b.logger.Debug("tx not found", "hash", hash)
		return nil, err
	}

	// check if block number is 0
	if transaction.Height == 0 {
		return nil, pkgerrors.New("genesis is not traceable")
	}

	blk, err := b.TendermintBlockByNumber(rpc.BlockNumber(transaction.Height))
	if err != nil {
		b.logger.Debug("block not found", "height", transaction.Height)
		return nil, err
	}

	// check tx index is not out of bound
	if len(blk.Block.Txs) > math.MaxUint32 {
		return nil, fmt.Errorf("tx count %d is overflowing", len(blk.Block.Txs))
	}
	txsLen := uint32(len(blk.Block.Txs)) // #nosec G701 -- checked for int overflow already
	if txsLen < transaction.TxIndex {
		b.logger.Debug("tx index out of bounds", "index", transaction.TxIndex, "hash", hash.String(), "height", blk.Block.Height)
		return nil, fmt.Errorf("transaction not included in block %v", blk.Block.Height)
	}

	var predecessors []*evm.MsgEthereumTx
	for _, txBz := range blk.Block.Txs[:transaction.TxIndex] {
		tx, err := b.clientCtx.TxConfig.TxDecoder()(txBz)
		if err != nil {
			b.logger.Debug("failed to decode transaction in block", "height", blk.Block.Height, "error", err.Error())
			continue
		}
		for _, msg := range tx.GetMsgs() {
			ethMsg, ok := msg.(*evm.MsgEthereumTx)
			if !ok {
				continue
			}

			predecessors = append(predecessors, ethMsg)
		}
	}

	tx, err := b.clientCtx.TxConfig.TxDecoder()(blk.Block.Txs[transaction.TxIndex])
	if err != nil {
		b.logger.Debug("tx not found", "hash", hash)
		return nil, err
	}

	// add predecessor messages in current cosmos tx
	for i := range tx.GetMsgs()[:int(transaction.MsgIndex)] {
		ethMsg, ok := tx.GetMsgs()[i].(*evm.MsgEthereumTx)
		if !ok {
			continue
		}
		predecessors = append(predecessors, ethMsg)
	}

	ethMessage, ok := tx.GetMsgs()[transaction.MsgIndex].(*evm.MsgEthereumTx)
	if !ok {
		b.logger.Debug("invalid transaction type", "type", fmt.Sprintf("%T", tx))
		return nil, fmt.Errorf("invalid transaction type %T", tx)
	}

	nc, ok := b.clientCtx.Client.(cmtrpcclient.NetworkClient)
	if !ok {
		return nil, pkgerrors.New("invalid rpc client")
	}

	cp, err := nc.ConsensusParams(b.ctx, &blk.Block.Height)
	if err != nil {
		return nil, err
	}

	traceTxRequest := evm.QueryTraceTxRequest{
		Msg:             ethMessage,
		Predecessors:    predecessors,
		BlockNumber:     blk.Block.Height,
		BlockTime:       blk.Block.Time,
		BlockHash:       gethcommon.Bytes2Hex(blk.BlockID.Hash),
		ProposerAddress: sdk.ConsAddress(blk.Block.ProposerAddress),
		ChainId:         b.chainID.Int64(),
		BlockMaxGas:     cp.ConsensusParams.Block.MaxGas,
	}

	if config != nil {
		traceTxRequest.TraceConfig = config
	}

	// Run "TraceTx":
	// For the trace context, use tx height minus one to get the context of start
	// of the block. We set the minimum value of "contextHeight" is set to 1
	// because 0 is a special value in [rpc.ContextWithHeight].
	traceResult, err := b.queryClient.TraceTx(
		rpc.NewContextWithHeight(max(transaction.Height-1, 1)),
		&traceTxRequest,
	)
	if err != nil {
		return nil, err
	}

	// Response format is unknown due to custom tracer config param
	// More information can be found here https://geth.ethereum.org/docs/dapp/tracing-filtered
	err = json.Unmarshal(traceResult.Data, &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// TraceBlock configures a new tracer according to the provided configuration, and
// executes all the transactions contained within. The return value will be one item
// per transaction, dependent on the requested tracer.
func (b *Backend) TraceBlock(height rpc.BlockNumber,
	config *evm.TraceConfig,
	block *tmrpctypes.ResultBlock,
) ([]*evm.TxTraceResult, error) {
	txs := block.Block.Txs
	txsLength := len(txs)

	if txsLength == 0 {
		// If there are no transactions return empty array
		return []*evm.TxTraceResult{}, nil
	}

	txDecoder := b.clientCtx.TxConfig.TxDecoder()

	var txsMessages []*evm.MsgEthereumTx
	for i, tx := range txs {
		decodedTx, err := txDecoder(tx)
		if err != nil {
			b.logger.Error("failed to decode transaction", "hash", txs[i].Hash(), "error", err.Error())
			continue
		}

		for _, msg := range decodedTx.GetMsgs() {
			ethMessage, ok := msg.(*evm.MsgEthereumTx)
			if !ok {
				// Just considers Ethereum transactions
				continue
			}
			txsMessages = append(txsMessages, ethMessage)
		}
	}

	// minus one to get the context at the beginning of the block
	contextHeight := max(height-1, 1) // 0 is a special value for `ContextWithHeight`.
	ctxWithHeight := rpc.NewContextWithHeight(int64(contextHeight))

	nc, ok := b.clientCtx.Client.(cmtrpcclient.NetworkClient)
	if !ok {
		return nil, pkgerrors.New("invalid rpc client")
	}

	cp, err := nc.ConsensusParams(b.ctx, &block.Block.Height)
	if err != nil {
		return nil, err
	}

	traceBlockRequest := &evm.QueryTraceBlockRequest{
		Txs:             txsMessages,
		TraceConfig:     config,
		BlockNumber:     block.Block.Height,
		BlockTime:       block.Block.Time,
		BlockHash:       gethcommon.Bytes2Hex(block.BlockID.Hash),
		ProposerAddress: sdk.ConsAddress(block.Block.ProposerAddress),
		ChainId:         b.chainID.Int64(),
		BlockMaxGas:     cp.ConsensusParams.Block.MaxGas,
	}

	res, err := b.queryClient.TraceBlock(ctxWithHeight, traceBlockRequest)
	if err != nil {
		return nil, err
	}

	decodedResults := make([]*evm.TxTraceResult, txsLength)
	if err := json.Unmarshal(res.Data, &decodedResults); err != nil {
		return nil, err
	}

	return decodedResults, nil
}

// TraceCall implements eth debug_traceCall method which lets you run an eth_call
// within the context of the given block execution using the final state of parent block as the base.
// Method returns the structured logs created during the execution of EVM.
// The method returns the same output as debug_traceTransaction.
// https://geth.ethereum.org/docs/interacting-with-geth/rpc/ns-debug#debugtracecall
func (b *Backend) TraceCall(
	txArgs evm.JsonTxArgs,
	contextBlock rpc.BlockNumber,
	config *evm.TraceConfig,
) (traceResult json.RawMessage, err error) {
	blk, err := b.TendermintBlockByNumber(contextBlock)
	if err != nil {
		b.logger.Debug("block not found", "contextBlock", contextBlock)
		return nil, err
	}
	nc, ok := b.clientCtx.Client.(cmtrpcclient.NetworkClient)
	if !ok {
		return nil, pkgerrors.New("invalid rpc client")
	}

	cp, err := nc.ConsensusParams(b.ctx, &blk.Block.Height)
	if err != nil {
		return nil, err
	}

	traceTxRequest := evm.QueryTraceTxRequest{
		Msg:             txArgs.ToMsgEthTx(),
		Predecessors:    nil,
		BlockNumber:     blk.Block.Height,
		BlockTime:       blk.Block.Time,
		BlockHash:       gethcommon.Bytes2Hex(blk.BlockID.Hash),
		ProposerAddress: sdk.ConsAddress(blk.Block.ProposerAddress),
		ChainId:         b.chainID.Int64(),
		BlockMaxGas:     cp.ConsensusParams.Block.MaxGas,
	}

	if config != nil {
		traceTxRequest.TraceConfig = config
	}
	traceResp, err := b.queryClient.TraceCall(rpc.NewContextWithHeight(contextBlock.Int64()), &traceTxRequest)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(traceResp.Data, &traceResult)
	if err != nil {
		return nil, err
	}
	return traceResult, nil
}
