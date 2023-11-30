//go:build integrationtests
// +build integrationtests

package processes_test

import (
	"context"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	rippledata "github.com/rubblelabs/ripple/data"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum-tools/pkg/retry"
	"github.com/CoreumFoundation/coreum/v3/pkg/client"
	coreumintegration "github.com/CoreumFoundation/coreum/v3/testutil/integration"
	assetfttypes "github.com/CoreumFoundation/coreum/v3/x/asset/ft/types"
	integrationtests "github.com/CoreumFoundation/coreumbridge-xrpl/integration-tests"
	"github.com/CoreumFoundation/coreumbridge-xrpl/relayer/coreum"
	"github.com/CoreumFoundation/coreumbridge-xrpl/relayer/xrpl"
)

func TestSendXRPLOriginatedTokensFromXRPLToCoreumAndBack(t *testing.T) {
	t.Parallel()

	ctx, chains := integrationtests.NewTestingContext(t)

	envCfg := DefaultRunnerEnvConfig()
	runnerEnv := NewRunnerEnv(ctx, t, envCfg, chains)
	runnerEnv.StartAllRunnerProcesses(ctx, t)
	runnerEnv.AllocateTickets(ctx, t, uint32(200))

	coreumSender := chains.Coreum.GenAccount()
	chains.Coreum.FundAccountWithOptions(ctx, t, coreumSender, coreumintegration.BalancesOptions{
		Amount: sdkmath.NewIntFromUint64(1_000_000),
	})
	t.Logf("Coreum sender: %s", coreumSender.String())
	xrplRecipientAddress := chains.XRPL.GenAccount(ctx, t, 0)
	t.Logf("XRPL recipient: %s", xrplRecipientAddress.String())

	registeredXRPLCurrency, err := rippledata.NewCurrency("RCP")
	require.NoError(t, err)

	xrplIssuerAddress := chains.XRPL.GenAccount(ctx, t, 1)
	// enable to be able to send to any address
	runnerEnv.EnableXRPLAccountRippling(ctx, t, xrplIssuerAddress)
	registeredXRPLToken := runnerEnv.RegisterXRPLOriginatedToken(ctx, t, xrplIssuerAddress, registeredXRPLCurrency, int32(6), integrationtests.ConvertStringWithDecimalsToSDKInt(t, "1", 30))

	valueToSendFromXRPLtoCoreum, err := rippledata.NewValue("1e10", false)
	require.NoError(t, err)
	amountToSendFromXRPLtoCoreum := rippledata.Amount{
		Value:    valueToSendFromXRPLtoCoreum,
		Currency: registeredXRPLCurrency,
		Issuer:   xrplIssuerAddress,
	}
	memo, err := xrpl.EncodeCoreumRecipientToMemo(coreumSender)
	require.NoError(t, err)

	runnerEnv.SendXRPLPaymentTx(ctx, t, xrplIssuerAddress, runnerEnv.bridgeXRPLAddress, amountToSendFromXRPLtoCoreum, memo)
	runnerEnv.AwaitCoreumBalance(ctx, t, chains.Coreum, coreumSender, sdk.NewCoin(registeredXRPLToken.CoreumDenom, integrationtests.ConvertStringWithDecimalsToSDKInt(t, valueToSendFromXRPLtoCoreum.String(), XRPLTokenDecimals)))

	// send the full amount in 4 transactions to XRPL
	amountToSend := integrationtests.ConvertStringWithDecimalsToSDKInt(t, valueToSendFromXRPLtoCoreum.String(), XRPLTokenDecimals).QuoRaw(4)

	// send 2 transactions without the trust set to be reverted
	// TODO(dzmitryhil) update assertion once we add the final tx revert/recovery
	_, err = runnerEnv.ContractClient.SendToXRPL(ctx, coreumSender, xrplRecipientAddress.String(), sdk.NewCoin(registeredXRPLToken.CoreumDenom, amountToSend))
	require.NoError(t, err)
	_, err = runnerEnv.ContractClient.SendToXRPL(ctx, coreumSender, xrplRecipientAddress.String(), sdk.NewCoin(registeredXRPLToken.CoreumDenom, amountToSend))
	require.NoError(t, err)
	runnerEnv.AwaitNoPendingOperations(ctx, t)

	// send TrustSet to be able to receive coins
	runnerEnv.SendXRPLMaxTrustSetTx(ctx, t, xrplRecipientAddress, xrplIssuerAddress, registeredXRPLCurrency)

	_, err = runnerEnv.ContractClient.SendToXRPL(ctx, coreumSender, xrplRecipientAddress.String(), sdk.NewCoin(registeredXRPLToken.CoreumDenom, amountToSend))
	require.NoError(t, err)
	_, err = runnerEnv.ContractClient.SendToXRPL(ctx, coreumSender, xrplRecipientAddress.String(), sdk.NewCoin(registeredXRPLToken.CoreumDenom, amountToSend))
	require.NoError(t, err)
	runnerEnv.AwaitNoPendingOperations(ctx, t)

	balance := runnerEnv.Chains.XRPL.GetAccountBalance(ctx, t, xrplRecipientAddress, xrplIssuerAddress, registeredXRPLCurrency)
	require.Equal(t, "5000000000", balance.Value.String())
}

func TestSendXRPLOriginatedTokenFromXRPLToCoreumWithMaliciousRelayer(t *testing.T) {
	t.Parallel()

	ctx, chains := integrationtests.NewTestingContext(t)

	envCfg := DefaultRunnerEnvConfig()
	envCfg.MaliciousRelayerNumber = 1
	runnerEnv := NewRunnerEnv(ctx, t, envCfg, chains)
	runnerEnv.StartAllRunnerProcesses(ctx, t)
	runnerEnv.AllocateTickets(ctx, t, uint32(200))

	coreumSender := chains.Coreum.GenAccount()
	chains.Coreum.FundAccountWithOptions(ctx, t, coreumSender, coreumintegration.BalancesOptions{
		Amount: sdkmath.NewIntFromUint64(1_000_000),
	})
	t.Logf("Coreum sender: %s", coreumSender.String())
	xrplRecipientAddress := chains.XRPL.GenAccount(ctx, t, 0)
	t.Logf("XRPL recipient: %s", xrplRecipientAddress.String())

	currencyHexSymbol := hex.EncodeToString([]byte(strings.Repeat("X", 20)))
	registeredXRPLCurrency, err := rippledata.NewCurrency(currencyHexSymbol)
	require.NoError(t, err)

	xrplIssuerAddress := chains.XRPL.GenAccount(ctx, t, 1)
	// enable to be able to send to any address
	runnerEnv.EnableXRPLAccountRippling(ctx, t, xrplIssuerAddress)
	registeredXRPLToken := runnerEnv.RegisterXRPLOriginatedToken(ctx, t, xrplIssuerAddress, registeredXRPLCurrency, int32(6), integrationtests.ConvertStringWithDecimalsToSDKInt(t, "1", 30))

	valueToSendFromXRPLtoCoreum, err := rippledata.NewValue("1e10", false)
	require.NoError(t, err)
	amountToSendFromXRPLtoCoreum := rippledata.Amount{
		Value:    valueToSendFromXRPLtoCoreum,
		Currency: registeredXRPLCurrency,
		Issuer:   xrplIssuerAddress,
	}
	memo, err := xrpl.EncodeCoreumRecipientToMemo(coreumSender)
	require.NoError(t, err)

	runnerEnv.SendXRPLPaymentTx(ctx, t, xrplIssuerAddress, runnerEnv.bridgeXRPLAddress, amountToSendFromXRPLtoCoreum, memo)
	runnerEnv.AwaitCoreumBalance(ctx, t, chains.Coreum, coreumSender, sdk.NewCoin(registeredXRPLToken.CoreumDenom, integrationtests.ConvertStringWithDecimalsToSDKInt(t, valueToSendFromXRPLtoCoreum.String(), XRPLTokenDecimals)))

	// send TrustSet to be able to receive coins
	runnerEnv.SendXRPLMaxTrustSetTx(ctx, t, xrplRecipientAddress, xrplIssuerAddress, registeredXRPLCurrency)

	amountToSend := integrationtests.ConvertStringWithDecimalsToSDKInt(t, valueToSendFromXRPLtoCoreum.String(), XRPLTokenDecimals).QuoRaw(4)
	_, err = runnerEnv.ContractClient.SendToXRPL(ctx, coreumSender, xrplRecipientAddress.String(), sdk.NewCoin(registeredXRPLToken.CoreumDenom, amountToSend))
	require.NoError(t, err)
	runnerEnv.AwaitNoPendingOperations(ctx, t)

	balance := runnerEnv.Chains.XRPL.GetAccountBalance(ctx, t, xrplRecipientAddress, xrplIssuerAddress, registeredXRPLCurrency)
	require.Equal(t, "2500000000", balance.Value.String())
}

func TestSendXRPLOriginatedTokenFromXRPLToCoreumWithTicketsReallocation(t *testing.T) {
	t.Parallel()

	ctx, chains := integrationtests.NewTestingContext(t)

	envCfg := DefaultRunnerEnvConfig()
	envCfg.UsedTicketSequenceThreshold = 3
	runnerEnv := NewRunnerEnv(ctx, t, envCfg, chains)
	runnerEnv.StartAllRunnerProcesses(ctx, t)
	runnerEnv.AllocateTickets(ctx, t, uint32(5))
	sendingCount := 10

	coreumSender := chains.Coreum.GenAccount()
	chains.Coreum.FundAccountWithOptions(ctx, t, coreumSender, coreumintegration.BalancesOptions{
		Amount: sdkmath.NewIntFromUint64(1_000_000),
	})
	t.Logf("Coreum sender: %s", coreumSender.String())
	xrplRecipientAddress := chains.XRPL.GenAccount(ctx, t, 0)
	t.Logf("XRPL recipient: %s", xrplRecipientAddress.String())

	currencyHexSymbol := hex.EncodeToString([]byte(strings.Repeat("X", 20)))
	registeredXRPLCurrency, err := rippledata.NewCurrency(currencyHexSymbol)
	require.NoError(t, err)

	xrplIssuerAddress := chains.XRPL.GenAccount(ctx, t, 1)
	// enable to be able to send to any address
	runnerEnv.EnableXRPLAccountRippling(ctx, t, xrplIssuerAddress)
	registeredXRPLToken := runnerEnv.RegisterXRPLOriginatedToken(ctx, t, xrplIssuerAddress, registeredXRPLCurrency, int32(6), integrationtests.ConvertStringWithDecimalsToSDKInt(t, "1", 30))

	valueToSendFromXRPLtoCoreum, err := rippledata.NewValue("1e10", false)
	require.NoError(t, err)
	amountToSendFromXRPLtoCoreum := rippledata.Amount{
		Value:    valueToSendFromXRPLtoCoreum,
		Currency: registeredXRPLCurrency,
		Issuer:   xrplIssuerAddress,
	}
	memo, err := xrpl.EncodeCoreumRecipientToMemo(coreumSender)
	require.NoError(t, err)

	runnerEnv.SendXRPLPaymentTx(ctx, t, xrplIssuerAddress, runnerEnv.bridgeXRPLAddress, amountToSendFromXRPLtoCoreum, memo)
	runnerEnv.AwaitCoreumBalance(ctx, t, chains.Coreum, coreumSender, sdk.NewCoin(registeredXRPLToken.CoreumDenom, integrationtests.ConvertStringWithDecimalsToSDKInt(t, valueToSendFromXRPLtoCoreum.String(), XRPLTokenDecimals)))

	// send TrustSet to be able to receive coins
	runnerEnv.SendXRPLMaxTrustSetTx(ctx, t, xrplRecipientAddress, xrplIssuerAddress, registeredXRPLCurrency)

	totalSent := sdkmath.ZeroInt()
	amountToSend := integrationtests.ConvertStringWithDecimalsToSDKInt(t, "10", XRPLTokenDecimals)
	for i := 0; i < sendingCount; i++ {
		retryCtx, retryCtxCancel := context.WithTimeout(ctx, 15*time.Second)
		require.NoError(t, retry.Do(retryCtx, 500*time.Millisecond, func() error {
			_, err = runnerEnv.ContractClient.SendToXRPL(ctx, coreumSender, xrplRecipientAddress.String(), sdk.NewCoin(registeredXRPLToken.CoreumDenom, amountToSend))
			if err == nil {
				return nil
			}
			if coreum.IsLastTicketReservedError(err) || coreum.IsNoAvailableTicketsError(err) {
				t.Logf("No tickets left, waiting for new tickets...")
				return retry.Retryable(err)
			}
			require.NoError(t, err)
			return nil
		}))
		retryCtxCancel()
		totalSent = totalSent.Add(amountToSend)
	}
	runnerEnv.AwaitNoPendingOperations(ctx, t)

	balance := runnerEnv.Chains.XRPL.GetAccountBalance(ctx, t, xrplRecipientAddress, xrplIssuerAddress, registeredXRPLCurrency)
	require.Equal(t, totalSent.Quo(sdkmath.NewIntWithDecimal(1, XRPLTokenDecimals)).String(), balance.Value.String())
}

func TestSendXRPLOriginatedTokensFromXRPLToCoreumWithDifferentAmountAndPartialAmount(t *testing.T) {
	t.Parallel()

	ctx, chains := integrationtests.NewTestingContext(t)

	xrplIssuerAddress := chains.XRPL.GenAccount(ctx, t, 100)
	t.Logf("XRPL currency issuer address: %s", xrplIssuerAddress)

	coreumRecipient := chains.Coreum.GenAccount()
	t.Logf("Coreum recipient: %s", coreumRecipient.String())

	sendingPrecision := int32(6)
	maxHoldingAmount := integrationtests.ConvertStringWithDecimalsToSDKInt(t, "1", 30)

	envCfg := DefaultRunnerEnvConfig()
	runnerEnv := NewRunnerEnv(ctx, t, envCfg, chains)

	registeredXRPLCurrency, err := rippledata.NewCurrency("RCP")
	require.NoError(t, err)

	// register token with 20 chars
	currencyHexSymbol := hex.EncodeToString([]byte(strings.Repeat("R", 20)))
	registeredXRPLHexCurrency, err := rippledata.NewCurrency(currencyHexSymbol)
	require.NoError(t, err)

	// fund owner to cover asset FT issuance fees
	chains.Coreum.FundAccountWithOptions(ctx, t, runnerEnv.ContractOwner, coreumintegration.BalancesOptions{
		Amount: chains.Coreum.QueryAssetFTParams(ctx, t).IssueFee.Amount.MulRaw(2),
	})

	// start relayers
	runnerEnv.StartAllRunnerProcesses(ctx, t)
	// recover tickets so we can register tokens
	runnerEnv.AllocateTickets(ctx, t, 200)

	// register XRPL originated token with 3 chars
	_, err = runnerEnv.ContractClient.RegisterXRPLToken(ctx, runnerEnv.ContractOwner, xrplIssuerAddress.String(), xrpl.ConvertCurrencyToString(registeredXRPLCurrency), sendingPrecision, maxHoldingAmount)
	require.NoError(t, err)

	// register XRPL originated token with 20 chars
	_, err = runnerEnv.ContractClient.RegisterXRPLToken(ctx, runnerEnv.ContractOwner, xrplIssuerAddress.String(), xrpl.ConvertCurrencyToString(registeredXRPLHexCurrency), sendingPrecision, maxHoldingAmount)
	require.NoError(t, err)

	// await for the trust set
	runnerEnv.AwaitNoPendingOperations(ctx, t)

	registeredXRPLToken, err := runnerEnv.ContractClient.GetXRPLTokenByIssuerAndCurrency(ctx, xrplIssuerAddress.String(), xrpl.ConvertCurrencyToString(registeredXRPLCurrency))
	require.NoError(t, err)
	require.Equal(t, coreum.TokenStateEnabled, registeredXRPLToken.State)

	registeredXRPLHexCurrencyToken, err := runnerEnv.ContractClient.GetXRPLTokenByIssuerAndCurrency(ctx, xrplIssuerAddress.String(), xrpl.ConvertCurrencyToString(registeredXRPLHexCurrency))
	require.NoError(t, err)
	require.Equal(t, coreum.TokenStateEnabled, registeredXRPLHexCurrencyToken.State)

	lowValue, err := rippledata.NewValue("1.00000111", false)
	require.NoError(t, err)
	maxDecimalsRegisterCurrencyAmount := rippledata.Amount{
		Value:    lowValue,
		Currency: registeredXRPLCurrency,
		Issuer:   xrplIssuerAddress,
	}

	highValue, err := rippledata.NewValue("100000", false)
	require.NoError(t, err)
	highValueRegisteredCurrencyAmount := rippledata.Amount{
		Value:    highValue,
		Currency: registeredXRPLCurrency,
		Issuer:   xrplIssuerAddress,
	}

	normalValue, err := rippledata.NewValue("9.9", false)
	require.NoError(t, err)
	registeredHexCurrencyAmount := rippledata.Amount{
		Value:    normalValue,
		Currency: registeredXRPLHexCurrency,
		Issuer:   xrplIssuerAddress,
	}

	memo, err := xrpl.EncodeCoreumRecipientToMemo(coreumRecipient)
	require.NoError(t, err)

	// incorrect memo
	runnerEnv.SendXRPLPaymentTx(ctx, t, xrplIssuerAddress, runnerEnv.bridgeXRPLAddress, maxDecimalsRegisterCurrencyAmount, rippledata.Memo{})

	// send tx with partial payment
	runnerEnv.SendXRPLPartialPaymentTx(ctx, t, xrplIssuerAddress, runnerEnv.bridgeXRPLAddress, highValueRegisteredCurrencyAmount, maxDecimalsRegisterCurrencyAmount, memo)

	// send tx with high amount
	runnerEnv.SendXRPLPaymentTx(ctx, t, xrplIssuerAddress, runnerEnv.bridgeXRPLAddress, highValueRegisteredCurrencyAmount, memo)

	// send tx with hex currency
	runnerEnv.SendXRPLPaymentTx(ctx, t, xrplIssuerAddress, runnerEnv.bridgeXRPLAddress, registeredHexCurrencyAmount, memo)

	runnerEnv.AwaitCoreumBalance(ctx, t, chains.Coreum, coreumRecipient, sdk.NewCoin(registeredXRPLToken.CoreumDenom, integrationtests.ConvertStringWithDecimalsToSDKInt(t, "100001.000001", XRPLTokenDecimals)))
	runnerEnv.AwaitCoreumBalance(ctx, t, chains.Coreum, coreumRecipient, sdk.NewCoin(registeredXRPLHexCurrencyToken.CoreumDenom, integrationtests.ConvertStringWithDecimalsToSDKInt(t, "9.9", XRPLTokenDecimals)))
}

func TestRecoverXRPLOriginatedTokenRegistrationAndSendFromXRPLToCoreumAndBack(t *testing.T) {
	t.Parallel()

	ctx, chains := integrationtests.NewTestingContext(t)

	envCfg := DefaultRunnerEnvConfig()
	runnerEnv := NewRunnerEnv(ctx, t, envCfg, chains)
	runnerEnv.StartAllRunnerProcesses(ctx, t)
	runnerEnv.AllocateTickets(ctx, t, uint32(200))

	coreumSender := chains.Coreum.GenAccount()
	chains.Coreum.FundAccountWithOptions(ctx, t, coreumSender, coreumintegration.BalancesOptions{
		Amount: sdkmath.NewIntFromUint64(1_000_000),
	})
	t.Logf("Coreum sender: %s", coreumSender.String())
	xrplRecipientAddress := chains.XRPL.GenAccount(ctx, t, 0)
	t.Logf("XRPL recipient: %s", xrplRecipientAddress.String())

	registeredXRPLCurrency, err := rippledata.NewCurrency("RCP")
	require.NoError(t, err)

	// register the XRPL token and await for the tx to be failed
	runnerEnv.Chains.Coreum.FundAccountWithOptions(ctx, t, runnerEnv.ContractOwner, coreumintegration.BalancesOptions{
		Amount: runnerEnv.Chains.Coreum.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})

	// gen account but don't fund it to let the tx to fail since the account won't exist on the XRPL side
	xrplIssuerAddress := chains.XRPL.GenEmptyAccount(t)

	_, err = runnerEnv.ContractClient.RegisterXRPLToken(ctx, runnerEnv.ContractOwner, xrplIssuerAddress.String(), xrpl.ConvertCurrencyToString(registeredXRPLCurrency), int32(6), integrationtests.ConvertStringWithDecimalsToSDKInt(t, "1", 30))
	require.NoError(t, err)
	runnerEnv.AwaitNoPendingOperations(ctx, t)
	registeredXRPLToken, err := runnerEnv.ContractClient.GetXRPLTokenByIssuerAndCurrency(ctx, xrplIssuerAddress.String(), xrpl.ConvertCurrencyToString(registeredXRPLCurrency))
	require.NoError(t, err)
	require.Equal(t, coreum.TokenStateInactive, registeredXRPLToken.State)

	// create the account on the XRPL and send some XRP on top to cover fees
	runnerEnv.Chains.XRPL.CreateAccount(ctx, t, xrplIssuerAddress, 1)
	// recover from owner
	_, err = runnerEnv.ContractClient.RecoverXRPLTokenRegistration(ctx, runnerEnv.ContractOwner, xrplIssuerAddress.String(), registeredXRPLCurrency.String())
	require.NoError(t, err)
	runnerEnv.AwaitNoPendingOperations(ctx, t)
	// now the token is enabled
	registeredXRPLToken, err = runnerEnv.ContractClient.GetXRPLTokenByIssuerAndCurrency(ctx, xrplIssuerAddress.String(), xrpl.ConvertCurrencyToString(registeredXRPLCurrency))
	require.NoError(t, err)
	require.Equal(t, coreum.TokenStateEnabled, registeredXRPLToken.State)

	// enable to be able to send to any address
	runnerEnv.EnableXRPLAccountRippling(ctx, t, xrplIssuerAddress)

	valueToSendFromXRPLtoCoreum, err := rippledata.NewValue("1e10", false)
	require.NoError(t, err)
	amountToSendFromXRPLtoCoreum := rippledata.Amount{
		Value:    valueToSendFromXRPLtoCoreum,
		Currency: registeredXRPLCurrency,
		Issuer:   xrplIssuerAddress,
	}
	memo, err := xrpl.EncodeCoreumRecipientToMemo(coreumSender)
	require.NoError(t, err)

	runnerEnv.SendXRPLPaymentTx(ctx, t, xrplIssuerAddress, runnerEnv.bridgeXRPLAddress, amountToSendFromXRPLtoCoreum, memo)
	runnerEnv.AwaitCoreumBalance(ctx, t, chains.Coreum, coreumSender, sdk.NewCoin(registeredXRPLToken.CoreumDenom, integrationtests.ConvertStringWithDecimalsToSDKInt(t, valueToSendFromXRPLtoCoreum.String(), XRPLTokenDecimals)))

	// send TrustSet to be able to receive coins
	runnerEnv.SendXRPLMaxTrustSetTx(ctx, t, xrplRecipientAddress, xrplIssuerAddress, registeredXRPLCurrency)

	_, err = runnerEnv.ContractClient.SendToXRPL(ctx, coreumSender, xrplRecipientAddress.String(), sdk.NewCoin(registeredXRPLToken.CoreumDenom, integrationtests.ConvertStringWithDecimalsToSDKInt(t, valueToSendFromXRPLtoCoreum.String(), XRPLTokenDecimals)))
	require.NoError(t, err)
	runnerEnv.AwaitNoPendingOperations(ctx, t)

	balance := runnerEnv.Chains.XRPL.GetAccountBalance(ctx, t, xrplRecipientAddress, xrplIssuerAddress, registeredXRPLCurrency)
	require.Equal(t, valueToSendFromXRPLtoCoreum.String(), balance.Value.String())
}

func TestSendCoreumOriginatedTokenFromCoreumToXRPLAndBackWithDifferentAmountsAndPartialAmount(t *testing.T) {
	t.Parallel()

	ctx, chains := integrationtests.NewTestingContext(t)

	xrplRecipientAddress := chains.XRPL.GenAccount(ctx, t, 0)
	t.Logf("XRPL recipient address: %s", xrplRecipientAddress)

	coreumSenderAddress := chains.Coreum.GenAccount()
	issueFee := chains.Coreum.QueryAssetFTParams(ctx, t).IssueFee
	chains.Coreum.FundAccountWithOptions(ctx, t, coreumSenderAddress, coreumintegration.BalancesOptions{
		Amount: issueFee.Amount.Add(sdkmath.NewInt(10_000_000)),
	})

	coreumRecipientAddress := chains.Coreum.GenAccount()
	t.Logf("Coreum recipient: %s", coreumRecipientAddress.String())

	// issue asset ft and register it
	sendingPrecision := int32(2)
	tokenDecimals := uint32(4)
	maxHoldingAmount, ok := sdk.NewIntFromString("10000000000000000")
	require.True(t, ok)
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        coreumSenderAddress.String(),
		Symbol:        "denom",
		Subunit:       "denom",
		Precision:     tokenDecimals, // token decimals in terms of the contract
		InitialAmount: maxHoldingAmount,
	}
	_, err := client.BroadcastTx(
		ctx,
		chains.Coreum.ClientContext.WithFromAddress(coreumSenderAddress),
		chains.Coreum.TxFactory().WithSimulateAndExecute(true),
		issueMsg,
	)
	require.NoError(t, err)

	envCfg := DefaultRunnerEnvConfig()
	runnerEnv := NewRunnerEnv(ctx, t, envCfg, chains)

	// start relayers
	runnerEnv.StartAllRunnerProcesses(ctx, t)
	// recover tickets so we can register tokens
	runnerEnv.AllocateTickets(ctx, t, 200)

	// register Coreum originated token
	require.NoError(t, err)
	denom := assetfttypes.BuildDenom(issueMsg.Subunit, coreumSenderAddress)
	_, err = runnerEnv.ContractClient.RegisterCoreumToken(ctx, runnerEnv.ContractOwner, denom, tokenDecimals, sendingPrecision, maxHoldingAmount)
	require.NoError(t, err)
	registeredCoreumOriginatedToken, err := runnerEnv.ContractClient.GetCoreumTokenByDenom(ctx, denom)
	require.NoError(t, err)

	// send TrustSet to be able to receive coins from the bridge
	xrplCurrency, err := rippledata.NewCurrency(registeredCoreumOriginatedToken.XRPLCurrency)
	require.NoError(t, err)
	runnerEnv.SendXRPLMaxTrustSetTx(ctx, t, xrplRecipientAddress, runnerEnv.bridgeXRPLAddress, xrplCurrency)

	// equal to 11.1111 on XRPL, but with the sending prec 2 we expect 11.11 to be received
	amountToSendToXRPL1 := sdkmath.NewInt(111111)
	// TODO(dzmitryhil) update assertion once we add the final tx revert/recovery
	_, err = runnerEnv.ContractClient.SendToXRPL(ctx, coreumSenderAddress, xrplRecipientAddress.String(), sdk.NewCoin(registeredCoreumOriginatedToken.Denom, amountToSendToXRPL1))
	require.NoError(t, err)

	runnerEnv.AwaitNoPendingOperations(ctx, t)

	// check the XRPL recipient balance
	balance := runnerEnv.Chains.XRPL.GetAccountBalance(ctx, t, xrplRecipientAddress, runnerEnv.bridgeXRPLAddress, xrplCurrency)
	require.Equal(t, "11.11", balance.Value.String())

	amountToSendToXRPL2 := maxHoldingAmount.QuoRaw(2)
	require.NoError(t, err)
	_, err = runnerEnv.ContractClient.SendToXRPL(ctx, coreumSenderAddress, xrplRecipientAddress.String(), sdk.NewCoin(registeredCoreumOriginatedToken.Denom, amountToSendToXRPL2))
	require.NoError(t, err)

	runnerEnv.AwaitNoPendingOperations(ctx, t)

	// check the XRPL recipient balance
	balance = runnerEnv.Chains.XRPL.GetAccountBalance(ctx, t, xrplRecipientAddress, runnerEnv.bridgeXRPLAddress, xrplCurrency)
	require.Equal(t, "50000000001111e-2", balance.Value.String())

	// now start sending from XRPL to coreum, coreum originated token

	// expected to receive `1.11`
	lowValue, err := rippledata.NewValue("1.112233", false)
	require.NoError(t, err)
	currencyAmountWithTruncation := rippledata.Amount{
		Value:    lowValue,
		Currency: xrplCurrency,
		Issuer:   runnerEnv.bridgeXRPLAddress,
	}

	highValue, err := rippledata.NewValue("100000", false)
	require.NoError(t, err)
	highValueRegisteredCurrencyAmount := rippledata.Amount{
		Value:    highValue,
		Currency: xrplCurrency,
		Issuer:   runnerEnv.bridgeXRPLAddress,
	}

	normalValue, err := rippledata.NewValue("9.9", false)
	require.NoError(t, err)
	registeredHexCurrencyAmount := rippledata.Amount{
		Value:    normalValue,
		Currency: xrplCurrency,
		Issuer:   runnerEnv.bridgeXRPLAddress,
	}

	memo, err := xrpl.EncodeCoreumRecipientToMemo(coreumRecipientAddress)
	require.NoError(t, err)

	// send tx with partial payment
	runnerEnv.SendXRPLPartialPaymentTx(ctx, t, xrplRecipientAddress, runnerEnv.bridgeXRPLAddress, highValueRegisteredCurrencyAmount, currencyAmountWithTruncation, memo)

	// send tx with high amount
	runnerEnv.SendXRPLPaymentTx(ctx, t, xrplRecipientAddress, runnerEnv.bridgeXRPLAddress, highValueRegisteredCurrencyAmount, memo)

	// send tx with hex currency
	runnerEnv.SendXRPLPaymentTx(ctx, t, xrplRecipientAddress, runnerEnv.bridgeXRPLAddress, registeredHexCurrencyAmount, memo)

	runnerEnv.AwaitCoreumBalance(ctx, t, chains.Coreum, coreumRecipientAddress, sdk.NewCoin(registeredCoreumOriginatedToken.Denom, integrationtests.ConvertStringWithDecimalsToSDKInt(t, "100011.01", int64(tokenDecimals))))
}

func TestSendCoreumOriginatedTokenFromCoreumToXRPLAndBackWithMaliciousRelayer(t *testing.T) {
	t.Parallel()

	ctx, chains := integrationtests.NewTestingContext(t)

	xrplRecipientAddress := chains.XRPL.GenAccount(ctx, t, 0)
	t.Logf("XRPL recipient address: %s", xrplRecipientAddress)

	coreumSenderAddress := chains.Coreum.GenAccount()
	issueFee := chains.Coreum.QueryAssetFTParams(ctx, t).IssueFee
	chains.Coreum.FundAccountWithOptions(ctx, t, coreumSenderAddress, coreumintegration.BalancesOptions{
		Amount: issueFee.Amount.Add(sdkmath.NewInt(10_000_000)),
	})

	coreumRecipientAddress := chains.Coreum.GenAccount()
	t.Logf("Coreum recipient: %s", coreumRecipientAddress.String())

	// issue asset ft and register it
	sendingPrecision := int32(4)
	tokenDecimals := uint32(10)
	maxHoldingAmount, ok := sdk.NewIntFromString("10000000000000000")
	require.True(t, ok)
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        coreumSenderAddress.String(),
		Symbol:        "denom",
		Subunit:       "denom",
		Precision:     tokenDecimals, // token decimals in terms of the contract
		InitialAmount: maxHoldingAmount,
	}
	_, err := client.BroadcastTx(
		ctx,
		chains.Coreum.ClientContext.WithFromAddress(coreumSenderAddress),
		chains.Coreum.TxFactory().WithSimulateAndExecute(true),
		issueMsg,
	)
	require.NoError(t, err)

	envCfg := DefaultRunnerEnvConfig()
	envCfg.MaliciousRelayerNumber = 1
	runnerEnv := NewRunnerEnv(ctx, t, envCfg, chains)

	// start relayers
	runnerEnv.StartAllRunnerProcesses(ctx, t)
	// recover tickets so we can register tokens
	runnerEnv.AllocateTickets(ctx, t, 200)

	// register Coreum originated token
	require.NoError(t, err)
	denom := assetfttypes.BuildDenom(issueMsg.Subunit, coreumSenderAddress)
	_, err = runnerEnv.ContractClient.RegisterCoreumToken(ctx, runnerEnv.ContractOwner, denom, tokenDecimals, sendingPrecision, maxHoldingAmount)
	require.NoError(t, err)
	registeredCoreumOriginatedToken, err := runnerEnv.ContractClient.GetCoreumTokenByDenom(ctx, denom)
	require.NoError(t, err)

	// send TrustSet to be able to receive coins from the bridge
	xrplCurrency, err := rippledata.NewCurrency(registeredCoreumOriginatedToken.XRPLCurrency)
	require.NoError(t, err)
	runnerEnv.SendXRPLMaxTrustSetTx(ctx, t, xrplRecipientAddress, runnerEnv.bridgeXRPLAddress, xrplCurrency)

	amountToSendToXRPL := sdkmath.NewInt(9999999)
	_, err = runnerEnv.ContractClient.SendToXRPL(ctx, coreumSenderAddress, xrplRecipientAddress.String(), sdk.NewCoin(registeredCoreumOriginatedToken.Denom, amountToSendToXRPL))
	require.NoError(t, err)

	runnerEnv.AwaitNoPendingOperations(ctx, t)

	// check the XRPL recipient balance
	balance := runnerEnv.Chains.XRPL.GetAccountBalance(ctx, t, xrplRecipientAddress, runnerEnv.bridgeXRPLAddress, xrplCurrency)
	require.Equal(t, "0.0009", balance.Value.String())

	// send the full amount back
	memo, err := xrpl.EncodeCoreumRecipientToMemo(coreumRecipientAddress)
	require.NoError(t, err)
	runnerEnv.SendXRPLPaymentTx(ctx, t, xrplRecipientAddress, runnerEnv.bridgeXRPLAddress, balance, memo)
	runnerEnv.AwaitCoreumBalance(ctx, t, chains.Coreum, coreumRecipientAddress, sdk.NewCoin(registeredCoreumOriginatedToken.Denom, sdk.NewInt(9000000)))
}