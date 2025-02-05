package indexer

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/cosmostation/cvms/internal/common"
	"github.com/cosmostation/cvms/internal/common/api"
	commonparser "github.com/cosmostation/cvms/internal/common/parser"
	"github.com/cosmostation/cvms/internal/helper/logger"
	"github.com/cosmostation/cvms/internal/packages/duty/finality-provider-indexer/repository"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

const BabylonBaseURL = "https://lcd-office.cosmostation.io/babylon-testnet"

var (
	p = common.Packager{
		ChainName:    "babylon",
		ChainID:      "bbn-test-5",
		ProtocolType: "cosmos",
		Endpoints: common.Endpoints{
			RPCs: []string{"https://rpc-office.cosmostation.io/babylon-testnet"},
			APIs: []string{"https://lcd-office.cosmostation.io/babylon-testnet"},
		},
		Logger: logger.GetTestLogger(),
	}
)

func TestStart(t *testing.T) {
	waitingDuration := 60 * time.Second
	// Step 1: Set up the database
	tempDBName := "temp"
	indexerDB, err := common.NewTestLoaclIndexerDB(tempDBName)
	assert.NoError(t, err)

	p.SetIndexerDB(indexerDB)
	p.SetRetentionTime("1h")
	p.IsConsumerChain = false

	// Step 2: Initialize the CheckpointIndexer
	idx, err := NewFinalityProviderIndexer(p)
	assert.NoError(t, err)

	// Modify Start() to accept a callback for testing purposes
	err = idx.Start()
	assert.NoError(t, err)

	// Step 4: Wait for the goroutine to finish
	done := make(chan bool)

	go func() {
		// Simulate some work
		time.Sleep(waitingDuration)
		done <- true
	}()

	<-done
	t.Log("Goroutine finished work")
}
func TestBatchSync(t *testing.T) {
	tempDBName := "temp"
	indexerDB, err := common.NewTestLoaclIndexerDB(tempDBName)
	p.SetIndexerDB(indexerDB)
	p.IsConsumerChain = false
	assert.NoError(t, err)

	idx, err := NewFinalityProviderIndexer(p)
	assert.NoError(t, err)

	err = idx.InitChainInfoID()
	assert.NoError(t, err)

	err = idx.repo.InitPartitionTablesByChainInfoID(repository.IndexName, idx.ChainID, 100)
	assert.NoError(t, err)

	err = idx.repo.CreateFinalityProviderInfoPartitionTableByChainID(idx.ChainID)
	assert.NoError(t, err)

	err = idx.FetchValidatorInfoList()
	assert.NoError(t, err)

	newIndexPointer, err := idx.batchSync(94810)
	assert.NoError(t, err)
	t.Logf("new index point: %d", newIndexPointer)
}

func TestGetFinalityProvidersInfo(t *testing.T) {
	app := common.NewCommonApp(p)
	app.SetAPIEndPoint(BabylonBaseURL)
	fpInfoList, err := api.GetBabylonFinalityProviderInfos(app.CommonClient)
	assert.NoError(t, err)
	t.Logf("new fp infos: %d", len(fpInfoList))
	for idx, fp := range fpInfoList {
		t.Logf("fp infos: \n%v", fp)

		if idx == 0 {
			break
		}
	}
}

const TestingHeight = 93329

func TestCheckNonVotingFds(t *testing.T) {
	requester := resty.New().
		SetBaseURL(BabylonBaseURL).
		R().
		SetContext(context.Background())
	resp, err := requester.Get(fmt.Sprintf("/babylon/finality/v1/finality_providers/%d", TestingHeight))
	assert.NoErrorf(t, err, "error from : %s", resp.Request.URL)

	fps, err := commonparser.ParseFinalityProviders(resp.Body())
	assert.NoErrorf(t, err, "error from : %s", resp.Request.URL)

	// make a map by fp votes
	fpVoteMap := make(map[string]bool, len(fps.FinalityProviders))
	for _, fp := range fps.FinalityProviders {
		// t.Logf("add a pk_hex into map: %s", fp.BtcPkHex)
		fpVoteMap[fp.BtcPkHex] = false
	}

	resp, err = requester.Get(fmt.Sprintf("/babylon/finality/v1/votes/%d", TestingHeight))
	assert.NoErrorf(t, err, "error from : %s", resp.Request.URL)

	votes, err := commonparser.ParseFinalityProviderVotings(resp.Body())
	assert.NoError(t, err)

	// if the pk is existed in the votings, update the value for fp
	trueCnt := 0
	for _, pk := range votes.BTCPKs {
		fpVoteMap[pk] = true
		trueCnt++
	}

	// t.Logf("%v", fpVoteMap)

	// Convert the map to JSON and log it
	jsonData, err := json.MarshalIndent(fpVoteMap, "", "  ") // Indent with 2 spaces
	if err != nil {
		t.Logf("Error converting map to JSON: %v", err)
	} else {
		t.Logf("fpVoteMap for %d TestingHeight: \n%s", TestingHeight, jsonData)
	}

	total := len(fps.FinalityProviders)
	t.Logf("total %d | true :%d / false: %d", total, trueCnt, (total - trueCnt))

}
