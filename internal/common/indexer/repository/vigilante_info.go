package repository

import (
	"context"
	"fmt"

	"github.com/cosmostation/cvms/internal/common/indexer/model"

	"github.com/cosmostation/cvms/internal/helper"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

const tableName = "vigilante_info"

func (repo *MetaRepository) CreateVigilanteInfoPartitionTableByChainID(chainID string) error {
	ctx := context.Background()
	defer ctx.Done()

	ci := &model.ChainInfo{}
	err := repo.
		NewSelect().
		Model(ci).
		Column("id").
		Where("chain_id = ?", chainID).
		Scan(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to select chain_info id by chain_id")
	}

	tableNameWithSuffix := fmt.Sprintf("meta.%s_%s", tableName, helper.ParseToSchemaName(chainID))
	query := fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS %s PARTITION OF "meta"."%s" FOR VALUES IN ('%d');`,
		tableNameWithSuffix, tableName, ci.ID,
	)

	_, err = repo.NewRaw(query).Exec(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create a new partition table")
	}

	return nil
}

func (repo *MetaRepository) GetVigilanteInfoListByChainInfoID(chainInfoID int64) ([]model.VigilanteInfo, error) {
	ctx := context.Background()
	defer ctx.Done()

	VigilanteInfoList := make([]model.VigilanteInfo, 0)
	err := repo.
		NewSelect().
		Model(&VigilanteInfoList).
		Where("chain_info_id = ?", chainInfoID).
		Scan(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query validator_info list by chain_info_id")
	}

	return VigilanteInfoList, nil
}

func (repo *MetaRepository) InsertVigilanteInfoList(VigilanteInfoList []model.VigilanteInfo) error {
	ctx := context.Background()
	defer ctx.Done()

	_, err := repo.NewInsert().
		Model(&VigilanteInfoList).
		ExcludeColumn("id").
		Exec(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to insert validator info list")
	}

	return nil
}

func (repo *MetaRepository) GetVigilanteInfoListByMonikers(chainInfoID int64, monikers []string) ([]model.VigilanteInfo, error) {
	ctx := context.Background()
	defer ctx.Done()

	VigilanteInfoList := make([]model.VigilanteInfo, 0)
	err := repo.
		NewSelect().
		Model(&VigilanteInfoList).
		ColumnExpr("*").
		Where("chain_info_id = ?", chainInfoID).
		Where("moniker in (?)", bun.In(monikers)).
		Scan(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query validator_info list by chain_info_id")
	}

	return VigilanteInfoList, nil
}
