package parser

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/cosmostation/cvms/internal/common"
	"github.com/cosmostation/cvms/internal/common/api"
	"github.com/cosmostation/cvms/internal/helper"
	"github.com/cosmostation/cvms/internal/helper/logger"
	"github.com/cosmostation/cvms/internal/packages/axelar/vald/heartbeats/types"
	"github.com/stretchr/testify/assert"
)

var (
	testHeight int64 = 17928651

	p = common.Packager{
		ChainName:    "axelar",
		ChainID:      "axelar-testnet-lisbon-3",
		ProtocolType: "cosmos",
		Endpoints: common.Endpoints{
			RPCs: []string{""},
			APIs: []string{""},
		},
		Logger: logger.GetTestLogger(),
	}
)

func TestAxelarHeartbeatsFilterInTx(t *testing.T) {
	app := common.NewCommonApp(p)
	app.SetAPIEndPoint(p.Endpoints.APIs[0])
	app.SetRPCEndPoint(p.Endpoints.RPCs[0])

	currentBlockHeight, _, err := api.GetStatus(app.CommonClient)
	assert.NoError(t, err)

	_ = currentBlockHeight

	ctx, cancel := context.WithTimeout(context.Background(), common.Timeout)
	defer cancel()
	// create requester
	requester := app.APIClient.R().SetContext(ctx)

	_, _, blockTxs, err := api.GetBlockAndTxs(app.CommonClient, testHeight)
	assert.NoError(t, err)

	// get on-chain validators
	resp, err := requester.Get(types.CommonValidatorQueryPath)
	assert.NoError(t, err)

	// json unmarsharling received validators data
	var validators types.CommonValidatorsQueryResponse
	err = json.Unmarshal(resp.Body(), &validators)
	assert.NoError(t, err)

	// init channel and waitgroup for go-routine
	ch := make(chan helper.Result)
	var wg sync.WaitGroup

	for _, validator := range validators.Validators {
		wg.Add(1)
		operatorAddr := validator.OperatorAddress
		moniker := validator.Description.Moniker

		go func(operatorAddr, moniker string) {
			defer wg.Done()
			rpcRequester := app.RPCClient.R().SetContext(ctx)
			abciQueryPath := strings.Replace(types.AxelarProxyResisterQueryPath, "{validator_operator_address}", operatorAddr, -1)
			resp, err := rpcRequester.Get(abciQueryPath)
			if err != nil {
				fmt.Errorf("api error: %s", err)
				ch <- helper.Result{Item: nil, Success: false}
				return
			}

			var AxelarProxyResisterResponse types.AxelarProxyResisterResponse
			if err := json.Unmarshal(resp.Body(), &AxelarProxyResisterResponse); err != nil {
				fmt.Errorf("json unmashal error: %s", err)
				ch <- helper.Result{Item: nil, Success: false}
				return
			}

			// Base64 decode
			decodedBytes, err := base64.StdEncoding.DecodeString(AxelarProxyResisterResponse.Result.Response.Value)
			if err != nil {
				fmt.Printf("failed to decode Base64: %s", err)
				return
			}

			var AxelarProxyResisterStatus types.AxelarProxyResisterStatus
			err = json.Unmarshal(decodedBytes, &AxelarProxyResisterStatus)
			if err != nil {
				fmt.Printf("json unmashal error: %s", err)
				return
			}

			for _, tx := range blockTxs {
				isHealthy, err := AxelarHeartbeatsFilterInTx(tx, AxelarProxyResisterStatus.Address)
				if err != nil {
					fmt.Printf("failed filtering tx: %s", err)
					ch <- helper.Result{
						Item:    nil,
						Success: false,
					}
				}

				if isHealthy {
					var newBroadcastorStatus = types.BroadcastorStatus{
						Moniker:                  moniker,
						ValidatorOperatorAddress: operatorAddr,
						BroadcastorAddress:       AxelarProxyResisterStatus.Address,
						Status:                   AxelarProxyResisterStatus.Status,
					}

					ch <- helper.Result{Item: newBroadcastorStatus, Success: true}
				}
			}
		}(operatorAddr, moniker)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for r := range ch {
		if r.Success {
			if item, ok := r.Item.(types.BroadcastorStatus); ok {
				fmt.Println("=============================================")
				fmt.Println("Moniker:", item.Moniker)
				fmt.Println("ValidatorOperatorAddress:", item.ValidatorOperatorAddress)
				fmt.Println("BroadcastorAddress:", item.BroadcastorAddress)
				fmt.Println("Status:", item.Status)
			}
		} else {
			fmt.Println("WTF")
		}
	}
}
