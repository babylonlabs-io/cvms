package repository

import (
	"context"
	"time"

	"github.com/cosmostation/cvms/internal/common"
	idxmodel "github.com/cosmostation/cvms/internal/common/indexer/model"
	indexerrepo "github.com/cosmostation/cvms/internal/common/indexer/repository"
	dbhelper "github.com/cosmostation/cvms/internal/helper/db"
	"github.com/cosmostation/cvms/internal/packages/babylon-btc-lightclient/model"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type BTCLightClientIndexerRepository struct {
	IndexName  string
	sqlTimeout time.Duration
	*bun.DB
	indexerrepo.IMetaRepository
}

func NewRepository(indexerDB common.IndexerDB, indexName string, sqlTimeout time.Duration) BTCLightClientIndexerRepository {
	// Instantiate the meta repository
	metarepo := indexerrepo.NewMetaRepository(indexerDB)

	// Return a repository that implements both IMetaRepository and vote-specific logic
	return BTCLightClientIndexerRepository{indexName, sqlTimeout, indexerDB.DB, metarepo}
}

func (repo *BTCLightClientIndexerRepository) InsertBabylonBTCRollList(chainInfoID int64, indexPointerHeight int64, modelList []model.BabylonBTCRoll) error {
	ctx, cancel := context.WithTimeout(context.Background(), repo.sqlTimeout)
	defer cancel()

	// if nothing to insert, just update index pointer
	if len(modelList) == 0 {
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

	// if there are records need to insert, insert records and index pointer in a tx
	err := repo.RunInTx(
		ctx,
		nil,
		func(ctx context.Context, tx bun.Tx) error {
			_, err := tx.NewInsert().
				Model(&modelList).
				ExcludeColumn("id").
				Exec(ctx)
			if err != nil {
				return errors.Wrapf(err, "failed to insert model list")
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
		return errors.Wrapf(err, "failed to update records and index point in a transaction: %v", modelList)
	}

	return nil
}

func (repo *BTCLightClientIndexerRepository) DeleteOldRecords(chainID, retentionPeriod string) (
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

	// Query Execution
	res, err := repo.NewDelete().
		Model((*model.BabylonBTCRoll)(nil)).
		ModelTableExpr(partitionTableName).
		Where("timestamp < ?", cutoffTime).
		Exec(ctx)
	if err != nil {
		return 0, err
	}

	rowsAffected, _ := res.RowsAffected()
	return rowsAffected, nil
}
