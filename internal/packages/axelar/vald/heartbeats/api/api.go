package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"

	"github.com/cosmostation/cvms/internal/common"
	"github.com/cosmostation/cvms/internal/common/api"
	commontypes "github.com/cosmostation/cvms/internal/common/types"
	"github.com/cosmostation/cvms/internal/helper"
	"github.com/cosmostation/cvms/internal/packages/axelar/vald/heartbeats/parser"
	"github.com/cosmostation/cvms/internal/packages/axelar/vald/heartbeats/types"
)

const pollingSize int = 5

// Always check heartbeat tx at 51, 101, 151, ... but only once at the 55th, 105th, 155th, ... block
// This avoids repeated checks and batches all post-event tx (51~55) at once
func GetAxelarHeartbeatsStatus(
	exporter *common.Exporter,
	CommonProxyResisterQueryPath string,
	CommonProxyResisterParser func([]byte) (types.AxelarProxyResisterStatus, error),
	latestHeartbeatsHeight int64,
) (types.CommonAxelarHeartbeats, error) {
	currentBlockHeight, _, err := api.GetStatus(exporter.CommonClient)
	if err != nil {
		exporter.Errorf("api error: %s", err)
		return types.CommonAxelarHeartbeats{}, common.ErrFailedHttpRequest
	}

	//Always 51,101,151,201... submit heartbeat tx
	lastEventBlock := (currentBlockHeight / 50) * 50
	findHeartbeatsHeight := lastEventBlock + 1

	// Skip if heartbeatsHeight is greater than or equal to
	if currentBlockHeight%50 < int64(pollingSize) || latestHeartbeatsHeight >= findHeartbeatsHeight {
		return types.CommonAxelarHeartbeats{}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), common.Timeout)
	defer cancel()

	// create requester
	requester := exporter.APIClient.R().SetContext(ctx)

	// get on-chain validators
	resp, err := requester.Get(types.CommonValidatorQueryPath)
	if err != nil {
		exporter.Errorf("api error: %s", err)
		return types.CommonAxelarHeartbeats{}, common.ErrFailedHttpRequest
	}
	if resp.StatusCode() != http.StatusOK {
		exporter.Errorf("api error: got %d code from %s", resp.StatusCode(), resp.Request.URL)
		return types.CommonAxelarHeartbeats{}, common.ErrGotStrangeStatusCode
	}

	// json unmarsharling received validators data
	var validators types.CommonValidatorsQueryResponse
	if err := json.Unmarshal(resp.Body(), &validators); err != nil {
		exporter.Errorf("api error: %s", err)
		return types.CommonAxelarHeartbeats{}, common.ErrFailedJsonUnmarshal
	}

	// // init channel and waitgroup for go-routine
	ch := make(chan helper.Result)
	var wg sync.WaitGroup
	totalBroadcastorStatus := make([]types.BroadcastorStatus, 0)

	updateHeight, err := findHeartbeats(ctx, &ch, &wg, exporter, validators, findHeartbeatsHeight)

	// close channel
	errorCounter := 0
	go func() {
		wg.Wait()
		close(ch)
	}()

	for r := range ch {
		if r.Success {
			if item, ok := r.Item.(types.BroadcastorStatus); ok {
				totalBroadcastorStatus = append(totalBroadcastorStatus, item)
				continue
			}
		}
		errorCounter++
	}

	if errorCounter > 0 {
		return types.CommonAxelarHeartbeats{}, errors.New("failed to get heartbeats from go-routine")
	}

	latestHeartbeatsHeight = updateHeight
	return types.CommonAxelarHeartbeats{Validators: totalBroadcastorStatus, LatestHeartBeatsHeight: updateHeight}, nil
}

func findHeartbeats(
	ctx context.Context,
	ch *chan helper.Result,
	wg *sync.WaitGroup,
	exporter *common.Exporter,
	validators types.CommonValidatorsQueryResponse,
	heartbeatsHeight int64,
) (int64, error) {
	blockTxsCache := make([]commontypes.CosmosTx, 0)

	// Sometimes, due to block size or timing,
	// the heartbeats tx gets put into the next block,
	// so we check +polling size block more to find it.
	for i := 0; i < pollingSize; i++ {
		blockHeight := heartbeatsHeight + int64(i)
		_, _, blockTxs, err := api.GetBlockAndTxs(exporter.CommonClient, blockHeight)
		if err != nil {
			return heartbeatsHeight, common.ErrFailedHttpRequest
		}

		blockTxsCache = append(blockTxsCache, blockTxs...)
	}

	for _, validator := range validators.Validators {
		wg.Add(1)
		operatorAddr := validator.OperatorAddress
		moniker := validator.Description.Moniker

		go func(operatorAddr, moniker string) {
			defer wg.Done()

			// Find Broadcastor address
			rpcRequester := exporter.RPCClient.R().SetContext(ctx)
			abciQueryPath := strings.Replace(types.AxelarProxyResisterQueryPath, "{validator_operator_address}", operatorAddr, -1)
			resp, err := rpcRequester.Get(abciQueryPath)
			if err != nil {
				exporter.Errorf("API error: %s", err)
				*ch <- helper.Result{Item: nil, Success: false}
				return
			}

			var AxelarProxyResisterResponse types.AxelarProxyResisterResponse
			if err := json.Unmarshal(resp.Body(), &AxelarProxyResisterResponse); err != nil {
				exporter.Errorf("JSON unmarshal error: %s", err)
				*ch <- helper.Result{Item: nil, Success: false}
				return
			}

			// The ABCI query returns an encoded value, so try to decode it to base64.
			decodedBytes, err := base64.StdEncoding.DecodeString(AxelarProxyResisterResponse.Result.Response.Value)
			if err != nil {
				exporter.Errorf("Failed to decode Base64: %s", err)
				*ch <- helper.Result{Item: nil, Success: false}
				return
			}

			var AxelarProxyResisterStatus types.AxelarProxyResisterStatus
			if err := json.Unmarshal(decodedBytes, &AxelarProxyResisterStatus); err != nil {
				exporter.Errorf("JSON unmarshal error: %s", err)
				*ch <- helper.Result{Item: nil, Success: false}
				return
			}

			newBroadcastorStatus := types.BroadcastorStatus{
				Moniker:                  moniker,
				ValidatorOperatorAddress: operatorAddr,
				BroadcastorAddress:       AxelarProxyResisterStatus.Address,
				Status:                   "",
			}

			foundHealthyTx := false

			for i := 0; i < pollingSize; i++ {
				for _, tx := range blockTxsCache {
					isHealthy, err := parser.AxelarHeartbeatsFilterInTx(tx, AxelarProxyResisterStatus.Address)
					if err != nil {
						exporter.Error(err)
						*ch <- helper.Result{Item: nil, Success: false}
						return
					}

					if isHealthy {
						newBroadcastorStatus.Status = "success"
						foundHealthyTx = true
						break
					}
					if foundHealthyTx {
						break
					}
				}
			}

			// If heartbeats tx is not found after scanning all blocks, handle missed
			if !foundHealthyTx {
				newBroadcastorStatus.Status = "missed"
			}

			*ch <- helper.Result{Item: newBroadcastorStatus, Success: true}
		}(operatorAddr, moniker)
	}

	return heartbeatsHeight, nil
}
