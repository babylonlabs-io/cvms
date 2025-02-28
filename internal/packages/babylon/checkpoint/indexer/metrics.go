package indexer

import (
	"time"

	"github.com/cosmostation/cvms/internal/common"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	totalMissCounterMetricName  = "bls_signature_missed_total"
	latestMissCounterMetricName = "validator_bls_signature_missed_counter"

	epochLabel = "epoch"
)

func (idx *CheckpointIndexer) initLabelsAndMetrics() {
	// validator miss metrics
	idx.MetricsVecMap[totalMissCounterMetricName] = idx.Factory.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   subsystem,
		Name:        totalMissCounterMetricName,
		ConstLabels: idx.PackageLabels,
		Help:        "Total BLS signatures missed by a CometBFT Validator",
	}, []string{
		common.MonikerLabel,
		common.StatusLabel,
	})
	idx.MetricsVecMap[latestMissCounterMetricName] = idx.Factory.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   common.Namespace,
		Subsystem:   subsystem,
		Name:        latestMissCounterMetricName,
		ConstLabels: idx.PackageLabels,
		Help:        "Total amount of Validators that missed a BLS signature on the latest epoch",
	}, []string{
		common.MonikerLabel,
		common.StatusLabel,
		epochLabel,
	})
}

func (idx *CheckpointIndexer) updateRootMetrics(indexPointer int64, indexPointerTimestamp time.Time) {
	common.IndexPointer.With(idx.RootLabels).Set(float64(indexPointer))
	common.IndexPointerTimestamp.With(idx.RootLabels).Set((float64(indexPointerTimestamp.Unix())))
	idx.Debugf("update prometheus metrics %d epoch", indexPointer)

}

func (idx *CheckpointIndexer) updateIndexerMetrics() {
	modelList, err := idx.repo.SelectTotalMissList(idx.ChainID)
	if err != nil {
		idx.Errorf("failed to update recent miss counter metric: %s", err)
	}
	for _, model := range modelList {
		idx.MetricsVecMap[totalMissCounterMetricName].
			With(prometheus.Labels{common.MonikerLabel: model.Moniker, common.StatusLabel: "BLOCK_ID_FLAG_UNKNOWN"}).
			Set(float64(model.UnknownCount))

		idx.MetricsVecMap[totalMissCounterMetricName].
			With(prometheus.Labels{common.MonikerLabel: model.Moniker, common.StatusLabel: "BLOCK_ID_FLAG_ABSENT"}).
			Set(float64(model.AbsentCount))

		idx.MetricsVecMap[totalMissCounterMetricName].
			With(prometheus.Labels{common.MonikerLabel: model.Moniker, common.StatusLabel: "BLOCK_ID_FLAG_COMMIT"}).
			Set(float64(model.CommitCount))

		idx.MetricsVecMap[totalMissCounterMetricName].
			With(prometheus.Labels{common.MonikerLabel: model.Moniker, common.StatusLabel: "BLOCK_ID_FLAG_NIL"}).
			Set(float64(model.NilCount))
	}
}
