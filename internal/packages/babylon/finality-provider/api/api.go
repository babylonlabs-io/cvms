package api

import (
	"github.com/cosmostation/cvms/internal/common"
	commonapi "github.com/cosmostation/cvms/internal/common/api"
	commontypes "github.com/cosmostation/cvms/internal/common/types"
	"github.com/cosmostation/cvms/internal/packages/babylon/finality-provider/types"
	"github.com/pkg/errors"
)

func GetFinalityProviderUptime(exporter *common.Exporter) (types.BabylonFinalityProviderUptimeStatues, error) {
	// 1. get finality provider infos
	finalityProviderInfos, err := commonapi.GetBabylonFinalityProviderInfos(exporter.CommonClient)
	if err != nil {
		return types.BabylonFinalityProviderUptimeStatues{}, errors.Wrap(err, "failed to get babylon finality provider infos")
	}
	exporter.Debugf("got total finality providers: %d", len(finalityProviderInfos))

	// 2. get latest height
	latestBlockHeight, _, err := commonapi.GetStatus(exporter.CommonClient)
	if err != nil {
		return types.BabylonFinalityProviderUptimeStatues{}, errors.Wrap(err, "failed to get babylon latest height")
	}

	exporter.Debugf("got latest block height: %d", latestBlockHeight)

	// 3. get active finality providers.
	activeProviders, err := commonapi.GetActiveFinalityProviderByHeight(exporter.CommonClient, latestBlockHeight)
	if err != nil {
		return types.BabylonFinalityProviderUptimeStatues{}, errors.Wrap(err, "failed to get babylon active finality providers")
	}

	exporter.Debugf("got active finality providers: %d", len(activeProviders))

	// temp := addActiveStatus(activeProviders, finalityProviderInfos)
	// for _, i := range temp {
	// 	exporter.Debugf("finality providers: %v", i)
	// }

	// 5. get lity providers' uptime status
	finalityProviderUptimeStatus, err := commonapi.GetFinalityProviderUptime(exporter.CommonClient, addActiveStatus(activeProviders, finalityProviderInfos))
	if err != nil {
		return types.BabylonFinalityProviderUptimeStatues{}, errors.Wrap(err, "failed to get babylon finality providers uptime")
	}
	exporter.Debugf("got active finality providers uptime status: %d", len(finalityProviderUptimeStatus))

	// 6. get on-chain uptime parameter
	signedBlocksWindow, minSignedPerWindow, err := commonapi.GetBabylonFinalityProviderParams(exporter.CommonClient)
	if err != nil {
		return types.BabylonFinalityProviderUptimeStatues{}, errors.Wrap(err, "failed to get babylon finality provider parameters")
	}

	return types.BabylonFinalityProviderUptimeStatues{
		SignedBlocksWindow:      signedBlocksWindow,
		MinSignedPerWindow:      minSignedPerWindow,
		FinalityProvidersStatus: finalityProviderUptimeStatus,
	}, nil
}

func addActiveStatus(activeProviders []commontypes.FinalityProvider, finalityProviderInfos []commontypes.FinalityProviderInfo) []commontypes.FinalityProviderInfo {
	activeFpMap := make(map[string]bool, len(activeProviders))
	for _, item := range activeProviders {
		activeFpMap[item.BtcPkHex] = true
	}

	// Modify the original slice using index-based iteration
	for i := range finalityProviderInfos {
		if activeFpMap[finalityProviderInfos[i].BTCPK] {
			finalityProviderInfos[i].Active = true
		}
	}

	return finalityProviderInfos
}
