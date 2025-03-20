package api

import (
	"testing"

	"github.com/cosmostation/cvms/internal/common"
	"github.com/stretchr/testify/assert"
)

func Test_Cosmos_GetCosmosConsensusParams(t *testing.T) {
	commonApp := common.NewCommonApp(p)
	commonApp.SetAPIEndPoint("https://lcd-office.cosmostation.io/babylon-testnet")

	maxBytes, maxGas, err := GetCosmosConsensusParams(commonApp.CommonClient)
	assert.NoError(t, err)
	assert.NotZero(t, maxBytes)
	assert.NotZero(t, maxGas)

	t.Logf("block max bytes: %.f", maxBytes)
	t.Logf("block max gas: %.f", maxGas)
}
