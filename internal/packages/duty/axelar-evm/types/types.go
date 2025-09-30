package types

import (
	"encoding/json"
	"time"
)

var (
	SupportedChains = []string{"axelar"}
)

const (
	// common
	CommonValidatorQueryPath = "/cosmos/staking/v1beta1/validators?status=BOND_STATUS_BONDED&pagination.count_total=true&pagination.limit=500"

	// axelar
	AxelarEvmChainsQueryPath        = "axelar/evm/v1beta1/chains?status=1"
	AxelarChainMaintainersQueryPath = "/axelar/nexus/v1beta1/chain_maintainers/{chain}"
	AxelarProxyResisterQueryPath    = `/abci_query?path="/custom/snapshot/proxy/{validator_operator_address}"`
)

type CommonAxelarNexus struct {
	ActiveEVMChains []string
	Validators      []ValidatorStatus
}

type CommonAxelarHeartbeats struct {
	Validators []BroadcastorStatus
}

type ValidatorStatus struct {
	Moniker                  string  `json:"moniker"`
	ValidatorOperatorAddress string  `json:"validator_operator_address"`
	Status                   float64 `json:"status"`
	EVMChainName             string  `json:"evm_chain_name"`
}

type BroadcastorStatus struct {
	Moniker                  string  `json:"moniker"`
	ValidatorOperatorAddress string  `json:"validator_operator_address"`
	BroadcastorAddress       string  `json:"broadcastor_address"`
	Status                   string  `json:"status"`
	LatestHeartBeat          float64 `json:"latest_heartbeat"`
}

type CommonValidatorsQueryResponse struct {
	Validators []struct {
		OperatorAddress string `json:"operator_address"`
		Description     struct {
			Moniker string `json:"moniker"`
		} `json:"description"`
	} `json:"validators"`
	Pagination struct {
		NextKey interface{} `json:"-"`
		Total   string      `json:"-"`
	} `json:"-"`
}

type AxelarEvmChainsResponse struct {
	Chains []string `json:"chains"`
}

type AxelarChainMaintainersResponse struct {
	Maintainers []string `json:"maintainers"`
}

type AxelarProxyResisterResponse struct {
	Result struct {
		Response struct {
			Code      int    `json:"code"`
			Log       string `json:"log"`
			Info      string `json:"info"`
			Index     string `json:"index"`
			Key       any    `json:"key"`
			Value     string `json:"value"`
			ProofOps  any    `json:"proofOps"`
			Height    string `json:"height"`
			Codespace string `json:"codespace"`
		} `json:"response"`
	} `json:"result"`
}

type AxelarProxyResisterStatus struct {
	Address string `json:"address"`
	Status  string `json:"status"`
}

// TODO: I think this types should move into common cosmos types
type CosmosTx struct {
	Body struct {
		Messages []json.RawMessage `json:"messages"`
	} `json:"body"`
	AuthInfo   interface{} `json:"-"`
	Signatures []string    `json:"-"`
}

type BlockTxsResponse struct {
	Txs   []CosmosTx `json:"txs"`
	Block struct {
		Header struct {
			ChainID         string    `json:"chain_id"`
			Height          string    `json:"height"`
			Time            time.Time `json:"time"`
			ProposerAddress string    `json:"proposer_address"`
		} `json:"header"`
	} `json:"block"`
}
