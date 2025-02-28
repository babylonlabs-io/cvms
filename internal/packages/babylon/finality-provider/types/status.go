package types

type BabylonFinalityProviderUptimeStatues struct {
	MinSignedPerWindow      float64
	SignedBlocksWindow      float64
	FinalityProvidersStatus []FinalityProviderUptimeStatus
}

type FinalityProviderUptimeStatus struct {
	Moniker            string
	Address            string
	BTCPK              string
	MissedBlockCounter float64
	Active             string
	Jailed             string
}
