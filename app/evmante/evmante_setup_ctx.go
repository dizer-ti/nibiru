// Copyright (c) 2023-2024 Nibi, Inc.
package evmante

import (
	sdkioerrors "cosmossdk.io/errors"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
)

// EthSetupContextDecorator is adapted from SetUpContextDecorator from cosmos-sdk, it ignores gas consumption
// by setting the gas meter to infinite
type EthSetupContextDecorator struct {
	evmKeeper *EVMKeeper
}

func NewEthSetUpContextDecorator(k *EVMKeeper) EthSetupContextDecorator {
	return EthSetupContextDecorator{
		evmKeeper: k,
	}
}

func (esc EthSetupContextDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	// all transactions must implement GasTx
	_, ok := tx.(authante.GasTx)
	if !ok {
		return ctx, sdkioerrors.Wrapf(
			sdkerrors.ErrInvalidType,
			"invalid transaction type %T, expected GasTx", tx,
		)
	}

	// We need to setup an empty gas config so that the gas is consistent with Ethereum.
	newCtx = ctx.WithGasMeter(sdk.NewInfiniteGasMeter()).
		WithKVGasConfig(storetypes.GasConfig{}).
		WithTransientKVGasConfig(storetypes.GasConfig{})

	return next(newCtx, tx, simulate)
}
