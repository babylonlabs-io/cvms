package indexer

import (
	"strconv"
	"testing"

	"github.com/cosmostation/cvms/internal/common"
	"github.com/cosmostation/cvms/internal/common/api"
	"github.com/cosmostation/cvms/internal/common/types"
	"github.com/cosmostation/cvms/internal/helper/logger"

	"github.com/stretchr/testify/assert"
)

var p = common.Packager{
	ChainName:    "babylon",
	ChainID:      "bbn-test-5",
	ProtocolType: "cosmos",
	Endpoints: common.Endpoints{
		RPCs: []string{"https://rpc-office.cosmostation.io/babylon-testnet"},
		APIs: []string{"https://lcd-office.cosmostation.io/babylon-testnet"},
	},
	Logger: logger.GetTestLogger(),
}

func Test_0_MakeLogic(t *testing.T) {
	app := common.NewCommonApp(p)
	app.SetRPCEndPoint(p.Endpoints.RPCs[0])
	app.SetAPIEndPoint(p.Endpoints.APIs[0])

	testHeights := []int64{538689}
	for _, h := range testHeights {
		_, timestamp, _, txs, _, _, err := api.GetBlock(app.CommonClient, h)
		assert.NoError(t, err)

		_ = timestamp

		_, _, blockData, err := api.GetBlockResults(app.CommonClient, h)
		assert.NoError(t, err)

		var decodedTxs []types.CosmosTx
		if len(txs) > 0 {
			_, _, decodedTxs, err = api.GetBlockAndTxs(app.CommonClient, h)
			assert.NoError(t, err)
		}

		totalBytes := 0
		sucessCnt := 0
		failedCnt := 0
		totalGasWanted := int64(0)
		totalGasUsed := int64(0)
		msgCounts := make(map[string]int)
		for idx, tr := range blockData.TxResults {
			if tr.Code == 0 {
				sucessCnt++
			} else {
				failedCnt++
			}

			messages := ExtractMessageTypes(decodedTxs[idx])
			for _, msg := range messages {
				t.Logf("msg type:: %s", msg)
				// Increment count by type
				msgCounts[msg]++
			}

			gasUsed, err := strconv.ParseInt(tr.GasUsed, 10, 64)
			assert.NoError(t, err)
			totalGasUsed += gasUsed

			gasWanted, err := strconv.ParseInt(tr.GasWanted, 10, 64)
			assert.NoError(t, err)
			totalGasWanted += gasWanted
		}

		// Get total byte size of all txs
		for _, tx := range txs {
			totalBytes += len(tx)
		}

		t.Logf("========= %d block data analysis ============", h)

		t.Logf("block max gas: %s", blockData.ConsensusParamUpdates.Block.MaxGas)
		t.Logf("block max bytes: %s", blockData.ConsensusParamUpdates.Block.MaxBytes)

		t.Logf("total bytes: %d", totalBytes)
		t.Logf("success tx: %d", sucessCnt)
		t.Logf("failed tx: %d", failedCnt)
		t.Logf("total gas used: %d", totalGasUsed)
		t.Logf("total gas wanted: %d", totalGasWanted)
		for msgType, cnt := range msgCounts {
			t.Logf("%s: %d", msgType, cnt)
		}
		t.Logf("=========================================")
	}
}

func Test_1_SummaryTest(t *testing.T) {
	app := common.NewCommonApp(p)
	app.SetRPCEndPoint(p.Endpoints.RPCs[0])
	app.SetAPIEndPoint(p.Endpoints.APIs[0])

	testHeights := []int64{538689}
	for _, h := range testHeights {
		_, timestamp, _, txs, _, _, err := api.GetBlock(app.CommonClient, h)
		assert.NoError(t, err)

		_, _, blockData, err := api.GetBlockResults(app.CommonClient, h)
		assert.NoError(t, err)

		var decodedTxs []types.CosmosTx
		if len(txs) > 0 {
			_, _, decodedTxs, err = api.GetBlockAndTxs(app.CommonClient, h)
			assert.NoError(t, err)
		}

		blockDataAnalysis, err := makeBlockDataSummary(h, timestamp, blockData, txs, decodedTxs)
		assert.NoError(t, err)

		// app.Infof("%v", blockDataAnalysis)
		app.Infof("fail: %v", blockDataAnalysis.FailedTxsCount)
		app.Infof("success: %v", blockDataAnalysis.SuccessTxsCount)
	}
}
