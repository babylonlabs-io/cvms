package indexer

import (
	"github.com/cosmostation/cvms/internal/common"
	"github.com/cosmostation/cvms/internal/common/api"
	"github.com/prometheus/client_golang/prometheus"
)

func (idx *BTCLightClientIndexer) initLabelsAndMetrics() {
	epochMetric := idx.Factory.NewGauge(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   subsystem,
		Name:        common.IndexPointerEpochMetricName,
		ConstLabels: idx.PackageLabels,
	})
	idx.MetricsMap[common.IndexPointerEpochMetricName] = epochMetric

	timestampMetric := idx.Factory.NewGauge(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   subsystem,
		Name:        common.IndexPointerBlockTimestampMetricName,
		ConstLabels: idx.PackageLabels,
	})
	idx.MetricsMap[common.IndexPointerBlockTimestampMetricName] = timestampMetric
}

func (idx *BTCLightClientIndexer) updatePrometheusMetrics(indexPointer int64) {
	idx.MetricsMap[common.IndexPointerEpochMetricName].Set(float64(indexPointer))
	_, timestamp, _, _, _, _, err := api.GetBlock(idx.CommonClient, indexPointer)
	if err != nil {
		idx.Errorf("failed to get block %d: %s", indexPointer, err)
		return
	}
	idx.MetricsMap[common.IndexPointerBlockTimestampMetricName].Set((float64(timestamp.Unix())))
	idx.Debugf("update prometheus metrics %d height", indexPointer)
}
