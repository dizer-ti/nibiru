package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/NibiruChain/nibiru/v2/x/common/testutil"
	"github.com/NibiruChain/nibiru/v2/x/common/testutil/testapp"
	inflationKeeper "github.com/NibiruChain/nibiru/v2/x/inflation/keeper"
	"github.com/NibiruChain/nibiru/v2/x/inflation/types"
)

func TestSuiteInflationSudo(t *testing.T) {
	suite.Run(t, new(SuiteInflationSudo))
}

type SuiteInflationSudo struct {
	suite.Suite
}

func (s *SuiteInflationSudo) TestMergeInflationParams() {
	currentParams := types.DefaultParams()

	newEpochsPerPeriod := sdkmath.NewInt(4)
	paramsChanges := types.MsgEditInflationParams{
		EpochsPerPeriod: &newEpochsPerPeriod,
	}

	paramsAfter, err := inflationKeeper.MergeInflationParams(paramsChanges, currentParams)
	s.Require().NoError(err)
	s.Require().EqualValues(4, paramsAfter.EpochsPerPeriod)

	// Test that the other params are unchanged.
	s.Require().EqualValues(currentParams.InflationEnabled, paramsAfter.InflationEnabled)
	s.Require().EqualValues(currentParams.PeriodsPerYear, paramsAfter.PeriodsPerYear)
	s.Require().EqualValues(currentParams.MaxPeriod, paramsAfter.MaxPeriod)
	s.Require().EqualValues(currentParams.PolynomialFactors, paramsAfter.PolynomialFactors)
	s.Require().EqualValues(currentParams.InflationDistribution, paramsAfter.InflationDistribution)

	// Test a change to all parameters
	newInflationDistribution := types.InflationDistribution{
		CommunityPool:     sdkmath.LegacyMustNewDecFromStr("0.8"),
		StakingRewards:    sdkmath.LegacyMustNewDecFromStr("0.1"),
		StrategicReserves: sdkmath.LegacyMustNewDecFromStr("0.1"),
	}

	paramsChanges = types.MsgEditInflationParams{
		EpochsPerPeriod: &newEpochsPerPeriod,
		PeriodsPerYear:  &newEpochsPerPeriod,
		MaxPeriod:       &newEpochsPerPeriod,
		PolynomialFactors: []sdkmath.LegacyDec{
			sdkmath.LegacyMustNewDecFromStr("0.1"),
			sdkmath.LegacyMustNewDecFromStr("0.2"),
		},
		InflationDistribution: &newInflationDistribution,
	}

	paramsAfter, err = inflationKeeper.MergeInflationParams(paramsChanges, currentParams)
	s.Require().NoError(err)
	s.Require().EqualValues(4, paramsAfter.EpochsPerPeriod)
	s.Require().EqualValues(4, paramsAfter.PeriodsPerYear)
	s.Require().EqualValues(4, paramsAfter.MaxPeriod)
	s.Require().EqualValues([]sdkmath.LegacyDec{
		sdkmath.LegacyMustNewDecFromStr("0.1"),
		sdkmath.LegacyMustNewDecFromStr("0.2"),
	}, paramsAfter.PolynomialFactors)
	s.Require().EqualValues(newInflationDistribution, paramsAfter.InflationDistribution)
}

func (s *SuiteInflationSudo) TestEditInflationParams() {
	nibiru, ctx := testapp.NewNibiruTestAppAndContext()

	// Change to all non-defaults to test EditInflationParams as a setter .
	epochsPerPeriod := sdkmath.NewInt(1_234)
	periodsPerYear := sdkmath.NewInt(1_234)
	maxPeriod := sdkmath.NewInt(1_234)
	polynomialFactors := []sdkmath.LegacyDec{
		sdkmath.LegacyMustNewDecFromStr("0.1"),
		sdkmath.LegacyMustNewDecFromStr("0.2"),
	}
	inflationDistribution := types.InflationDistribution{
		CommunityPool:     sdkmath.LegacyMustNewDecFromStr("0.8"),
		StakingRewards:    sdkmath.LegacyMustNewDecFromStr("0.1"),
		StrategicReserves: sdkmath.LegacyMustNewDecFromStr("0.1"),
	}
	msgEditParams := types.MsgEditInflationParams{
		EpochsPerPeriod:       &epochsPerPeriod,
		PeriodsPerYear:        &periodsPerYear,
		MaxPeriod:             &maxPeriod,
		PolynomialFactors:     polynomialFactors,
		InflationDistribution: &inflationDistribution,
	}

	s.T().Log("Params before MUST NOT be equal to default")
	defaultParams := types.DefaultParams()
	currParams, err := nibiru.InflationKeeper.Params.Get(ctx)
	s.Require().NoError(err)
	s.Require().Equal(currParams, defaultParams,
		"Current params should be eqaul to defaults")
	partialParams := msgEditParams

	s.T().Log("EditInflationParams should succeed")
	okSender := sdk.MustAccAddressFromBech32(testutil.ADDR_SUDO_ROOT)
	err = nibiru.InflationKeeper.Sudo().EditInflationParams(ctx, partialParams, okSender)
	s.Require().NoError(err)

	s.T().Log("Params after MUST be equal to partial params")
	paramsAfter, err := nibiru.InflationKeeper.Params.Get(ctx)
	s.Require().NoError(err)
	s.Require().EqualValues(1234, paramsAfter.EpochsPerPeriod)
	s.Require().EqualValues(1234, paramsAfter.PeriodsPerYear)
	s.Require().EqualValues(1234, paramsAfter.MaxPeriod)
	s.Require().EqualValues(polynomialFactors, paramsAfter.PolynomialFactors)
	s.Require().EqualValues(inflationDistribution, paramsAfter.InflationDistribution)
}

func (s *SuiteInflationSudo) TestToggleInflation() {
	nibiru, ctx := testapp.NewNibiruTestAppAndContext()

	err := nibiru.InflationKeeper.Sudo().ToggleInflation(ctx, true, sdk.MustAccAddressFromBech32(testutil.ADDR_SUDO_ROOT))
	s.Require().NoError(err)

	params, err := nibiru.InflationKeeper.Params.Get(ctx)
	s.Require().NoError(err)
	s.Require().True(params.InflationEnabled)

	err = nibiru.InflationKeeper.Sudo().ToggleInflation(ctx, false, sdk.MustAccAddressFromBech32(testutil.ADDR_SUDO_ROOT))
	s.Require().NoError(err)
	params, err = nibiru.InflationKeeper.Params.Get(ctx)
	s.Require().NoError(err)
	s.Require().False(params.InflationEnabled)
}
