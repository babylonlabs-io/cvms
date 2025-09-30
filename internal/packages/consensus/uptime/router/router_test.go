package router

import (
	"os"
	"testing"

	"github.com/cosmostation/cvms/internal/common"
	"github.com/cosmostation/cvms/internal/testutil"

	"github.com/stretchr/testify/assert"
)

func TestGetStatus(t *testing.T) {
	_ = testutil.SetupForTest()
	testMoniker := os.Getenv("TEST_MONIKER")
	testCases := []struct {
		testingName      string
		chainName        string
		protocolType     string
		chainID          string
		isConsumer       bool
		endpoint         common.Endpoints
		providerEndpoint common.Endpoints
		expectResult     float64
	}{
		{
			testingName:  "Cosmos Chain",
			chainName:    "cosmos",
			protocolType: "cosmos",
			isConsumer:   false,
			endpoint: common.Endpoints{
				RPCs: []string{os.Getenv("TEST_COSMOS_RPC_ENDPOINT")},
				APIs: []string{os.Getenv("TEST_COSMOS_API_ENDPOINT")},
			},
		},
		{
			testingName:  "Stride Chain",
			chainName:    "stride",
			protocolType: "cosmos",
			isConsumer:   true,
			chainID:      "stride-1",
			endpoint: common.Endpoints{
				RPCs: []string{os.Getenv("TEST_STRIDE_RPC_ENDPOINT")},
				APIs: []string{os.Getenv("TEST_STRIDE_API_ENDPOINT")},
			},
			providerEndpoint: common.Endpoints{
				RPCs: []string{os.Getenv("TEST_COSMOS_RPC_ENDPOINT")},
				APIs: []string{os.Getenv("TEST_COSMOS_API_ENDPOINT")},
			},
		},
		{
			testingName:  "Union Chain",
			chainName:    "union",
			protocolType: "cosmos",
			isConsumer:   false,
			chainID:      "union-testnet-8",
			endpoint: common.Endpoints{
				RPCs: []string{os.Getenv("TEST_UNION_RPC_ENDPOINT")},
				APIs: []string{os.Getenv("TEST_UNION_API_ENDPOINT")},
			},
		},
		{
			testingName:  "Initia Chain",
			chainName:    "initia",
			protocolType: "cosmos",
			isConsumer:   false,
			chainID:      "interwoven-1",
			endpoint: common.Endpoints{
				RPCs: []string{os.Getenv("TEST_INITIA_RPC_ENDPOINT")},
				APIs: []string{os.Getenv("TEST_INITIA_API_ENDPOINT")},
			},
		},
	}

	for _, tc := range testCases {
		exporter := testutil.GetTestExporter()
		t.Run(tc.testingName, func(t *testing.T) {
			if !assert.NotEqualValues(t, tc.endpoint, "") {
				// endpoint is empty
				t.FailNow()
			}

			// setup
			exporter.SetRPCEndPoint(tc.endpoint.RPCs[0])
			exporter.SetAPIEndPoint(tc.endpoint.APIs[0])
			exporter.ChainName = tc.chainName
			// additional setup for ics
			if tc.isConsumer {
				optionalClient := common.NewOptionalClient(exporter.Entry)
				optionalClient.SetRPCEndPoint(tc.providerEndpoint.RPCs[0])
				optionalClient.SetAPIEndPoint(tc.providerEndpoint.APIs[0])
				exporter.OptionalClient = optionalClient
			}

			// start test
			status, err := GetStatus(exporter, common.Packager{
				ProtocolType: tc.protocolType,
				ChainID:      tc.chainID,
				OptionPackager: common.OptionPackager{
					IsConsumerChain: tc.isConsumer,
				}})
			if err != nil {
				t.Fatalf("%s: %s", tc.testingName, err.Error())
			}

			for _, validator := range status.Validators {
				if validator.Moniker == testMoniker {
					t.Log("moniker                           :", validator.Moniker)
					t.Log("validator operator address        :", validator.ValidatorOperatorAddress)
					t.Log("validator consensus address       :", validator.ValidatorConsensusAddress)
					t.Log("validator proposer address        :", validator.ProposerAddress)
					t.Log("validator miss count              :", validator.MissedBlockCounter)
					if tc.isConsumer {
						t.Log("validator consumer valcons address:", validator.ConsumerConsensusAddress)
					}
				}
			}
		})
	}
}
