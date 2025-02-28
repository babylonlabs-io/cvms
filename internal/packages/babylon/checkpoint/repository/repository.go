package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/cosmostation/cvms/internal/common"
	idxmodel "github.com/cosmostation/cvms/internal/common/indexer/model"
	indexerrepo "github.com/cosmostation/cvms/internal/common/indexer/repository"
	dbhelper "github.com/cosmostation/cvms/internal/helper/db"
	"github.com/cosmostation/cvms/internal/packages/babylon/checkpoint/model"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type CheckpointIndexerRepository struct {
	IndexName  string
	sqlTimeout time.Duration
	*bun.DB
	indexerrepo.IMetaRepository
}

func NewRepository(indexerDB common.IndexerDB, indexName string, sqlTimeout time.Duration) CheckpointIndexerRepository {
	// Instantiate the meta repository
	metarepo := indexerrepo.NewMetaRepository(indexerDB)

	// Return a repository that implements both IMetaRepository and vote-specific logic
	return CheckpointIndexerRepository{indexName, sqlTimeout, indexerDB.DB, metarepo}
}

func (repo *CheckpointIndexerRepository) GetLastEpoch() (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), repo.sqlTimeout)
	defer cancel()

	bve := &model.BabylonVoteExtension{}
	err := repo.
		NewSelect().
		Model(bve).
		Order("height DESC").
		Limit(1).
		Scan(ctx)
	if err != nil && err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil && err != sql.ErrNoRows {
		return 0, errors.Wrapf(err, "failed to select last babylon vote extension")
	}

	return bve.Epoch, nil
}

func (repo *CheckpointIndexerRepository) InsertBabylonVoteExtensionList(chainInfoID int64, indexPointerHeight int64, bveList []model.BabylonVoteExtension) error {
	ctx, cancel := context.WithTimeout(context.Background(), repo.sqlTimeout)
	defer cancel()

	// if there are not any miss validators in this block, just update index pointer
	if len(bveList) == 0 {
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

	// insert miss validators for this block and update index pointer in one transaction
	err := repo.RunInTx(
		ctx,
		nil,
		func(ctx context.Context, tx bun.Tx) error {
			_, err := tx.NewInsert().
				Model(&bveList).
				ExcludeColumn("id").
				Exec(ctx)
			if err != nil {
				return errors.Wrapf(err, "failed to insert validator_miss list")
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
		return errors.Wrapf(err, "failed to exec validator miss in a transaction: %v", bveList)
	}

	return nil
}

func (repo *CheckpointIndexerRepository) DeleteOldValidatorExtensionVoteList(chainID, retentionPeriod string) (
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
		Model((*model.BabylonVoteExtension)(nil)).
		ModelTableExpr(partitionTableName).
		Where("timestamp < ?", cutoffTime).
		Exec(ctx)
	if err != nil {
		return 0, err
	}

	rowsAffected, _ := res.RowsAffected()
	return rowsAffected, nil
}

/*
	select vi.moniker,
			COUNT(CASE WHEN status = 0 THEN 0 END) AS "unknown",
			COUNT(CASE WHEN status = 1 THEN 1 END) AS "absent",
			COUNT(CASE WHEN status = 2 THEN 1 END) AS "commit",
			COUNT(CASE WHEN status = 3 THEN 1 END) AS "nil"
	from babylon_checkpoint_bbn_test_5 idx
	join meta.validator_info vi on idx.validator_hex_address_id = vi.id
	GROUP BY vi.moniker
*/

func (repo *CheckpointIndexerRepository) SelectTotalMissList(chainID string) ([]model.TotalBabylonVoteExtensionByMoniker, error) {
	ctx, cancel := context.WithTimeout(context.Background(), repo.sqlTimeout)
	defer cancel()

	// Make partition table name
	partitionTableName := dbhelper.MakePartitionTableName(repo.IndexName, chainID)

	// Make model
	rvvList := make([]model.TotalBabylonVoteExtensionByMoniker, 0)
	query := fmt.Sprintf(`
	select vi.moniker,
			COUNT(CASE WHEN status = 0 THEN 0 END) AS "unknown",
			COUNT(CASE WHEN status = 1 THEN 1 END) AS "absent",
			COUNT(CASE WHEN status = 2 THEN 1 END) AS "commit",
			COUNT(CASE WHEN status = 3 THEN 1 END) AS "nil"
	from %s idx
	join meta.validator_info vi on idx.validator_hex_address_id = vi.id
	GROUP BY vi.moniker;`, partitionTableName)
	err := repo.NewRaw(query).Scan(ctx, &rvvList)
	if err != nil {
		return nil, err
	}

	return rvvList, nil
}
