package indexer

import (
	"github.com/cosmostation/cvms/internal/common"
	"github.com/cosmostation/cvms/internal/common/api"
)

func (idx *FinalityProviderIndexer) updateRootMetrics(indexPointer int64) {
	common.IndexPointer.With(idx.RootLabels).Set(float64(indexPointer))
	_, timestamp, _, _, _, _, err := api.GetBlock(idx.CommonClient, indexPointer)
	if err != nil {
		idx.Errorf("failed to get block %d: %s", indexPointer, err)
		return
	}
	common.IndexPointerTimestamp.With(idx.RootLabels).Set((float64(timestamp.Unix())))
	idx.Debugf("update prometheus metrics %d epoch", indexPointer)

}
