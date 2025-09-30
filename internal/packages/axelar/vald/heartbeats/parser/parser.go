package parser

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	commontypes "github.com/cosmostation/cvms/internal/common/types"
	"github.com/cosmostation/cvms/internal/packages/axelar/vald/heartbeats/types"
)

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

func AxelarHeartbeatsFilterInTx(tx commontypes.CosmosTx, operatorAddr string) (bool, error) {
	for _, rawMessage := range tx.Body.Messages {
		var msg map[string]json.RawMessage
		if err := json.Unmarshal(rawMessage, &msg); err != nil {
			return false, fmt.Errorf("json unmashal error: %s", err)
		}

		if rawType, ok := msg["@type"]; ok {
			var msgType string
			if err := json.Unmarshal(rawType, &msgType); err != nil {
				return false, fmt.Errorf("json unmashal error: %s", err)
			}

			if msgType == "/axelar.reward.v1beta1.RefundMsgRequest" {
				if rawInnerMessage, ok := msg["inner_message"]; ok {
					var innerMessage map[string]interface{}
					if err := json.Unmarshal(rawInnerMessage, &innerMessage); err != nil {
						return false, fmt.Errorf("json unmashal error: %s", err)
					}

					if innerMessage["@type"] == "/axelar.tss.v1beta1.HeartBeatRequest" {
						if innerMessage["sender"] == operatorAddr {
							return true, nil
						}
					}
				}
			}
		}
	}
	return false, nil
}
