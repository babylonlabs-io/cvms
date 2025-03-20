package repository

import (
	"context"
	"time"

	"github.com/cosmostation/cvms/internal/common"
	idxmodel "github.com/cosmostation/cvms/internal/common/indexer/model"
	indexerrepo "github.com/cosmostation/cvms/internal/common/indexer/repository"
	dbhelper "github.com/cosmostation/cvms/internal/helper/db"

	"github.com/pkg/errors"

	"github.com/cosmostation/cvms/internal/packages/block-data-analytics/model"
	"github.com/uptrace/bun"
)

type BDAIndexerRepository struct {
	IndexName  string
	sqlTimeout time.Duration
	*bun.DB
	indexerrepo.IMetaRepository
}

func NewRepository(indexerDB common.IndexerDB, indexName string, sqlTimeout time.Duration) BDAIndexerRepository {
	metarepo := indexerrepo.NewMetaRepository(indexerDB)
	return BDAIndexerRepository{indexName, sqlTimeout, indexerDB.DB, metarepo}
}

func (repo *BDAIndexerRepository) InsertBlockDataList(
	chainInfoID int64,
	indexPointerHeight int64,
	blockDataList []model.BlockDataAnalytics,
	blockMessageList []model.BlockMessageAnalytics,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), repo.sqlTimeout)
	defer cancel()

	// if there are not any txs in this block, just update index pointer
	if len(blockDataList) == 0 {
		_, err := repo.
			NewUpdate().
			Model(&idxmodel.IndexPointer{}).
			Set("pointer = ?", indexPointerHeight).
			Where("chain_info_id = ?", chainInfoID).
			Where("index_name = ?", repo.IndexName).
			Exec(ctx)
		if err != nil {
			return errors.Wrapf(err, "failed to update new index pointer")
		}

		return nil
	}

	// insert records and udpate index pointer in one transaction
	err := repo.RunInTx(
		ctx,
		nil,
		func(ctx context.Context, tx bun.Tx) error {
			_, err := tx.NewInsert().
				Model(&blockDataList).
				ExcludeColumn("id").
				Exec(ctx)
			if err != nil {
				return errors.Wrapf(err, "failed to insert data")
			}

			_, err = tx.NewInsert().
				Model(&blockMessageList).
				ExcludeColumn("id").
				Exec(ctx)
			if err != nil {
				return errors.Wrapf(err, "failed to insert data")
			}

			_, err = tx.
				NewUpdate().
				Model(&idxmodel.IndexPointer{}).
				Set("pointer = ?", indexPointerHeight).
				Where("chain_info_id = ?", chainInfoID).
				Where("index_name = ?", repo.IndexName).
				Exec(ctx)
			if err != nil {
				return errors.Wrapf(err, "failed to update new index pointer")
			}

			return nil
		})

	if err != nil {
		return errors.Wrapf(err, "failed to insert data in a transaction")
	}

	return nil
}

func (repo *BDAIndexerRepository) DeleteOldRecords(chainID, retentionPeriod string) (
	/* deleted rows */ int64,
	/* unexpected error */ error,
) {
	ctx, cancel := context.WithTimeout(context.Background(), repo.sqlTimeout)
	defer cancel()

	// Parsing retention period
	duration, err := dbhelper.ParseRetentionPeriod(retentionPeriod)
	if err != nil {
		return 0, err
	}

	// Calculate cutoff time duration
	cutoffTime := time.Now().Add(duration)

	// Make partition table name
	partitionTableName := dbhelper.MakePartitionTableName(repo.IndexName, chainID)

	total := int64(0)
	// Query Execution
	res, err := repo.NewDelete().
		Model((*model.BlockDataAnalytics)(nil)).
		ModelTableExpr(partitionTableName).
		Where("timestamp < ?", cutoffTime).
		Exec(ctx)
	if err != nil {
		return 0, err
	}

	rowsAffected, _ := res.RowsAffected()
	total += (rowsAffected)

	res, err = repo.NewDelete().
		Model((*model.BlockMessageAnalytics)(nil)).
		ModelTableExpr(partitionTableName).
		Where("timestamp < ?", cutoffTime).
		Exec(ctx)
	if err != nil {
		return 0, err
	}

	rowsAffected, _ = res.RowsAffected()
	total += (rowsAffected)

	return rowsAffected, nil
}
