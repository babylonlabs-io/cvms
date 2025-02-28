package types

import "fmt"

type Pagination struct {
	NextKey string `json:"next_key"`
	Total   string `json:"total"`
}

var BabylonFinalityParamsQueryPath string = "/babylon/finality/v1/params"

type FinalityParams struct {
	Params struct {
		MaxActiveFinalityProviders int64  `json:"max_active_finality_providers"`
		SignedBlocksWindow         string `json:"signed_blocks_window"`
		FinalitySigTimeout         string `json:"finality_sig_timeout"`
		MinSignedPerWindow         string `json:"min_signed_per_window"`
		MinPubRand                 string `json:"min_pub_rand"`
		JailDuration               string `json:"jail_duration"`
		FinalityActivationHeight   string `json:"finality_activation_height"`
	} `json:"params"`
}

var BabylonFinalityProviderSigninInfoQueryPath = func(BTCPK string) string {
	return fmt.Sprintf("/babylon/finality/v1/signing_infos/%s", BTCPK)
}

type FinalityProviderSigningInfo struct {
	FPBTCPkHex          string `json:"fp_btc_pk_hex"`
	StartHeight         string `json:"start_height"`
	MissedBlocksCounter string `json:"missed_blocks_counter"`
	JailedUntil         string `json:"jailed_until"`
}

type FinalityProviderSigningInfoResponse struct {
	SigningInfo FinalityProviderSigningInfo `json:"signing_info"`
}

type FinalityVotesResponse struct {
	BTCPKs []string `json:"btc_pks"`
}

var BabylonFinalityProviderInfosQueryPath = func(key string) string {
	return fmt.Sprintf("/babylon/btcstaking/v1/finality_providers?pagination.key=%s", key)
}

type FinalityProviderInfosResponse struct {
	FinalityProviders []FinalityProviderInfo `json:"finality_providers"`
	Pagination        Pagination             `json:"pagination"`
}

type FinalityProviderInfo struct {
	Description struct {
		Moniker string `json:"moniker"`
	} `json:"description"`
	Address string `json:"addr"`
	BTCPK   string `json:"btc_pk"`
	Jailed  bool   `json:"jailed"`
	Active  bool
}

var BabylonFinalityProvidersQueryPath = func(height int64) string {
	return fmt.Sprintf("/babylon/finality/v1/finality_providers/%d", height)
}

type FinalityProvider struct {
	BtcPkHex             string `json:"btc_pk_hex"`
	Height               string `json:"height"`
	VotingPower          string `json:"voting_power"`
	SlashedBabylonHeight string `json:"slashed_babylon_height"`
	SlashedBtcHeight     int    `json:"slashed_btc_height"`
	Jailed               bool   `json:"jailed"`
	HighestVotedHeight   int    `json:"highest_voted_height"`
}

type FinalityProvidersResponse struct {
	FinalityProviders []FinalityProvider `json:"finality_providers"`
	Pagination        Pagination         `json:"pagination"`
}

var BabylonFinalityVotesQueryPath = func(height int64) string {
	return fmt.Sprintf("/babylon/finality/v1/votes/%d", height)
}

var BabylonBTCLightClientParamsQueryPath string = "/babylon/btclightclient/v1/params"

type BabylonBTCLightClientParams struct {
	Params struct {
		InsertHeadersAllowList []string `json:"insert_headers_allow_list"`
	} `json:"params"`
}

var BabylonCovenantCommitteeParamsQueryPath string = "/babylon/btcstaking/v1/params"

type CovenantCommitteeParams struct {
	Params struct {
		CovenantPks                  []string `json:"covenant_pks"`
		CovenantQuorum               int      `json:"covenant_quorum"`
		MinStakingValueSat           string   `json:"min_staking_value_sat"`
		MaxStakingValueSat           string   `json:"max_staking_value_sat"`
		MinStakingTimeBlocks         int      `json:"min_staking_time_blocks"`
		MaxStakingTimeBlocks         int      `json:"max_staking_time_blocks"`
		SlashingPkScript             string   `json:"slashing_pk_script"`
		MinSlashingTxFeeSat          string   `json:"min_slashing_tx_fee_sat"`
		SlashingRate                 string   `json:"slashing_rate"`
		UnbondingTimeBlocks          int      `json:"unbonding_time_blocks"`
		UnbondingFeeSat              string   `json:"unbonding_fee_sat"`
		MinCommissionRate            string   `json:"min_commission_rate"`
		DelegationCreationBaseGasFee string   `json:"delegation_creation_base_gas_fee"`
		AllowListExpirationHeight    string   `json:"allow_list_expiration_height"`
		BtcActivationHeight          int      `json:"btc_activation_height"`
	} `json:"params"`
}
