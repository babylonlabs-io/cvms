package indexer

import (
	"time"

	"github.com/cosmostation/cvms/internal/common"
	"github.com/cosmostation/cvms/internal/packages/babylon/covenant-committee/model"
	"github.com/prometheus/client_golang/prometheus"
)

// metrics name for indexer
const (
	CovenantSigCountMetricName        = "covenant_sigs_count"
	BtcDelegationCountTotalMetricName = "btc_delegation_count_total"
)

func (idx *CovenantSignatureIndexer) initLabelsAndMetrics() {
	covenantSigMetric := idx.Factory.NewCounterVec(prometheus.CounterOpts{
		Namespace:   common.Namespace,
		Subsystem:   subsystem,
		Name:        CovenantSigCountMetricName,
		ConstLabels: idx.PackageLabels,
	}, []string{
		"btc_pk",
	})
	idx.MetricsCountVecMap[CovenantSigCountMetricName] = covenantSigMetric

	findBtcDelegationMetric := idx.Factory.NewCounter(prometheus.CounterOpts{
		Namespace: common.Namespace,
		Subsystem: subsystem,
		Name:      BtcDelegationCountTotalMetricName,
		Help:      "The total number of BTC delegations found at the time the indexer started.",
	})

	idx.MetricsCountMap[BtcDelegationCountTotalMetricName] = findBtcDelegationMetric
}

func (idx *CovenantSignatureIndexer) initMetricState(covenantCommitteeMap map[string]int64) {
	for btcPk, _ := range covenantCommitteeMap {
		covenantSigMetric, ok := idx.MetricsCountVecMap[CovenantSigCountMetricName]
		if ok {
			covenantSigMetric.WithLabelValues(btcPk).Add(0)
		}
	}

	btcDelegationMetrics, ok := idx.MetricsCountMap[BtcDelegationCountTotalMetricName]
	if ok {
		btcDelegationMetrics.Add(0)
	}
}

func (idx *CovenantSignatureIndexer) updateRootMetrics(indexPointer int64, indexPointerTimestamp time.Time) {
	common.IndexPointer.With(idx.RootLabels).Set(float64(indexPointer))
	common.IndexPointerTimestamp.With(idx.RootLabels).Set((float64(indexPointerTimestamp.Unix())))
	idx.Debugf("update prometheus metrics %d epoch", indexPointer)

}

func (idx *CovenantSignatureIndexer) updateIndexerMetrics(
	covenantSignatureList []model.BabylonCovenantSignature,
	btcDelegationsList []model.BabylonBtcDelegation,
) {
	covenantSigMetric, ok := idx.MetricsCountVecMap[CovenantSigCountMetricName]
	if ok {
		for _, sig := range covenantSignatureList {
			for btcPk, id := range idx.covenantCommitteeMap {
				if id == sig.CovenantBtcPkID {
					// add count
					covenantSigMetric.WithLabelValues(btcPk).Add(1)
				}
			}
		}
	}

	btcDelegationMetrics, ok := idx.MetricsCountMap[BtcDelegationCountTotalMetricName]
	if ok {
		btcDelegationMetrics.Add(float64(len(btcDelegationsList)))
	}
}
