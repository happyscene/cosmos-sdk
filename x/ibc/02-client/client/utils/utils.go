package utils

import (
	"context"

	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	ibcclient "github.com/cosmos/cosmos-sdk/x/ibc/client"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
)

// QueryClientState returns a client state. If prove is true, it performs an ABCI store query
// in order to retrieve the merkle proof. Otherwise, it uses the gRPC query client.
func QueryClientState(
	clientCtx client.Context, clientID string, prove bool,
) (*types.QueryClientStateResponse, error) {
	if prove {
		return QueryClientStateABCI(clientCtx, clientID)
	}

	queryClient := types.NewQueryClient(clientCtx)
	req := &types.QueryClientStateRequest{
		ClientId: clientID,
	}

	return queryClient.ClientState(context.Background(), req)
}

// QueryClientStateABCI queries the store to get the light client state and a merkle proof.
func QueryClientStateABCI(
	clientCtx client.Context, clientID string,
) (*types.QueryClientStateResponse, error) {
	key := host.FullKeyClientPath(clientID, host.KeyClientState())

	value, proofBz, proofHeight, err := ibcclient.QueryTendermintProof(clientCtx, key)
	if err != nil {
		return nil, err
	}

	cdc := codec.NewProtoCodec(clientCtx.InterfaceRegistry)

	clientState, err := types.UnmarshalClientState(cdc, value)
	if err != nil {
		return nil, err
	}

	anyClientState, err := types.PackClientState(clientState)
	if err != nil {
		return nil, err
	}

	clientStateRes := types.NewQueryClientStateResponse(clientID, anyClientState, proofBz, proofHeight)
	return clientStateRes, nil
}

// QueryConsensusState returns a consensus state. If prove is true, it performs an ABCI store
// query in order to retrieve the merkle proof. Otherwise, it uses the gRPC query client.
func QueryConsensusState(
	clientCtx client.Context, clientID string, height exported.Height, prove, latestHeight bool,
) (*types.QueryConsensusStateResponse, error) {
	if prove {
		return QueryConsensusStateABCI(clientCtx, clientID, height)
	}

	queryClient := types.NewQueryClient(clientCtx)
	req := &types.QueryConsensusStateRequest{
		ClientId:     clientID,
		EpochNumber:  height.GetEpochNumber(),
		EpochHeight:  height.GetEpochHeight(),
		LatestHeight: latestHeight,
	}

	return queryClient.ConsensusState(context.Background(), req)
}

// QueryConsensusStateABCI queries the store to get the consensus state of a light client and a
// merkle proof of its existence or non-existence.
func QueryConsensusStateABCI(
	clientCtx client.Context, clientID string, height exported.Height,
) (*types.QueryConsensusStateResponse, error) {
	key := host.FullKeyClientPath(clientID, host.KeyConsensusState(height))

	value, proofBz, proofHeight, err := ibcclient.QueryTendermintProof(clientCtx, key)
	if err != nil {
		return nil, err
	}

	cdc := codec.NewProtoCodec(clientCtx.InterfaceRegistry)

	cs, err := types.UnmarshalConsensusState(cdc, value)
	if err != nil {
		return nil, err
	}

	anyConsensusState, err := types.PackConsensusState(cs)
	if err != nil {
		return nil, err
	}

	// TODO: retrieve epoch-number from chain-id
	return types.NewQueryConsensusStateResponse(clientID, anyConsensusState, proofBz, proofHeight), nil
}

// QueryTendermintHeader takes a client context and returns the appropriate
// tendermint header
func QueryTendermintHeader(clientCtx client.Context) (ibctmtypes.Header, int64, error) {
	node, err := clientCtx.GetNode()
	if err != nil {
		return ibctmtypes.Header{}, 0, err
	}

	info, err := node.ABCIInfo()
	if err != nil {
		return ibctmtypes.Header{}, 0, err
	}

	height := info.Response.LastBlockHeight

	commit, err := node.Commit(&height)
	if err != nil {
		return ibctmtypes.Header{}, 0, err
	}

	page := 0
	count := 10_000

	validators, err := node.Validators(&height, &page, &count)
	if err != nil {
		return ibctmtypes.Header{}, 0, err
	}

	protoCommit := commit.SignedHeader.ToProto()
	protoValset, err := tmtypes.NewValidatorSet(validators.Validators).ToProto()
	if err != nil {
		return ibctmtypes.Header{}, 0, err
	}

	header := ibctmtypes.Header{
		SignedHeader: protoCommit,
		ValidatorSet: protoValset,
	}

	return header, height, nil
}

// QueryNodeConsensusState takes a client context and returns the appropriate
// tendermint consensus state
func QueryNodeConsensusState(clientCtx client.Context) (*ibctmtypes.ConsensusState, int64, error) {
	node, err := clientCtx.GetNode()
	if err != nil {
		return &ibctmtypes.ConsensusState{}, 0, err
	}

	info, err := node.ABCIInfo()
	if err != nil {
		return &ibctmtypes.ConsensusState{}, 0, err
	}

	height := info.Response.LastBlockHeight

	commit, err := node.Commit(&height)
	if err != nil {
		return &ibctmtypes.ConsensusState{}, 0, err
	}

	page := 0
	count := 10_000

	nextHeight := height + 1
	nextVals, err := node.Validators(&nextHeight, &page, &count)
	if err != nil {
		return &ibctmtypes.ConsensusState{}, 0, err
	}

	state := &ibctmtypes.ConsensusState{
		Timestamp:          commit.Time,
		Root:               commitmenttypes.NewMerkleRoot(commit.AppHash),
		NextValidatorsHash: tmtypes.NewValidatorSet(nextVals.Validators).Hash(),
	}

	return state, height, nil
}
