package indexer

import (
	"github.com/cosmostation/cvms/internal/common"
	"github.com/cosmostation/cvms/internal/helper"
	"github.com/cosmostation/cvms/internal/helper/config"
	btclcindexer "github.com/cosmostation/cvms/internal/packages/babylon-btc-lightclient/indexer"
	bcindexer "github.com/cosmostation/cvms/internal/packages/consensus/babylon-checkpoint/indexer"
	veindexer "github.com/cosmostation/cvms/internal/packages/consensus/veindexer/indexer"
	voteindexer "github.com/cosmostation/cvms/internal/packages/consensus/voteindexer/indexer"
	fpindexer "github.com/cosmostation/cvms/internal/packages/duty/finality-provider-indexer/indexer"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
)

func selectPackage(
	m common.Mode, f promauto.Factory, l *logrus.Logger,
	idb *common.IndexerDB, mainnet bool, chainID, chainName, pkg, protocolType string,
	isConsumer bool,
	cc config.ChainConfig, monikers []string,
) error {

	// Add validation logic on each provided URL
	validAPIs := make([]string, 0)
	validRPCs := make([]string, 0)

	for _, node := range cc.Nodes {
		if helper.ValidateURL(node.RPC) {
			validRPCs = append(validRPCs, node.RPC)
		}

		if helper.ValidateURL(node.API) {
			validAPIs = append(validAPIs, node.API)
		}
	}

	providerRPCs := make([]string, 0)
	providerAPIs := make([]string, 0)
	if isConsumer {
		for _, node := range cc.ProviderNodes {
			if helper.ValidateURL(node.RPC) {
				providerRPCs = append(providerRPCs, node.RPC)
			}
			if helper.ValidateURL(node.API) {
				providerAPIs = append(providerAPIs, node.API)
			}
		}
	}

	switch {
	case pkg == "voteindexer":
		endpoints := common.Endpoints{RPCs: validRPCs, CheckRPC: true, APIs: validAPIs, CheckAPI: true}
		p, err := common.NewPackager(m, f, l, mainnet, chainID, chainName, pkg, protocolType, cc, endpoints, monikers...)
		if err != nil {
			return errors.Wrap(err, common.ErrFailedToBuildPackager)
		}
		p.SetIndexerDB(idb)
		if isConsumer {
			providerEndpoints := common.Endpoints{RPCs: providerRPCs, CheckRPC: true, APIs: providerAPIs, CheckAPI: true}
			p.SetAddtionalEndpoints(providerEndpoints)
			p.SetConsumer()
		}
		voteindexer, err := voteindexer.NewVoteIndexer(*p)
		if err != nil {
			return errors.Wrap(err, common.ErrFailedToBuildPackager)
		}
		return voteindexer.Start()
	case pkg == "veindexer":
		endpoints := common.Endpoints{RPCs: validRPCs, CheckRPC: true, APIs: validAPIs, CheckAPI: true}
		p, err := common.NewPackager(m, f, l, mainnet, chainID, chainName, pkg, protocolType, cc, endpoints, monikers...)
		if err != nil {
			return errors.Wrap(err, common.ErrFailedToBuildPackager)
		}
		p.SetIndexerDB(idb)
		if isConsumer {
			providerEndpoints := common.Endpoints{RPCs: providerRPCs, CheckRPC: true, APIs: providerAPIs, CheckAPI: true}
			p.SetAddtionalEndpoints(providerEndpoints)
			p.SetConsumer()
		}
		veindexer, err := veindexer.NewVEIndexer(*p)
		if err != nil {
			return errors.Wrap(err, common.ErrFailedToBuildPackager)
		}
		return veindexer.Start()
	case pkg == "babylon_checkpoint":
		endpoints := common.Endpoints{RPCs: validRPCs, CheckRPC: true, APIs: validAPIs, CheckAPI: true}
		p, err := common.NewPackager(m, f, l, mainnet, chainID, chainName, pkg, protocolType, cc, endpoints, monikers...)
		if err != nil {
			return errors.Wrap(err, common.ErrFailedToBuildPackager)
		}
		p.SetIndexerDB(idb)
		if isConsumer {
			providerEndpoints := common.Endpoints{RPCs: providerRPCs, CheckRPC: true, APIs: providerAPIs, CheckAPI: true}
			p.SetAddtionalEndpoints(providerEndpoints)
			p.SetConsumer()
		}
		bcindexer, err := bcindexer.NewCheckpointIndexer(*p)
		if err != nil {
			return errors.Wrap(err, common.ErrFailedToBuildPackager)
		}
		return bcindexer.Start()
	case pkg == "finality-provider-indexer":
		endpoints := common.Endpoints{RPCs: validRPCs, CheckRPC: true, APIs: validAPIs, CheckAPI: true}
		p, err := common.NewPackager(m, f, l, mainnet, chainID, chainName, pkg, protocolType, cc, endpoints, monikers...)
		if err != nil {
			return errors.Wrap(err, common.ErrFailedToBuildPackager)
		}
		p.SetIndexerDB(idb)
		if isConsumer {
			providerEndpoints := common.Endpoints{RPCs: providerRPCs, CheckRPC: true, APIs: providerAPIs, CheckAPI: true}
			p.SetAddtionalEndpoints(providerEndpoints)
			p.SetConsumer()
		}
		fpindexer, err := fpindexer.NewFinalityProviderIndexer(*p)
		if err != nil {
			return errors.Wrap(err, common.ErrFailedToBuildPackager)
		}
		return fpindexer.Start()
	case pkg == "babylon-btc-lightclient":
		endpoints := common.Endpoints{RPCs: validRPCs, CheckRPC: true, APIs: validAPIs, CheckAPI: true}
		p, err := common.NewPackager(m, f, l, mainnet, chainID, chainName, pkg, protocolType, cc, endpoints, monikers...)
		if err != nil {
			return errors.Wrap(err, common.ErrFailedToBuildPackager)
		}
		p.SetIndexerDB(idb)
		btclcindexer, err := btclcindexer.NewBTCLightClientIndexer(*p)
		if err != nil {
			return errors.Wrap(err, common.ErrFailedToBuildPackager)
		}
		return btclcindexer.Start()
	}

	return common.ErrUnSupportedPackage
}
