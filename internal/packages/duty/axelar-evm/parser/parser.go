package parser

import (
	"encoding/base64"
	"encoding/json"

	"github.com/cosmostation/cvms/internal/packages/duty/axelar-evm/types"
)

// axelar
func AxelarEvmChainsParser(resp []byte) ([]string, error) {
	var result types.AxelarEvmChainsResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return []string{}, nil
	}
	return result.Chains, nil
}

func AxelarChainMaintainersParser(resp []byte) ([]string, error) {
	var result types.AxelarChainMaintainersResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return []string{}, nil
	}
	return result.Maintainers, nil
}

func AxelarProxyResisterParser(resp []byte) (types.AxelarProxyResisterStatus, error) {
	var result types.AxelarProxyResisterResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return types.AxelarProxyResisterStatus{}, err
	}

	decoded, err := base64.StdEncoding.DecodeString(result.Result.Response.Value)
	if err != nil {
		return types.AxelarProxyResisterStatus{}, err
	}

	var proxyResister types.AxelarProxyResisterStatus
	err = json.Unmarshal(decoded, &proxyResister)
	if err != nil {
		return types.AxelarProxyResisterStatus{}, err
	}

	return proxyResister, nil
}
