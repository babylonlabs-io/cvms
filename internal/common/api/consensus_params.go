package api

import (
	"context"
	"net/http"

	"github.com/cosmostation/cvms/internal/common"
	"github.com/cosmostation/cvms/internal/common/parser"
	"github.com/cosmostation/cvms/internal/common/types"
	"github.com/pkg/errors"
)

func GetCosmosConsensusParams(c common.CommonClient) (float64, float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), common.Timeout)
	defer cancel()

	requester := c.APIClient.R().SetContext(ctx)
	resp, err := requester.Get(types.CosmosConsensusParamsQueryPath)
	if err != nil {
		return 0, 0, errors.Errorf("rpc call is failed from %s: %s", resp.Request.URL, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return 0, 0, errors.Errorf("stanage status code from %s: [%d]", resp.Request.URL, resp.StatusCode())
	}

	maxBytes, maxGas, err := parser.CosmosConsensusmParamsParser(resp.Body())
	if err != nil {
		return 0, 0, errors.WithStack(err)
	}

	return maxBytes, maxGas, nil
}
