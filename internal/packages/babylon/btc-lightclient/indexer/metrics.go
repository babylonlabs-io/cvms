package indexer

import (
	"github.com/cosmostation/cvms/internal/common"
	"github.com/cosmostation/cvms/internal/common/api"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	EventCountMetricName = "event_count"
	BTCHeightMetricName  = "btc_height"
	EventTypeLabel       = "event_type"
)

func (idx *BTCLightClientIndexer) initMetrics() {
	idx.MetricsCountVecMap[EventCountMetricName] = idx.Factory.NewCounterVec(prometheus.CounterOpts{
		Namespace:   common.Namespace,
		Subsystem:   subsystem,
		Name:        EventCountMetricName,
		ConstLabels: idx.PackageLabels,
		Help:        "Count BTC Light Client events by event type",
	}, []string{EventTypeLabel})

	idx.MetricsMap[BTCHeightMetricName] = idx.Factory.NewGauge(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   subsystem,
		Name:        BTCHeightMetricName,
		ConstLabels: idx.PackageLabels,
		Help:        "The BTC Light Client height",
	})
}

func (idx *BTCLightClientIndexer) updateRootMetrics(indexPointer int64) {
	// update index pointer
	common.IndexPointer.With(idx.RootLabels).Set(float64(indexPointer))

	// update index pointer timestamp
	_, timestamp, _, _, _, _, err := api.GetBlock(idx.CommonClient, indexPointer)
	if err != nil {
		idx.Warnf("failed to update index pointer timestamp metric: %s", err)
		return
	}
	common.IndexPointerTimestamp.With(idx.RootLabels).Set((float64(timestamp.Unix())))
	idx.Debugf("update prometheus metrics %d height", indexPointer)
}

func (idx *BTCLightClientIndexer) updateIndexerMetrics(forwardCnt, backCnt, lastBTCHeight int64) {
	idx.MetricsCountVecMap[EventCountMetricName].With(prometheus.Labels{EventTypeLabel: "BTCRollForward"}).Add(float64(forwardCnt))
	idx.MetricsCountVecMap[EventCountMetricName].With(prometheus.Labels{EventTypeLabel: "BTCRollBack"}).Add(float64(backCnt))
	idx.MetricsMap[BTCHeightMetricName].Set(float64(lastBTCHeight))
}
