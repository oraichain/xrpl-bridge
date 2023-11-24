package processes_test

import (
	"context"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/golang/mock/gomock"
	rippledata "github.com/rubblelabs/ripple/data"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreumbridge-xrpl/relayer/coreum"
	"github.com/CoreumFoundation/coreumbridge-xrpl/relayer/logger"
	"github.com/CoreumFoundation/coreumbridge-xrpl/relayer/processes"
	"github.com/CoreumFoundation/coreumbridge-xrpl/relayer/xrpl"
)

func TestXRPLTxSubmitter_Start(t *testing.T) {
	// intentionally not parallel since uses the common data
	bridgeXRPLAddress := xrpl.GenPrivKeyTxSigner().Account()
	xrplTxSignerKeyName := "xrpl-tx-signer"

	contractRelayers, xrplTxSigners, bridgeXRPLSignerAccountWithSigners := genContractRelayers(3)

	// ********** AllocateTickets **********

	allocateTicketsOperation,
		allocateTicketOperationWithUnexpectedSeqNumber,
		allocateTicketOperationWithSignatures,
		allocateTicketOperationValidSigners := buildAllocateTicketsTestData(t, xrplTxSigners, bridgeXRPLAddress, contractRelayers)

	// ********** TrustSet **********

	trustSetOperation,
		trustSetOperationWithSignatures,
		trustSetOperationValidSigners := buildTrustSetTestData(t, xrplTxSigners, bridgeXRPLAddress, contractRelayers)

	// ********** CoreumToXRPLTransfer **********

	// XRPLOriginated

	coreumToXRPLXRPLOriginatedTokenTransferOperation,
		coreumToXRPLXRPLOriginatedTokenTransferOperationWithSignatures,
		coreumToXRPLXRPLOriginatedTokenTransferOperationValidSigners := buildCoreumToXRPLXRPLOriginatedTokenTransferTestData(t, xrplTxSigners, bridgeXRPLAddress, contractRelayers)

	// CoreumOriginated
	coreumOriginatedTokenDecimals := uint32(6)
	coreumToXRPLCoreumOriginatedTokenTransferOperation,
		coreumToXRPLCoreumOriginatedTokenTransferOperationWithSignatures,
		coreumToXRPLCoreumOriginatedTokenTransferOperationValidSigners := buildCoreumToXRPLCoreumOriginatedTokenTransferTestData(t, xrplTxSigners, bridgeXRPLAddress, contractRelayers, coreumOriginatedTokenDecimals)

	tests := []struct {
		name                  string
		contractClientBuilder func(ctrl *gomock.Controller) processes.ContractClient
		xrplRPCClientBuilder  func(ctrl *gomock.Controller) processes.XRPLRPCClient
		xrplTxSignerBuilder   func(ctrl *gomock.Controller) processes.XRPLTxSigner
	}{
		{
			name: "no_pending_operations",
			contractClientBuilder: func(ctrl *gomock.Controller) processes.ContractClient {
				contractClientMock := NewMockContractClient(ctrl)
				contractClientMock.EXPECT().GetPendingOperations(gomock.Any()).Return([]coreum.Operation{}, nil)
				return contractClientMock
			},
		},
		{
			name: "register_signature_for_create_ticket_tx",
			contractClientBuilder: func(ctrl *gomock.Controller) processes.ContractClient {
				contractClientMock := NewMockContractClient(ctrl)
				contractClientMock.EXPECT().GetPendingOperations(gomock.Any()).Return([]coreum.Operation{allocateTicketsOperation}, nil)
				contractClientMock.EXPECT().GetContractConfig(gomock.Any()).Return(coreum.ContractConfig{
					Relayers: contractRelayers,
				}, nil)
				contractClientMock.EXPECT().SaveSignature(gomock.Any(), contractRelayers[0].CoreumAddress, allocateTicketsOperation.AccountSequence, allocateTicketOperationValidSigners[0].Signer.TxnSignature.String())
				return contractClientMock
			},
			xrplRPCClientBuilder: func(ctrl *gomock.Controller) processes.XRPLRPCClient {
				xrplRPCClientMock := NewMockXRPLRPCClient(ctrl)
				// 2 times one for the signatures and one more for the seq number
				xrplRPCClientMock.EXPECT().AccountInfo(gomock.Any(), bridgeXRPLAddress).Return(bridgeXRPLSignerAccountWithSigners, nil).Times(2)
				return xrplRPCClientMock
			},
			xrplTxSignerBuilder: func(ctrl *gomock.Controller) processes.XRPLTxSigner {
				xrplTxSignerMock := NewMockXRPLTxSigner(ctrl)
				tx, err := processes.BuildTicketCreateTxForMultiSigning(bridgeXRPLAddress, allocateTicketsOperation)
				require.NoError(t, err)
				xrplTxSignerMock.EXPECT().MultiSign(tx, xrplTxSignerKeyName).Return(allocateTicketOperationValidSigners[0], nil)

				return xrplTxSignerMock
			},
		},
		{
			name: "submit_create_ticket_tx_with_filtered_signature",
			contractClientBuilder: func(ctrl *gomock.Controller) processes.ContractClient {
				contractClientMock := NewMockContractClient(ctrl)
				contractClientMock.EXPECT().GetPendingOperations(gomock.Any()).Return([]coreum.Operation{allocateTicketOperationWithSignatures}, nil)
				contractClientMock.EXPECT().GetContractConfig(gomock.Any()).Return(coreum.ContractConfig{
					Relayers: contractRelayers,
				}, nil)
				return contractClientMock
			},
			xrplRPCClientBuilder: func(ctrl *gomock.Controller) processes.XRPLRPCClient {
				xrplRPCClientMock := NewMockXRPLRPCClient(ctrl)
				xrplRPCClientMock.EXPECT().AccountInfo(gomock.Any(), bridgeXRPLAddress).Return(bridgeXRPLSignerAccountWithSigners, nil)
				expectedTx, err := processes.BuildTicketCreateTxForMultiSigning(bridgeXRPLAddress, allocateTicketOperationWithSignatures)
				require.NoError(t, err)
				require.NoError(t, rippledata.SetSigners(expectedTx, allocateTicketOperationValidSigners...))
				xrplRPCClientMock.EXPECT().Submit(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, tx rippledata.Transaction) (xrpl.SubmitResult, error) {
					_, expectedTxRaw, err := rippledata.Raw(expectedTx)
					require.NoError(t, err)
					_, txRaw, err := rippledata.Raw(tx)
					require.NoError(t, err)
					require.Equal(t, expectedTxRaw, txRaw)
					return xrpl.SubmitResult{}, nil
				})

				return xrplRPCClientMock
			},
		},
		{
			name: "register_invalid_create_ticket_tx",
			contractClientBuilder: func(ctrl *gomock.Controller) processes.ContractClient {
				contractClientMock := NewMockContractClient(ctrl)
				contractClientMock.EXPECT().GetPendingOperations(gomock.Any()).Return([]coreum.Operation{allocateTicketOperationWithUnexpectedSeqNumber}, nil)
				contractClientMock.EXPECT().GetContractConfig(gomock.Any()).Return(coreum.ContractConfig{
					Relayers: contractRelayers,
				}, nil)
				contractClientMock.EXPECT().SendXRPLTicketsAllocationTransactionResultEvidence(gomock.Any(), contractRelayers[0].CoreumAddress, coreum.XRPLTransactionResultTicketsAllocationEvidence{
					XRPLTransactionResultEvidence: coreum.XRPLTransactionResultEvidence{
						AccountSequence:   &allocateTicketOperationWithUnexpectedSeqNumber.AccountSequence,
						TransactionResult: coreum.TransactionResultInvalid,
					},
				})
				return contractClientMock
			},
			xrplRPCClientBuilder: func(ctrl *gomock.Controller) processes.XRPLRPCClient {
				xrplRPCClientMock := NewMockXRPLRPCClient(ctrl)
				// 2 times one for the signatures and one more for the seq number
				xrplRPCClientMock.EXPECT().AccountInfo(gomock.Any(), bridgeXRPLAddress).Return(bridgeXRPLSignerAccountWithSigners, nil).Times(2)
				return xrplRPCClientMock
			},
		},
		{
			name: "register_signature_for_trust_set_tx",
			contractClientBuilder: func(ctrl *gomock.Controller) processes.ContractClient {
				contractClientMock := NewMockContractClient(ctrl)
				contractClientMock.EXPECT().GetPendingOperations(gomock.Any()).Return([]coreum.Operation{trustSetOperation}, nil)
				contractClientMock.EXPECT().GetContractConfig(gomock.Any()).Return(coreum.ContractConfig{
					Relayers: contractRelayers,
				}, nil)
				contractClientMock.EXPECT().SaveSignature(gomock.Any(), contractRelayers[0].CoreumAddress, trustSetOperation.TicketSequence, trustSetOperationValidSigners[0].Signer.TxnSignature.String())
				return contractClientMock
			},
			xrplRPCClientBuilder: func(ctrl *gomock.Controller) processes.XRPLRPCClient {
				xrplRPCClientMock := NewMockXRPLRPCClient(ctrl)
				xrplRPCClientMock.EXPECT().AccountInfo(gomock.Any(), bridgeXRPLAddress).Return(bridgeXRPLSignerAccountWithSigners, nil)
				return xrplRPCClientMock
			},
			xrplTxSignerBuilder: func(ctrl *gomock.Controller) processes.XRPLTxSigner {
				xrplTxSignerMock := NewMockXRPLTxSigner(ctrl)
				tx, err := processes.BuildTrustSetTxForMultiSigning(bridgeXRPLAddress, trustSetOperation)
				require.NoError(t, err)
				xrplTxSignerMock.EXPECT().MultiSign(tx, xrplTxSignerKeyName).Return(trustSetOperationValidSigners[0], nil)

				return xrplTxSignerMock
			},
		},
		{
			name: "submit_trust_set_tx_with_filtered_signature",
			contractClientBuilder: func(ctrl *gomock.Controller) processes.ContractClient {
				contractClientMock := NewMockContractClient(ctrl)
				contractClientMock.EXPECT().GetPendingOperations(gomock.Any()).Return([]coreum.Operation{trustSetOperationWithSignatures}, nil)
				contractClientMock.EXPECT().GetContractConfig(gomock.Any()).Return(coreum.ContractConfig{
					Relayers: contractRelayers,
				}, nil)
				return contractClientMock
			},
			xrplRPCClientBuilder: func(ctrl *gomock.Controller) processes.XRPLRPCClient {
				xrplRPCClientMock := NewMockXRPLRPCClient(ctrl)
				xrplRPCClientMock.EXPECT().AccountInfo(gomock.Any(), bridgeXRPLAddress).Return(bridgeXRPLSignerAccountWithSigners, nil)
				expectedTx, err := processes.BuildTrustSetTxForMultiSigning(bridgeXRPLAddress, trustSetOperationWithSignatures)
				require.NoError(t, err)
				require.NoError(t, rippledata.SetSigners(expectedTx, trustSetOperationValidSigners...))
				xrplRPCClientMock.EXPECT().Submit(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, tx rippledata.Transaction) (xrpl.SubmitResult, error) {
					_, expectedTxRaw, err := rippledata.Raw(expectedTx)
					require.NoError(t, err)
					_, txRaw, err := rippledata.Raw(tx)
					require.NoError(t, err)
					require.Equal(t, expectedTxRaw, txRaw)
					return xrpl.SubmitResult{}, nil
				})

				return xrplRPCClientMock
			},
		},
		{
			name: "register_signature_for_coreum_to_XRPL_XRPL_originated_token_transfer_payment_tx",
			contractClientBuilder: func(ctrl *gomock.Controller) processes.ContractClient {
				contractClientMock := NewMockContractClient(ctrl)
				contractClientMock.EXPECT().GetPendingOperations(gomock.Any()).Return([]coreum.Operation{coreumToXRPLXRPLOriginatedTokenTransferOperation}, nil)
				contractClientMock.EXPECT().GetContractConfig(gomock.Any()).Return(coreum.ContractConfig{
					Relayers: contractRelayers,
				}, nil)
				contractClientMock.EXPECT().SaveSignature(gomock.Any(), contractRelayers[0].CoreumAddress, coreumToXRPLXRPLOriginatedTokenTransferOperation.TicketSequence, coreumToXRPLXRPLOriginatedTokenTransferOperationValidSigners[0].Signer.TxnSignature.String())
				return contractClientMock
			},
			xrplRPCClientBuilder: func(ctrl *gomock.Controller) processes.XRPLRPCClient {
				xrplRPCClientMock := NewMockXRPLRPCClient(ctrl)
				xrplRPCClientMock.EXPECT().AccountInfo(gomock.Any(), bridgeXRPLAddress).Return(bridgeXRPLSignerAccountWithSigners, nil)
				return xrplRPCClientMock
			},
			xrplTxSignerBuilder: func(ctrl *gomock.Controller) processes.XRPLTxSigner {
				xrplTxSignerMock := NewMockXRPLTxSigner(ctrl)
				tx, err := processes.BuildCoreumToXRPLXRPLOriginatedTokenTransferPaymentTxForMultiSigning(bridgeXRPLAddress, coreumToXRPLXRPLOriginatedTokenTransferOperation)
				require.NoError(t, err)
				xrplTxSignerMock.EXPECT().MultiSign(tx, xrplTxSignerKeyName).Return(coreumToXRPLXRPLOriginatedTokenTransferOperationValidSigners[0], nil)

				return xrplTxSignerMock
			},
		},
		{
			name: "submit_coreum_to_XRPL_XRPL_originated_token_transfer_payment_tx_with_filtered_signature",
			contractClientBuilder: func(ctrl *gomock.Controller) processes.ContractClient {
				contractClientMock := NewMockContractClient(ctrl)
				contractClientMock.EXPECT().GetPendingOperations(gomock.Any()).Return([]coreum.Operation{coreumToXRPLXRPLOriginatedTokenTransferOperationWithSignatures}, nil)
				contractClientMock.EXPECT().GetContractConfig(gomock.Any()).Return(coreum.ContractConfig{
					Relayers: contractRelayers,
				}, nil)
				return contractClientMock
			},
			xrplRPCClientBuilder: func(ctrl *gomock.Controller) processes.XRPLRPCClient {
				xrplRPCClientMock := NewMockXRPLRPCClient(ctrl)
				xrplRPCClientMock.EXPECT().AccountInfo(gomock.Any(), bridgeXRPLAddress).Return(bridgeXRPLSignerAccountWithSigners, nil)
				expectedTx, err := processes.BuildCoreumToXRPLXRPLOriginatedTokenTransferPaymentTxForMultiSigning(bridgeXRPLAddress, coreumToXRPLXRPLOriginatedTokenTransferOperationWithSignatures)
				require.NoError(t, err)
				require.NoError(t, rippledata.SetSigners(expectedTx, coreumToXRPLXRPLOriginatedTokenTransferOperationValidSigners...))
				xrplRPCClientMock.EXPECT().Submit(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, tx rippledata.Transaction) (xrpl.SubmitResult, error) {
					_, expectedTxRaw, err := rippledata.Raw(expectedTx)
					require.NoError(t, err)
					_, txRaw, err := rippledata.Raw(tx)
					require.NoError(t, err)
					require.Equal(t, expectedTxRaw, txRaw)
					return xrpl.SubmitResult{}, nil
				})

				return xrplRPCClientMock
			},
		},
		{
			name: "register_signature_for_coreum_to_XRPL_Coreum_originated_token_transfer_payment_tx",
			contractClientBuilder: func(ctrl *gomock.Controller) processes.ContractClient {
				contractClientMock := NewMockContractClient(ctrl)
				contractClientMock.EXPECT().GetPendingOperations(gomock.Any()).Return([]coreum.Operation{coreumToXRPLCoreumOriginatedTokenTransferOperation}, nil)
				contractClientMock.EXPECT().GetContractConfig(gomock.Any()).Return(coreum.ContractConfig{
					Relayers: contractRelayers,
				}, nil)
				contractClientMock.EXPECT().GetCoreumTokenByXRPLCurrency(gomock.Any(), coreumToXRPLCoreumOriginatedTokenTransferOperation.OperationType.CoreumToXRPLTransfer.Currency).
					Return(coreum.CoreumToken{
						Decimals: coreumOriginatedTokenDecimals,
					}, nil)
				contractClientMock.EXPECT().SaveSignature(gomock.Any(), contractRelayers[0].CoreumAddress, coreumToXRPLCoreumOriginatedTokenTransferOperation.TicketSequence, coreumToXRPLCoreumOriginatedTokenTransferOperationValidSigners[0].Signer.TxnSignature.String())
				return contractClientMock
			},
			xrplRPCClientBuilder: func(ctrl *gomock.Controller) processes.XRPLRPCClient {
				xrplRPCClientMock := NewMockXRPLRPCClient(ctrl)
				xrplRPCClientMock.EXPECT().AccountInfo(gomock.Any(), bridgeXRPLAddress).Return(bridgeXRPLSignerAccountWithSigners, nil)
				return xrplRPCClientMock
			},
			xrplTxSignerBuilder: func(ctrl *gomock.Controller) processes.XRPLTxSigner {
				xrplTxSignerMock := NewMockXRPLTxSigner(ctrl)
				tx, err := processes.BuildCoreumToXRPLCoreumOriginatedTokenTransferPaymentTxForMultiSigning(bridgeXRPLAddress, coreumToXRPLCoreumOriginatedTokenTransferOperation, coreumOriginatedTokenDecimals)
				require.NoError(t, err)
				xrplTxSignerMock.EXPECT().MultiSign(tx, xrplTxSignerKeyName).Return(coreumToXRPLCoreumOriginatedTokenTransferOperationValidSigners[0], nil)

				return xrplTxSignerMock
			},
		},
		{
			name: "submit_coreum_to_XRPL_Coreum_originated_token_transfer_payment_tx_with_filtered_signature",
			contractClientBuilder: func(ctrl *gomock.Controller) processes.ContractClient {
				contractClientMock := NewMockContractClient(ctrl)
				contractClientMock.EXPECT().GetPendingOperations(gomock.Any()).Return([]coreum.Operation{coreumToXRPLCoreumOriginatedTokenTransferOperationWithSignatures}, nil)
				contractClientMock.EXPECT().GetContractConfig(gomock.Any()).Return(coreum.ContractConfig{
					Relayers: contractRelayers,
				}, nil)
				contractClientMock.EXPECT().GetCoreumTokenByXRPLCurrency(gomock.Any(), coreumToXRPLCoreumOriginatedTokenTransferOperationWithSignatures.OperationType.CoreumToXRPLTransfer.Currency).
					Return(coreum.CoreumToken{
						Decimals: coreumOriginatedTokenDecimals,
					}, nil).Times(3)
				return contractClientMock
			},
			xrplRPCClientBuilder: func(ctrl *gomock.Controller) processes.XRPLRPCClient {
				xrplRPCClientMock := NewMockXRPLRPCClient(ctrl)
				xrplRPCClientMock.EXPECT().AccountInfo(gomock.Any(), bridgeXRPLAddress).Return(bridgeXRPLSignerAccountWithSigners, nil)
				expectedTx, err := processes.BuildCoreumToXRPLCoreumOriginatedTokenTransferPaymentTxForMultiSigning(bridgeXRPLAddress, coreumToXRPLCoreumOriginatedTokenTransferOperationWithSignatures, coreumOriginatedTokenDecimals)
				require.NoError(t, err)
				require.NoError(t, rippledata.SetSigners(expectedTx, coreumToXRPLCoreumOriginatedTokenTransferOperationValidSigners...))
				xrplRPCClientMock.EXPECT().Submit(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, tx rippledata.Transaction) (xrpl.SubmitResult, error) {
					_, expectedTxRaw, err := rippledata.Raw(expectedTx)
					require.NoError(t, err)
					_, txRaw, err := rippledata.Raw(tx)
					require.NoError(t, err)
					require.Equal(t, expectedTxRaw, txRaw)
					return xrpl.SubmitResult{}, nil
				})

				return xrplRPCClientMock
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)

			ctrl := gomock.NewController(t)
			logMock := logger.NewAnyLogMock(ctrl)
			var contractClient processes.ContractClient
			if tt.contractClientBuilder != nil {
				contractClient = tt.contractClientBuilder(ctrl)
			}

			var xrplRPCClient processes.XRPLRPCClient
			if tt.xrplRPCClientBuilder != nil {
				xrplRPCClient = tt.xrplRPCClientBuilder(ctrl)
			}

			var xrplTxSigner processes.XRPLTxSigner
			if tt.xrplTxSignerBuilder != nil {
				xrplTxSigner = tt.xrplTxSignerBuilder(ctrl)
			}

			o := processes.NewXRPLTxSubmitter(
				processes.XRPLTxSubmitterConfig{
					BridgeXRPLAddress:    bridgeXRPLAddress,
					RelayerCoreumAddress: contractRelayers[0].CoreumAddress,
					XRPLTxSignerKeyName:  xrplTxSignerKeyName,
				},
				logMock,
				contractClient,
				xrplRPCClient,
				xrplTxSigner,
			)
			require.NoError(t, o.Start(ctx))
		})
	}
}

func genContractRelayers(relayersCount int) ([]coreum.Relayer, []*xrpl.PrivKeyTxSigner, xrpl.AccountInfoResult) {
	contractRelayers := make([]coreum.Relayer, 0)
	xrplTxSigners := make([]*xrpl.PrivKeyTxSigner, 0)
	for i := 0; i < relayersCount; i++ {
		xrplRelayerSigner := xrpl.GenPrivKeyTxSigner()
		xrplTxSigners = append(xrplTxSigners, xrplRelayerSigner)

		coreumRelayerAddress := coreum.GenAccount()
		contractRelayers = append(contractRelayers, coreum.Relayer{
			CoreumAddress: coreumRelayerAddress,
			XRPLAddress:   xrplRelayerSigner.Account().String(),
			XRPLPubKey:    xrplRelayerSigner.PubKey().String(),
		})
	}

	signerQuorum := uint32(2)
	signerEntries := make([]rippledata.SignerEntry, 0)
	for _, xrplTxSigner := range xrplTxSigners {
		signerAcc := xrplTxSigner.Account()
		signerEntries = append(signerEntries, rippledata.SignerEntry{
			SignerEntry: rippledata.SignerEntryItem{
				Account:      &signerAcc,
				SignerWeight: lo.ToPtr(uint16(1)),
			},
		})
	}

	bridgeXRPLSignerAccountWithSigners := xrpl.AccountInfoResult{
		AccountData: xrpl.AccountDataWithSigners{
			AccountRoot: rippledata.AccountRoot{
				Sequence: lo.ToPtr(uint32(1)),
			},
			SignerList: []rippledata.SignerList{{
				SignerEntries: signerEntries,
				SignerQuorum:  &signerQuorum,
			}},
		},
	}

	return contractRelayers, xrplTxSigners, bridgeXRPLSignerAccountWithSigners
}

func buildAllocateTicketsTestData(
	t *testing.T,
	xrplTxSigners []*xrpl.PrivKeyTxSigner,
	bridgeXRPLAddress rippledata.Account,
	contractRelayers []coreum.Relayer,
) (
	coreum.Operation, coreum.Operation, coreum.Operation, []rippledata.Signer,
) {
	operation := coreum.Operation{
		AccountSequence: 1,
		Signatures:      nil,
		OperationType: coreum.OperationType{
			AllocateTickets: &coreum.OperationTypeAllocateTickets{
				Number: 3,
			},
		},
	}

	operationUnexpectedSeqNumber := coreum.Operation{
		AccountSequence: 999,
		Signatures:      nil,
		OperationType: coreum.OperationType{
			AllocateTickets: &coreum.OperationTypeAllocateTickets{
				Number: 3,
			},
		},
	}

	operationWithSignatures, validSigners := multiSignOperationFromMultipleSignersWithLastInvalidSignature(
		t,
		operation,
		xrplTxSigners,
		contractRelayers,
		bridgeXRPLAddress,
		multiSignAllocateTicketsOperation,
	)

	return operation, operationUnexpectedSeqNumber, operationWithSignatures, validSigners
}

func buildTrustSetTestData(
	t *testing.T,
	xrplTxSigners []*xrpl.PrivKeyTxSigner,
	bridgeXRPLAddress rippledata.Account,
	contractRelayers []coreum.Relayer,
) (
	coreum.Operation, coreum.Operation, []rippledata.Signer,
) {
	operation := coreum.Operation{
		TicketSequence: 1,
		Signatures:     nil,
		OperationType: coreum.OperationType{
			TrustSet: &coreum.OperationTypeTrustSet{
				Issuer:              xrpl.GenPrivKeyTxSigner().Account().String(),
				Currency:            "TRC",
				TrustSetLimitAmount: sdkmath.NewInt(1000000000000),
			},
		},
	}

	operationWithSignatures, validSigners := multiSignOperationFromMultipleSignersWithLastInvalidSignature(
		t,
		operation,
		xrplTxSigners,
		contractRelayers,
		bridgeXRPLAddress,
		multiSignTrustSetOperation,
	)

	return operation, operationWithSignatures, validSigners
}

func buildCoreumToXRPLXRPLOriginatedTokenTransferTestData(
	t *testing.T,
	xrplTxSigners []*xrpl.PrivKeyTxSigner,
	bridgeXRPLAddress rippledata.Account,
	contractRelayers []coreum.Relayer,
) (
	coreum.Operation, coreum.Operation, []rippledata.Signer,
) {
	operation := coreum.Operation{
		TicketSequence: 1,
		Signatures:     nil,
		OperationType: coreum.OperationType{
			CoreumToXRPLTransfer: &coreum.OperationTypeCoreumToXRPLTransfer{
				Issuer:    xrpl.GenPrivKeyTxSigner().Account().String(),
				Currency:  "TRC",
				Amount:    sdkmath.NewIntFromUint64(123),
				Recipient: xrpl.GenPrivKeyTxSigner().Account().String(),
			},
		},
	}

	operationWithSignatures, validSigners := multiSignOperationFromMultipleSignersWithLastInvalidSignature(
		t,
		operation,
		xrplTxSigners,
		contractRelayers,
		bridgeXRPLAddress,
		multiSignCoreumToXRPLXRPLOriginatedTokeTransferOperation,
	)

	return operation, operationWithSignatures, validSigners
}

func buildCoreumToXRPLCoreumOriginatedTokenTransferTestData(
	t *testing.T,
	xrplTxSigners []*xrpl.PrivKeyTxSigner,
	bridgeXRPLAddress rippledata.Account,
	contractRelayers []coreum.Relayer,
	decimals uint32,
) (
	coreum.Operation, coreum.Operation, []rippledata.Signer,
) {
	operation := coreum.Operation{
		TicketSequence: 1,
		Signatures:     nil,
		OperationType: coreum.OperationType{
			CoreumToXRPLTransfer: &coreum.OperationTypeCoreumToXRPLTransfer{
				// the bridge is issuer of the token
				Issuer:    bridgeXRPLAddress.String(),
				Currency:  "TRC",
				Amount:    sdkmath.NewIntFromUint64(123),
				Recipient: xrpl.GenPrivKeyTxSigner().Account().String(),
			},
		},
	}

	operationWithSignatures, validSigners := multiSignOperationFromMultipleSignersWithLastInvalidSignature(
		t,
		operation,
		xrplTxSigners,
		contractRelayers,
		bridgeXRPLAddress,
		func(t *testing.T, signer *xrpl.PrivKeyTxSigner, account rippledata.Account, operation coreum.Operation) rippledata.Signer {
			return multiSignCoreumToXRPLCoreumOriginatedTokeTransferOperation(t, signer, account, operation, decimals)
		},
	)

	return operation, operationWithSignatures, validSigners
}

func multiSignOperationFromMultipleSignersWithLastInvalidSignature(
	t *testing.T,
	operation coreum.Operation,
	xrplTxSigners []*xrpl.PrivKeyTxSigner,
	contractRelayers []coreum.Relayer,
	bridgeXRPLAddress rippledata.Account,
	signingFunc func(*testing.T, *xrpl.PrivKeyTxSigner, rippledata.Account, coreum.Operation) rippledata.Signer,
) (coreum.Operation, []rippledata.Signer) {
	require.Equal(t, len(xrplTxSigners), len(contractRelayers))
	require.Greater(t, len(xrplTxSigners), 2)
	operationWithSignatures := operation

	validSigners := make([]rippledata.Signer, 0)
	for i := 0; i < len(xrplTxSigners); i++ {
		signer := signingFunc(
			t,
			xrplTxSigners[i],
			bridgeXRPLAddress,
			operationWithSignatures,
		)
		operationWithSignatures.Signatures = append(operationWithSignatures.Signatures, coreum.Signature{
			RelayerCoreumAddress: contractRelayers[i].CoreumAddress,
			Signature:            signer.Signer.TxnSignature.String(),
		})
		// the set last signature equal first to make it invalid
		if len(xrplTxSigners)-1 == i {
			operationWithSignatures.Signatures[len(xrplTxSigners)-1] = operationWithSignatures.Signatures[0]
			break
		}
		validSigners = append(validSigners, signer)
	}

	return operationWithSignatures, validSigners
}

func multiSignAllocateTicketsOperation(
	t *testing.T,
	relayerXRPLSigner *xrpl.PrivKeyTxSigner,
	bridgeXRPLAddress rippledata.Account,
	operation coreum.Operation,
) rippledata.Signer {
	tx, err := processes.BuildTicketCreateTxForMultiSigning(bridgeXRPLAddress, operation)
	require.NoError(t, err)
	signer, err := relayerXRPLSigner.MultiSign(tx)
	require.NoError(t, err)

	return signer
}

func multiSignTrustSetOperation(
	t *testing.T,
	relayerXRPLSigner *xrpl.PrivKeyTxSigner,
	bridgeXRPLAcc rippledata.Account,
	operation coreum.Operation,
) rippledata.Signer {
	tx, err := processes.BuildTrustSetTxForMultiSigning(bridgeXRPLAcc, operation)
	require.NoError(t, err)
	signer, err := relayerXRPLSigner.MultiSign(tx)
	require.NoError(t, err)

	return signer
}

func multiSignCoreumToXRPLXRPLOriginatedTokeTransferOperation(
	t *testing.T,
	relayerXRPLSigner *xrpl.PrivKeyTxSigner,
	bridgeXRPLAcc rippledata.Account,
	operation coreum.Operation,
) rippledata.Signer {
	tx, err := processes.BuildCoreumToXRPLXRPLOriginatedTokenTransferPaymentTxForMultiSigning(bridgeXRPLAcc, operation)
	require.NoError(t, err)
	signer, err := relayerXRPLSigner.MultiSign(tx)
	require.NoError(t, err)

	return signer
}

func multiSignCoreumToXRPLCoreumOriginatedTokeTransferOperation(
	t *testing.T,
	relayerXRPLSigner *xrpl.PrivKeyTxSigner,
	bridgeXRPLAcc rippledata.Account,
	operation coreum.Operation,
	decimals uint32,
) rippledata.Signer {
	tx, err := processes.BuildCoreumToXRPLCoreumOriginatedTokenTransferPaymentTxForMultiSigning(bridgeXRPLAcc, operation, decimals)
	require.NoError(t, err)
	signer, err := relayerXRPLSigner.MultiSign(tx)
	require.NoError(t, err)

	return signer
}
