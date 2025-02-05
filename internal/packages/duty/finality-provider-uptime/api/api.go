package api

import (
	"github.com/cosmostation/cvms/internal/common"
	commonapi "github.com/cosmostation/cvms/internal/common/api"
	commontypes "github.com/cosmostation/cvms/internal/common/types"
	"github.com/cosmostation/cvms/internal/packages/duty/finality-provider-uptime/types"
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

	// 4. filter finality provider infos by active status
	newFinalityProviderInfos := filterActiveProviders(activeProviders, finalityProviderInfos)

	exporter.Debugf("call filter function active finality provider infos: %d", len(newFinalityProviderInfos))

	// 5. get lity providers' uptime status
	finalityProviderUptimeStatus, err := commonapi.GetFinalityProviderUptime(exporter.CommonClient, newFinalityProviderInfos)
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

func filterActiveProviders(activeProviders []commontypes.FinalityProvider, finalityProviderInfos []commontypes.FinalityProviderInfo) []commontypes.FinalityProviderInfo {
	// Create a map for quick lookup
	fpInfoMap := make(map[string]commontypes.FinalityProviderInfo, len(finalityProviderInfos))
	for _, item := range finalityProviderInfos {
		fpInfoMap[item.BTCPK] = item
	}

	// Create a new slice for filtered results
	newFinalityProviderInfos := make([]commontypes.FinalityProviderInfo, 0, len(activeProviders))
	for _, fp := range activeProviders {
		if info, exists := fpInfoMap[fp.BtcPkHex]; exists {
			newFinalityProviderInfos = append(newFinalityProviderInfos, info)
		}
	}

	return newFinalityProviderInfos
}
