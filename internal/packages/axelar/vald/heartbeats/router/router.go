package router

import (
	"github.com/cosmostation/cvms/internal/common"
	"github.com/cosmostation/cvms/internal/packages/axelar/vald/heartbeats/api"
	"github.com/cosmostation/cvms/internal/packages/axelar/vald/heartbeats/parser"
	"github.com/cosmostation/cvms/internal/packages/axelar/vald/heartbeats/types"
)

func GetHeartbeats(exporter *common.Exporter, chainName string, latestHeartbeatsHeight int64) (types.CommonAxelarHeartbeats, error) {
	var (
		commonProxyResisterQueryPath string
		commonProxyResisterParser    func(resp []byte) (types.AxelarProxyResisterStatus, error)
	)

	switch chainName {
	case "axelar":
		commonProxyResisterQueryPath = types.AxelarProxyResisterQueryPath
		commonProxyResisterParser = parser.AxelarProxyResisterParser

		return api.GetAxelarHeartbeatsStatus(
			exporter,
			commonProxyResisterQueryPath, commonProxyResisterParser,
			latestHeartbeatsHeight,
		)

	default:
		return types.CommonAxelarHeartbeats{}, common.ErrOutOfSwitchCases
	}
}
