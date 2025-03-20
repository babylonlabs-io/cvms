package indexer

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/cosmostation/cvms/internal/common"
)

const MessageTypeLabel = "msg_type"

var (
	TransactionCounter           *prometheus.CounterVec
	TransactionCounterMetricName = "transactions_count"

	FailedTransactionCounter           *prometheus.CounterVec
	FailedTransactionCounterMetricName = "transaction_failures_count"

	BlockGasUsedGauge           prometheus.Gauge
	BlockGasUsedGaugeMetricName = "block_gas_used"

	BlockGasWantedGauge           prometheus.Gauge
	BlockGasWantedGaugeMetricName = "block_gas_wanted"

	BlockTxsBytesGauge           prometheus.Gauge
	BlockTxsBytesGaugeMetricName = "block_txs_bytes"
)

func (idx *BDAIndexer) initMetrics() {
	TransactionCounter = idx.Factory.NewCounterVec(prometheus.CounterOpts{
		Namespace:   common.Namespace,
		Subsystem:   idx.subsystem,
		Name:        TransactionCounterMetricName,
		ConstLabels: idx.PackageLabels},
		[]string{
			MessageTypeLabel,
		})

	FailedTransactionCounter = idx.Factory.NewCounterVec(prometheus.CounterOpts{
		Namespace:   common.Namespace,
		Subsystem:   idx.subsystem,
		Name:        FailedTransactionCounterMetricName,
		ConstLabels: idx.PackageLabels},
		[]string{
			MessageTypeLabel,
		})

	BlockGasUsedGauge = idx.Factory.NewGauge(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   idx.subsystem,
		Name:        BlockGasUsedGaugeMetricName,
		ConstLabels: idx.PackageLabels,
	})

	BlockGasWantedGauge = idx.Factory.NewGauge(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   idx.subsystem,
		Name:        BlockGasWantedGaugeMetricName,
		ConstLabels: idx.PackageLabels,
	})

	BlockTxsBytesGauge = idx.Factory.NewGauge(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   idx.subsystem,
		Name:        BlockTxsBytesGaugeMetricName,
		ConstLabels: idx.PackageLabels,
	})
}

func (idx *BDAIndexer) updateRootMetrics(indexPointer int64, timestamp time.Time) {
	common.IndexPointer.With(idx.RootLabels).Set(float64(indexPointer))
	common.IndexPointerTimestamp.With(idx.RootLabels).Set((float64(timestamp.Unix())))
	idx.Debugf("update prometheus metrics to %d", indexPointer)
}

func (idx *BDAIndexer) updateIndexerMetrics(summary BlockDataSummary) {
	for msg, cnt := range summary.MessageCounts {
		TransactionCounter.
			With(prometheus.Labels{MessageTypeLabel: msg}).
			Add(float64(cnt))
	}

	for _, msg := range summary.Messages {
		if !msg.success {
			FailedTransactionCounter.
				With(prometheus.Labels{MessageTypeLabel: msg.messageType}).
				Inc()
		}
	}

	BlockGasUsedGauge.Set(float64(summary.TotalGasUsed))
	BlockGasWantedGauge.Set(float64(summary.TotalGasWanted))
	BlockTxsBytesGauge.Set(float64(summary.TotalTxsBytes))

	idx.Debugf("update prometheus metrics to %d", summary.BlockHeight)
}
