// Copyright (c) 2023-2024 Nibi, Inc.
package evmante

import (
	"math/big"

	sdkioerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	gethcore "github.com/ethereum/go-ethereum/core/types"

	"github.com/NibiruChain/nibiru/v2/x/evm"
)

// EthSigVerificationDecorator validates an ethereum signatures
type EthSigVerificationDecorator struct {
	evmKeeper *EVMKeeper
}

// NewEthSigVerificationDecorator creates a new EthSigVerificationDecorator
func NewEthSigVerificationDecorator(k *EVMKeeper) EthSigVerificationDecorator {
	return EthSigVerificationDecorator{
		evmKeeper: k,
	}
}

// AnteHandle validates checks that the registered chain id is the same as the
// one on the message, and that the signer address matches the one defined on the
// message. It's not skipped for RecheckTx, because it set `From` address which
// is critical from other ante handler to work. Failure in RecheckTx will prevent
// tx to be included into block, especially when CheckTx succeed, in which case
// user won't see the error message.
func (esvd EthSigVerificationDecorator) AnteHandle(
	ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	chainID := esvd.evmKeeper.EthChainID(ctx)
	ethCfg := evm.EthereumConfig(chainID)
	blockNum := big.NewInt(ctx.BlockHeight())
	signer := gethcore.MakeSigner(
		ethCfg,
		blockNum,
		evm.ParseBlockTimeUnixU64(ctx),
	)

	for _, msg := range tx.GetMsgs() {
		msgEthTx, ok := msg.(*evm.MsgEthereumTx)
		if !ok {
			return ctx, sdkioerrors.Wrapf(
				sdkerrors.ErrUnknownRequest,
				"invalid message type %T, expected %T", msg, (*evm.MsgEthereumTx)(nil),
			)
		}

		ethTx := msgEthTx.AsTransaction()
		// if !ethTx.Protected() {
		// 	return ctx, errors.Wrapf(
		// 		sdkerrors.ErrNotSupported,
		// 		"rejected unprotected Ethereum transaction. "+
		// 			"Please EIP155 sign your transaction to protect it against replay-attacks",
		// 	)
		// }

		sender, err := signer.Sender(ethTx)
		if err != nil {
			return ctx, sdkioerrors.Wrapf(
				sdkerrors.ErrorInvalidSigner,
				"couldn't retrieve sender address from the ethereum transaction: %s",
				err.Error(),
			)
		}

		// set up the sender to the transaction field if not already
		msgEthTx.From = sender.Hex()
	}

	return next(ctx, tx, simulate)
}
