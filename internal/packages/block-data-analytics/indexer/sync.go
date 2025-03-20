package indexer

import (
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"github.com/cosmostation/cvms/internal/common/api"
	indexermodel "github.com/cosmostation/cvms/internal/common/indexer/model"
	indexertypes "github.com/cosmostation/cvms/internal/common/indexer/types"
	"github.com/cosmostation/cvms/internal/common/types"
	"github.com/cosmostation/cvms/internal/helper"
	"github.com/cosmostation/cvms/internal/packages/block-data-analytics/model"
	"github.com/pkg/errors"
)

type BlockDataSummary struct {
	BlockHeight     int64
	Timestamp       time.Time
	TotalTxsBytes   int64
	SuccessTxsCount int64
	FailedTxsCount  int64
	TotalGasWanted  int64
	TotalGasUsed    int64
	MessageCounts   map[string]int
	Messages        []TxMessage
}

type TxMessage struct {
	messageType string
	success     bool
}

func (idx *BDAIndexer) batchSync(lastIndexPoint int64) (
	/* new index pointer */ int64,
	/* error */ error,
) {
	// set starntHeight and endHeight for batch sync
	// NOTE: end height will use latest height - 10 for block expiry height
	startHeight := (lastIndexPoint + 1)
	endHeight := idx.Lh.LatestHeight
	if startHeight > endHeight {
		idx.Debugf("no need to sync from %d height to %d height, so it'll skip the logic", startHeight, endHeight)
		return lastIndexPoint, nil
	}

	// set limit at end-height in this batch sync logic
	if (endHeight - startHeight) > indexertypes.BatchSyncLimit {
		endHeight = startHeight + indexertypes.BatchSyncLimit
		idx.Debugf("by batch sync limit, end height will change to %d", endHeight)
	}

	// init channel and waitgroup for go-routine
	ch := make(chan helper.Result)
	var wg sync.WaitGroup

	// init list
	summaryList := make(map[int64]BlockDataSummary, 0)
	for height := startHeight; height <= endHeight; height++ {
		wg.Add(1)
		height := height

		go func(ch chan helper.Result) {
			defer helper.HandleOutOfNilResponse(idx.Entry)
			defer wg.Done()

			_, timestamp, _, txs, _, _, err := api.GetBlock(idx.CommonClient, height)
			if err != nil {
				idx.Errorf("failed to call at %d height data, %s", height, err)
				ch <- helper.Result{Item: nil, Success: false}
				return
			}

			_, _, blockData, err := api.GetBlockResults(idx.CommonClient, height)
			if err != nil {
				idx.Errorf("failed to call at %d height data, %s", height, err)
				ch <- helper.Result{Item: nil, Success: false}
				return
			}

			var decodedTxs []types.CosmosTx
			if len(txs) > 0 {
				_, _, decodedTxs, err = api.GetBlockAndTxs(idx.CommonClient, height)
				if err != nil {
					idx.Errorf("failed to call at %d height txs, %s", height, err)
					ch <- helper.Result{Item: nil, Success: false}
					return
				}
			}

			blockDataAnalysis, err := makeBlockDataSummary(height, timestamp, blockData, txs, decodedTxs)
			if err != nil {
				idx.Errorf("failed to make block data analysis at %d height, %s", height, err)
				ch <- helper.Result{Item: nil, Success: false}
				return
			}

			ch <- helper.Result{
				Item:    blockDataAnalysis,
				Success: true,
			}
		}(ch)

		time.Sleep(10 * time.Millisecond)
	}

	// close channel
	go func() {
		wg.Wait()
		close(ch)
	}()

	// collect block summary data into block summary list
	errorCount := 0
	for r := range ch {
		if r.Success {
			item := r.Item.(BlockDataSummary)
			summaryList[item.BlockHeight] = item
			continue
		}
		errorCount++
	}

	// check error count
	if errorCount > 0 {
		return lastIndexPoint, errors.Errorf("failed to collect batch data, total errors: %d", errorCount)
	}

	isNewMessageType := false
	newMessageTypeMap := make(map[string]bool)
	for _, summary := range summaryList {
		for msgType := range summary.MessageCounts {
			_, exist := idx.Vim[msgType]
			if !exist {
				idx.Debugf("found new message type: %s", msgType)
				newMessageTypeMap[msgType] = true
				isNewMessageType = true
			}
		}
	}

	if isNewMessageType {
		newMessageTypeList := make([]indexermodel.MessageType, 0)
		for msgType := range newMessageTypeMap {
			newMessageTypeList = append(newMessageTypeList, indexermodel.MessageType{
				ChainInfoID: idx.ChainInfoID,
				MessageType: msgType,
			})
		}

		err := idx.InsertMessageTypeList(newMessageTypeList)
		if err != nil {
			idx.FetchValidatorInfoList()
			return lastIndexPoint, errors.WithStack(err)
		}

		messageTypeList, err := idx.GetMessageTypeListByChainInfoID(idx.ChainInfoID)
		if err != nil {
			return lastIndexPoint, errors.Wrap(err, "failed to reload message type list")

		}

		for _, msgType := range messageTypeList {
			idx.Vim[msgType.MessageType] = int64(msgType.ID)
		}

		idx.Debugf("changed vim length: %d", len(idx.Vim))
	}

	modelList1 := make([]model.BlockDataAnalytics, 0)
	modelList2 := make([]model.BlockMessageAnalytics, 0)
	for height := startHeight; height <= endHeight; height++ {
		summary, exist := summaryList[height]
		if !exist {
			idx.Errorf("failed to get %d height data", height)
			continue
		}

		model1 := model.BlockDataAnalytics{
			ChainInfoID:     idx.ChainInfoID,
			Height:          summary.BlockHeight,
			Timestamp:       summary.Timestamp,
			TotalTxsBytes:   summary.TotalTxsBytes,
			TotalGasUsed:    summary.TotalGasUsed,
			TotalGasWanted:  summary.TotalGasWanted,
			SuccessTxsCount: summary.SuccessTxsCount,
			FailedTxsCount:  summary.FailedTxsCount,
		}
		modelList1 = append(modelList1, model1)

		for _, msg := range summary.Messages {
			model2 := model.BlockMessageAnalytics{
				ChainInfoID:   idx.ChainInfoID,
				Height:        summary.BlockHeight,
				Timestamp:     summary.Timestamp,
				MessageTypeID: idx.Vim[msg.messageType],
				Success:       msg.success,
			}
			modelList2 = append(modelList2, model2)
		}

		// update indexer specific metrics
		idx.updateIndexerMetrics(summary)
	}

	newIndexPointer := summaryList[endHeight].BlockHeight
	newIndexPointerTime := summaryList[endHeight].Timestamp
	err := idx.InsertBlockDataList(idx.ChainInfoID, newIndexPointer, modelList1, modelList2)
	if err != nil {
		return lastIndexPoint, errors.Wrapf(err, "failed to insert from %d to %d height", startHeight, endHeight)
	}

	// update metrics
	idx.updateRootMetrics(newIndexPointer, newIndexPointerTime)
	return newIndexPointer, nil
}

func makeBlockDataSummary(height int64, timestamp time.Time, blockData types.CosmosBlockData, txs []types.Tx, decodedTxs []types.CosmosTx) (BlockDataSummary, error) {
	totalBytes := 0
	sucessCnt := 0
	failedCnt := 0
	totalGasWanted := int64(0)
	totalGasUsed := int64(0)
	msgCounts := make(map[string]int)
	msgList := make([]TxMessage, 0)

	for idx, tr := range blockData.TxResults {
		if tr.Code == 0 {
			sucessCnt++
		} else {
			failedCnt++
		}

		messages := ExtractMessageTypes(decodedTxs[idx])
		for _, msg := range messages {
			msgCounts[msg]++

			if tr.Code == 0 {
				msgList = append(msgList, TxMessage{messageType: msg, success: true})
			} else {
				msgList = append(msgList, TxMessage{messageType: msg, success: false})
			}
		}

		gasUsed, err := strconv.ParseInt(tr.GasUsed, 10, 64)
		if err != nil {
			errors.Wrap(err, "failed to parse gas used")
		}

		gasWanted, err := strconv.ParseInt(tr.GasWanted, 10, 64)
		if err != nil {
			errors.Wrap(err, "failed to parse gas wanted")
		}

		totalGasUsed += gasUsed
		totalGasWanted += gasWanted
	}

	for _, tx := range txs {
		totalBytes += len(tx)
	}

	return BlockDataSummary{
		BlockHeight:     height,
		Timestamp:       timestamp,
		TotalTxsBytes:   int64(totalBytes),
		SuccessTxsCount: int64(sucessCnt),
		FailedTxsCount:  int64(failedCnt),
		TotalGasWanted:  totalGasWanted,
		TotalGasUsed:    totalGasUsed,
		MessageCounts:   msgCounts,
		Messages:        msgList,
	}, nil
}

// ExtractMessageTypes extracts the @type values from transaction messages
func ExtractMessageTypes(tx types.CosmosTx) []string {
	var messageTypes []string

	for _, message := range tx.Body.Messages {
		messageTypes = append(messageTypes, mustExtractType(message))
	}

	return messageTypes
}

// NOTE: we expected already response data always should be json marshal
func mustExtractType(message json.RawMessage) string {
	var preResult map[string]json.RawMessage
	json.Unmarshal(message, &preResult)

	var typeValue string
	if rawType, ok := preResult["@type"]; ok {
		json.Unmarshal(rawType, &typeValue)
	}
	return typeValue
}
