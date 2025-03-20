package collector

import (
	"github.com/cosmostation/cvms/internal/common"
	commonapi "github.com/cosmostation/cvms/internal/common/api"
	"github.com/pkg/errors"
)

func getStatus(exporter *common.Exporter) (CosmosBlockParamsStatus, error) {
	maxBytes, maxGas, err := commonapi.GetCosmosConsensusParams(exporter.CommonClient)
	if err != nil {
		return CosmosBlockParamsStatus{}, errors.Wrap(err, "failed to get cosmos consensus param")
	}

	return CosmosBlockParamsStatus{MaxBytes: maxBytes, MaxGas: maxGas}, nil
}
