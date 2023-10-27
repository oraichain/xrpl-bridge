//go:build integrationtests
// +build integrationtests

package processes_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreumbridge-xrpl/integration-tests"
)

// TODO(dzmitryhil) add the additional test for the re-allocation of the tickets without provided number once we have more operations

func TestTicketsAllocationRecoveryWithSequenceNumber(t *testing.T) {
	t.Parallel()

	numberOfTicketsToAllocate := uint32(250)
	ctx, chains := integrationtests.NewTestingContext(t)

	runnerEnv := NewRunnerEnv(ctx, t, DefaultRunnerEnvConfig(), chains)
	availableTickets, err := runnerEnv.ContractClient.GetAvailableTickets(ctx)
	require.NoError(t, err)
	require.Empty(t, availableTickets)

	runnerEnv.StartAllRunnerProcesses(ctx, t)
	chains.XRPL.FundAccountForTicketAllocation(ctx, t, runnerEnv.XRPLBridgeAccount, numberOfTicketsToAllocate)

	xrplBridgeAccountInfo, err := chains.XRPL.RPCClient().AccountInfo(ctx, runnerEnv.XRPLBridgeAccount)
	require.NoError(t, err)

	_, err = runnerEnv.ContractClient.RecoverTickets(
		ctx,
		runnerEnv.ContractOwner,
		*xrplBridgeAccountInfo.AccountData.Sequence,
		&numberOfTicketsToAllocate,
	)
	require.NoError(t, err)

	runnerEnv.AwaitNoPendingOperations(ctx, t)

	availableTickets, err = runnerEnv.ContractClient.GetAvailableTickets(ctx)
	require.NoError(t, err)
	require.Len(t, availableTickets, int(numberOfTicketsToAllocate))
}

func TestTicketsAllocationRecoveryWithRejection(t *testing.T) {
	t.Parallel()

	numberOfTicketsToAllocate := uint32(250)
	ctx, chains := integrationtests.NewTestingContext(t)

	runnerEnv := NewRunnerEnv(ctx, t, DefaultRunnerEnvConfig(), chains)
	availableTickets, err := runnerEnv.ContractClient.GetAvailableTickets(ctx)
	require.NoError(t, err)
	require.Empty(t, availableTickets)

	runnerEnv.StartAllRunnerProcesses(ctx, t)
	// we don't fund the contract for the tickets allocation to let the chain reject the allocation transaction

	xrplBridgeAccountInfo, err := chains.XRPL.RPCClient().AccountInfo(ctx, runnerEnv.XRPLBridgeAccount)
	require.NoError(t, err)

	// we don't have enough balance on the contract so the recovery will be rejected
	_, err = runnerEnv.ContractClient.RecoverTickets(
		ctx,
		runnerEnv.ContractOwner,
		*xrplBridgeAccountInfo.AccountData.Sequence,
		&numberOfTicketsToAllocate,
	)
	require.NoError(t, err)

	runnerEnv.AwaitNoPendingOperations(ctx, t)

	availableTickets, err = runnerEnv.ContractClient.GetAvailableTickets(ctx)
	require.NoError(t, err)
	require.Empty(t, availableTickets)
}

func TestTicketsAllocationRecoveryWithMaliciousRelayers(t *testing.T) {
	t.Parallel()

	numberOfTicketsToAllocate := uint32(200)
	ctx, chains := integrationtests.NewTestingContext(t)

	envCfg := DefaultRunnerEnvConfig()
	// add malicious relayers to the config
	envCfg.RelayerNumber = 5
	envCfg.MaliciousRelayerNumber = 2
	envCfg.SigningThreshold = 3

	runnerEnv := NewRunnerEnv(ctx, t, envCfg, chains)
	availableTickets, err := runnerEnv.ContractClient.GetAvailableTickets(ctx)
	require.NoError(t, err)
	require.Empty(t, availableTickets)

	runnerEnv.StartAllRunnerProcesses(ctx, t)

	chains.XRPL.FundAccountForTicketAllocation(ctx, t, runnerEnv.XRPLBridgeAccount, numberOfTicketsToAllocate)

	xrplBridgeAccountInfo, err := chains.XRPL.RPCClient().AccountInfo(ctx, runnerEnv.XRPLBridgeAccount)
	require.NoError(t, err)

	_, err = runnerEnv.ContractClient.RecoverTickets(
		ctx,
		runnerEnv.ContractOwner,
		*xrplBridgeAccountInfo.AccountData.Sequence,
		&numberOfTicketsToAllocate,
	)
	require.NoError(t, err)

	runnerEnv.AwaitNoPendingOperations(ctx, t)

	availableTickets, err = runnerEnv.ContractClient.GetAvailableTickets(ctx)
	require.NoError(t, err)
	require.Len(t, availableTickets, int(numberOfTicketsToAllocate))
}