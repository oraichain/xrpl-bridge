//go:build integrationtests
// +build integrationtests

package processes_test

import (
	"context"
	"sync"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	rippledata "github.com/rubblelabs/ripple/data"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum-tools/pkg/retry"
	coreumapp "github.com/CoreumFoundation/coreum/v3/app"
	coreumconfig "github.com/CoreumFoundation/coreum/v3/pkg/config"
	coreumintegration "github.com/CoreumFoundation/coreum/v3/testutil/integration"
	integrationtests "github.com/CoreumFoundation/coreumbridge-xrpl/integration-tests"
	"github.com/CoreumFoundation/coreumbridge-xrpl/relayer/coreum"
	"github.com/CoreumFoundation/coreumbridge-xrpl/relayer/runner"
)

// RunnerEnvConfig is runner environment config.
type RunnerEnvConfig struct {
	AwaitTimeout           time.Duration
	SigningThreshold       int
	RelayerNumber          int
	MaliciousRelayerNumber int
	DisableMasterKey       bool
	UsedTicketsThreshold   int
}

// DefaultRunnerEnvConfig returns default runner environment config.
func DefaultRunnerEnvConfig() RunnerEnvConfig {
	return RunnerEnvConfig{
		AwaitTimeout:           15 * time.Second,
		SigningThreshold:       2,
		RelayerNumber:          3,
		MaliciousRelayerNumber: 0,
		DisableMasterKey:       true,
		UsedTicketsThreshold:   150,
	}
}

// RunnerEnv is runner environment used for the integration tests.
type RunnerEnv struct {
	Cfg               RunnerEnvConfig
	XRPLBridgeAccount rippledata.Account
	ContractClient    *coreum.ContractClient
	ContractOwner     sdk.AccAddress
	Runners           []*runner.Runner
	ProcessErrorsMu   sync.RWMutex
	ProcessErrors     []error
}

// NewRunnerEnv returns new instance of the RunnerEnv.
func NewRunnerEnv(ctx context.Context, t *testing.T, cfg RunnerEnvConfig, chains integrationtests.Chains) *RunnerEnv {
	coreumRelayerAddresses := genCoreumRelayers(
		ctx,
		t,
		chains.Coreum,
		cfg.RelayerNumber,
	)
	xrplBridgeAccount, xrplRelayerAccounts, xrplRelayersPubKeys := genXRPLBridgeAccountWithRelayers(
		ctx,
		t,
		chains.XRPL,
		cfg.RelayerNumber,
		uint32(cfg.SigningThreshold),
		cfg.DisableMasterKey,
	)

	contractRelayer := make([]coreum.Relayer, 0, cfg.RelayerNumber)
	for i := 0; i < cfg.RelayerNumber; i++ {
		contractRelayer = append(contractRelayer, coreum.Relayer{
			CoreumAddress: coreumRelayerAddresses[i],
			XRPLAddress:   xrplRelayerAccounts[i].String(),
			XRPLPubKey:    xrplRelayersPubKeys[i].String(),
		})
	}

	contractOwner, contractClient := integrationtests.DeployAndInstantiateContract(
		ctx,
		t,
		chains,
		contractRelayer,
		cfg.SigningThreshold,
		cfg.UsedTicketsThreshold,
	)

	runners := make([]*runner.Runner, 0, cfg.RelayerNumber)
	// add correct relayers
	for i := 0; i < cfg.RelayerNumber-cfg.MaliciousRelayerNumber; i++ {
		runners = append(
			runners,
			createDevRunner(
				t,
				chains,
				xrplBridgeAccount,
				xrplRelayerAccounts[i],
				contractClient.GetContractAddress(),
				coreumRelayerAddresses[i],
			),
		)
	}
	// add malicious relayers
	// we keep the relayer indexes to make all config valid apart from the XRPL signing
	for i := cfg.RelayerNumber - cfg.MaliciousRelayerNumber; i < cfg.RelayerNumber; i++ {
		maliciousSignerAcc := chains.XRPL.GenAccount(ctx, t, 10)
		runners = append(
			runners,
			createDevRunner(
				t,
				chains,
				xrplBridgeAccount,
				maliciousSignerAcc,
				contractClient.GetContractAddress(),
				coreumRelayerAddresses[i],
			),
		)
	}

	runnerEnv := &RunnerEnv{
		Cfg:               cfg,
		XRPLBridgeAccount: xrplBridgeAccount,
		ContractClient:    contractClient,
		ContractOwner:     contractOwner,
		Runners:           runners,
		ProcessErrorsMu:   sync.RWMutex{},
		ProcessErrors:     make([]error, 0),
	}
	t.Cleanup(func() {
		runnerEnv.RequireNoErrors(t)
	})

	return runnerEnv
}

// StartAllRunnerProcesses starts all relayer processes.
func (r *RunnerEnv) StartAllRunnerProcesses(ctx context.Context, t *testing.T) {
	errCh := make(chan error, len(r.Runners))
	go func() {
		for {
			select {
			case <-ctx.Done():
				if !errors.Is(ctx.Err(), context.Canceled) {
					r.ProcessErrorsMu.Lock()
					r.ProcessErrors = append(r.ProcessErrors, ctx.Err())
					r.ProcessErrorsMu.Unlock()
				}
				return
			case err := <-errCh:
				r.ProcessErrorsMu.Lock()
				r.ProcessErrors = append(r.ProcessErrors, err)
				r.ProcessErrorsMu.Unlock()
			}
		}
	}()

	for _, relayerRunner := range r.Runners {
		go func(relayerRunner *runner.Runner) {
			// disable restart on error to handler unexpected errors
			xrplTxObserverProcess := relayerRunner.Processes.XRPLTxObserver
			xrplTxObserverProcess.IsRestartableOnError = false
			xrplTxSubmitterProcess := relayerRunner.Processes.XRPLTxSubmitter
			xrplTxSubmitterProcess.IsRestartableOnError = false

			err := relayerRunner.Processor.StartProcesses(ctx, xrplTxObserverProcess, xrplTxSubmitterProcess)
			if err != nil && !errors.Is(err, context.Canceled) {
				t.Logf("Unexpected error on process start:%s", err)
				errCh <- err
			}
		}(relayerRunner)
	}
}

// AwaitNoPendingOperations waits for no pendoing contract transactions.
func (r *RunnerEnv) AwaitNoPendingOperations(ctx context.Context, t *testing.T) {
	t.Helper()

	r.AwaitState(ctx, t, func(t *testing.T) error {
		operations, err := r.ContractClient.GetPendingOperations(ctx)
		require.NoError(t, err)
		if len(operations) != 0 {
			return errors.Errorf("there are still pending operatrions: %+v", operations)
		}
		return nil
	})
}

// AwaitCoreumBalance waits for expected coreum balance.
func (r *RunnerEnv) AwaitCoreumBalance(
	ctx context.Context,
	t *testing.T,
	coreumChain integrationtests.CoreumChain,
	address sdk.AccAddress,
	expectedBalance sdk.Coin,
) {
	t.Helper()
	awaitContext, awaitContextCancel := context.WithTimeout(ctx, r.Cfg.AwaitTimeout)
	t.Cleanup(awaitContextCancel)
	require.NoError(t, coreumChain.AwaitForBalance(awaitContext, t, address, expectedBalance))
}

// AwaitState waits for stateChecker function to rerun nil and retires in case of failure.
func (r *RunnerEnv) AwaitState(ctx context.Context, t *testing.T, stateChecker func(t *testing.T) error) {
	t.Helper()
	retryCtx, retryCancel := context.WithTimeout(ctx, r.Cfg.AwaitTimeout)
	defer retryCancel()
	err := retry.Do(retryCtx, 500*time.Millisecond, func() error {
		if err := stateChecker(t); err != nil {
			return retry.Retryable(err)
		}

		return nil
	})
	require.NoError(t, err)
}

func (r *RunnerEnv) RequireNoErrors(t *testing.T) {
	r.ProcessErrorsMu.RLock()
	defer r.ProcessErrorsMu.RUnlock()
	require.Empty(t, r.ProcessErrors, "Found unexpected process errors after the execution")
}

// SendTrustSet sends TrustSet transaction.
func SendTrustSet(
	ctx context.Context,
	t *testing.T,
	xrplChain integrationtests.XRPLChain,
	issuer, sender rippledata.Account,
	currency rippledata.Currency,
) {
	trustSetValue, err := rippledata.NewValue("10e20", false)
	require.NoError(t, err)
	senderCurrencyTrustSetTx := rippledata.TrustSet{
		LimitAmount: rippledata.Amount{
			Value:    trustSetValue,
			Currency: currency,
			Issuer:   issuer,
		},
		TxBase: rippledata.TxBase{
			TransactionType: rippledata.TRUST_SET,
		},
	}
	require.NoError(t, xrplChain.AutoFillSignAndSubmitTx(ctx, t, &senderCurrencyTrustSetTx, sender))
}

// SendXRPLPaymentTx sends Payment transaction.
func SendXRPLPaymentTx(
	ctx context.Context,
	t *testing.T,
	xrplChain integrationtests.XRPLChain,
	senderAcc, recipientAcc rippledata.Account,
	amount rippledata.Amount,
	memo rippledata.Memo,
) {
	xrpPaymentTx := rippledata.Payment{
		Destination: recipientAcc,
		Amount:      amount,
		TxBase: rippledata.TxBase{
			TransactionType: rippledata.PAYMENT,
			Memos: rippledata.Memos{
				memo,
			},
		},
	}
	require.NoError(t, xrplChain.AutoFillSignAndSubmitTx(ctx, t, &xrpPaymentTx, senderAcc))
}

// SendXRPLPartialPaymentTx sends Payment transaction with partial payment.
func SendXRPLPartialPaymentTx(
	ctx context.Context,
	t *testing.T,
	xrplChain integrationtests.XRPLChain,
	senderAcc, recipientAcc rippledata.Account,
	amount rippledata.Amount,
	maxAmount rippledata.Amount,
	memo rippledata.Memo,
) {
	xrpPaymentTx := rippledata.Payment{
		Destination: recipientAcc,
		Amount:      amount,
		SendMax:     &maxAmount,
		TxBase: rippledata.TxBase{
			TransactionType: rippledata.PAYMENT,
			Memos: rippledata.Memos{
				memo,
			},
			Flags: lo.ToPtr(rippledata.TxPartialPayment),
		},
	}
	require.NoError(t, xrplChain.AutoFillSignAndSubmitTx(ctx, t, &xrpPaymentTx, senderAcc))
}

func genCoreumRelayers(
	ctx context.Context,
	t *testing.T,
	coreumChain integrationtests.CoreumChain,
	relayersCount int,
) []sdk.AccAddress {
	t.Helper()

	addresses := make([]sdk.AccAddress, 0, relayersCount)
	for i := 0; i < relayersCount; i++ {
		relayerAddress := coreumChain.GenAccount()
		coreumChain.FundAccountWithOptions(ctx, t, relayerAddress, coreumintegration.BalancesOptions{
			Amount: sdkmath.NewIntFromUint64(1_000_000),
		})
		addresses = append(addresses, relayerAddress)
	}

	return addresses
}

func genXRPLBridgeAccountWithRelayers(
	ctx context.Context,
	t *testing.T,
	xrplChain integrationtests.XRPLChain,
	signersCount int,
	signerQuorum uint32,
	disableMasterKey bool,
) (rippledata.Account, []rippledata.Account, []rippledata.PublicKey) {
	t.Helper()

	bridgeAcc := xrplChain.GenAccount(ctx, t, 10)
	t.Logf("Bridge account is generated, address:%s", bridgeAcc.String())
	signerEntries := make([]rippledata.SignerEntry, 0, signersCount)
	signerAccounts := make([]rippledata.Account, 0, signersCount)
	signerPubKeys := make([]rippledata.PublicKey, 0, signersCount)
	for i := 0; i < signersCount; i++ {
		signerAcc := xrplChain.GenAccount(ctx, t, 10)
		signerAccounts = append(signerAccounts, signerAcc)
		t.Logf("Signer %d is generated, address:%s", i+1, signerAcc.String())
		signerEntries = append(signerEntries, rippledata.SignerEntry{
			SignerEntry: rippledata.SignerEntryItem{
				Account:      &signerAcc,
				SignerWeight: lo.ToPtr(uint16(1)),
			},
		})
		signerPubKeys = append(signerPubKeys, xrplChain.GetSignerPubKey(t, signerAcc))
	}

	signerListSetTx := rippledata.SignerListSet{
		SignerQuorum:  signerQuorum,
		SignerEntries: signerEntries,
		TxBase: rippledata.TxBase{
			TransactionType: rippledata.SIGNER_LIST_SET,
		},
	}
	require.NoError(t, xrplChain.AutoFillSignAndSubmitTx(ctx, t, &signerListSetTx, bridgeAcc))
	t.Logf("The bridge signers set is updated")

	if disableMasterKey {
		// disable master key
		disableMasterKeyTx := rippledata.AccountSet{
			TxBase: rippledata.TxBase{
				Account:         bridgeAcc,
				TransactionType: rippledata.ACCOUNT_SET,
			},
			SetFlag: lo.ToPtr(uint32(rippledata.TxSetDisableMaster)),
		}
		require.NoError(t, xrplChain.AutoFillSignAndSubmitTx(ctx, t, &disableMasterKeyTx, bridgeAcc))
		t.Logf("Bridge account master key is disabled")
	}

	return bridgeAcc, signerAccounts, signerPubKeys
}

func createDevRunner(
	t *testing.T,
	chains integrationtests.Chains,
	xrplBridgeAccount rippledata.Account,
	xrplRelayerAcc rippledata.Account,
	contractAddress sdk.AccAddress,
	coreumRelayerAddress sdk.AccAddress,
) *runner.Runner {
	t.Helper()

	const (
		coreumRelayerKeyName = "coreum"
		xrplRelayerKeyName   = "xrpl"
	)

	encodingConfig := coreumconfig.NewEncodingConfig(coreumapp.ModuleBasics)
	kr := keyring.NewInMemory(encodingConfig.Codec)

	// reimport coreum key
	coreumKr := chains.Coreum.ClientContext.Keyring()
	keyInfo, err := coreumKr.KeyByAddress(coreumRelayerAddress)
	require.NoError(t, err)
	pass := uuid.NewString()
	armor, err := coreumKr.ExportPrivKeyArmor(keyInfo.Name, pass)
	require.NoError(t, err)
	require.NoError(t, kr.ImportPrivKey(coreumRelayerKeyName, armor, pass))

	// reimport XRPL key
	xrplKr := chains.XRPL.GetSignerKeyring()
	keyInfo, err = xrplKr.Key(xrplRelayerAcc.String())
	require.NoError(t, err)
	armor, err = xrplKr.ExportPrivKeyArmor(keyInfo.Name, pass)
	require.NoError(t, err)
	require.NoError(t, kr.ImportPrivKey(xrplRelayerKeyName, armor, pass))

	relayerRunnerCfg := runner.DefaultConfig()
	relayerRunnerCfg.LoggingConfig.Level = "debug"

	relayerRunnerCfg.XRPL.BridgeAccount = xrplBridgeAccount.String()
	relayerRunnerCfg.XRPL.MultiSignerKeyName = xrplRelayerKeyName
	relayerRunnerCfg.XRPL.RPC.URL = chains.XRPL.Config().RPCAddress
	// make the scanner fast
	relayerRunnerCfg.XRPL.Scanner.RetryDelay = 500 * time.Millisecond

	relayerRunnerCfg.Coreum.GRPC.URL = chains.Coreum.Config().GRPCAddress
	relayerRunnerCfg.Coreum.RelayerKeyName = coreumRelayerKeyName
	relayerRunnerCfg.Coreum.Contract.ContractAddress = contractAddress.String()
	// We use high gas adjustment since our relayers might send transactions in one block.
	// They estimate gas based on the same state, but since transactions are executed one by one the next transaction uses
	// the state different from the one it used for the estimation as a result the out-of-gas error might appear.
	relayerRunnerCfg.Coreum.Contract.GasAdjustment = 2
	relayerRunnerCfg.Coreum.Network.ChainID = chains.Coreum.ChainSettings.ChainID
	// make operation fetcher fast
	relayerRunnerCfg.Processes.XRPLTxSubmitter.RepeatDelay = 500 * time.Millisecond

	relayerRunner, err := runner.NewRunner(relayerRunnerCfg, kr)
	require.NoError(t, err)
	return relayerRunner
}