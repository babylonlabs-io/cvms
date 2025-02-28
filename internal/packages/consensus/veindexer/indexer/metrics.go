package indexer

import (
	"time"

	"github.com/cosmostation/cvms/internal/common"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	RecentMissCounterMetricName = "recent_miss_counter"
)

func (vidx *VEIndexer) initLabelsAndMetrics() {
	recentMissCounterMetric := vidx.Factory.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   subsystem,
		Name:        RecentMissCounterMetricName,
		ConstLabels: vidx.PackageLabels,
	}, []string{
		common.MonikerLabel,
	})

	vidx.MetricsVecMap[RecentMissCounterMetricName] = recentMissCounterMetric
}

func (vidx *VEIndexer) updateRecentMissCounterMetric() {
	rveList, err := vidx.repo.SelectRecentValidatorExtensionVoteList(vidx.ChainID)
	if err != nil {
		vidx.Errorf("failed to update recent miss counter metric: %s", err)
	}

	for _, rve := range rveList {
		missCount := (rve.UnknownCount + rve.AbsentCount + rve.NilCount)
		vidx.MetricsVecMap[RecentMissCounterMetricName].
			With(prometheus.Labels{common.MonikerLabel: rve.Moniker}).
			Set(float64(missCount))
	}
}

func (idx *VEIndexer) updateRootMetrics(indexPointer int64, indexPointerTimestamp time.Time) {
	common.IndexPointer.With(idx.RootLabels).Set(float64(indexPointer))
	common.IndexPointerTimestamp.With(idx.RootLabels).Set((float64(indexPointerTimestamp.Unix())))
	idx.Debugf("update prometheus metrics %d height", indexPointer)
}
