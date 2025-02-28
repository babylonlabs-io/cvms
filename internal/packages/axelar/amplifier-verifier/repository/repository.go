package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/cosmostation/cvms/internal/common"
	idxmodel "github.com/cosmostation/cvms/internal/common/indexer/model"
	indexerrepo "github.com/cosmostation/cvms/internal/common/indexer/repository"
	dbhelper "github.com/cosmostation/cvms/internal/helper/db"
	"github.com/cosmostation/cvms/internal/packages/axelar/amplifier-verifier/model"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

type AmplifierIndexerRepository struct {
	IndexName  string
	sqlTimeout time.Duration
	*bun.DB
	indexerrepo.IMetaRepository
}

func NewRepository(indexerDB common.IndexerDB, indexName string, sqlTimeout time.Duration) AmplifierIndexerRepository {
	// Instantiate the meta repository
	metarepo := indexerrepo.NewMetaRepository(indexerDB)

	// Return a repository that implements both IMetaRepository and vote-specific logic
	return AmplifierIndexerRepository{indexName, sqlTimeout, indexerDB.DB, metarepo}
}

// InsertInitPollVoteList
func (repo *AmplifierIndexerRepository) InsertInitPollVoteList(chainInfoID int64, modelList []model.AxelarAmplifierVerifierVote) error {
	ctx, cancel := context.WithTimeout(context.Background(), repo.sqlTimeout)
	defer cancel()

	_, err := repo.NewInsert().
		Model(&modelList).
		ExcludeColumn("id").
		On("CONFLICT ON CONSTRAINT uniq_verifier_id_by_poll DO NOTHING").
		Exec(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to insert init poll vote list")
	}

	return nil
}

// UpdatePollVoteList
func (repo *AmplifierIndexerRepository) UpdatePollVoteList(
	chainInfoID int64,
	indexPointerHeight int64,
	AmplifierVoteList []model.AxelarAmplifierVerifierVote,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), repo.sqlTimeout)
	defer cancel()

	// if there are not any miss validators in this block, just update index pointer
	if len(AmplifierVoteList) == 0 {
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

	// insert miss validators for this block and udpate index pointer in one transaction
	err := repo.RunInTx(
		ctx,
		nil,
		func(ctx context.Context, tx bun.Tx) error {
			values := tx.NewValues(&AmplifierVoteList)
			_, err := tx.NewUpdate().
				With("_data", values).
				Model((*model.AxelarAmplifierVerifierVote)(nil)).
				ModelTableExpr("axelar_amplifier_verifier").
				TableExpr("_data").
				Set("status = _data.status").
				Set("poll_vote_height = _data.poll_vote_height").
				Where("axelar_amplifier_verifier.chain_info_id = _data.chain_info_id").
				Where("axelar_amplifier_verifier.chain_and_poll_id = _data.chain_and_poll_id").
				Where("axelar_amplifier_verifier.verifier_id = _data.verifier_id").
				Where("_data.poll_vote_height < axelar_amplifier_verifier.poll_start_height + 10"). // poll_vote_height < (poll_start_height + block expiry)
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
		return errors.Wrapf(err, "failed to exec validator miss in a transaction")
	}

	return nil
}

// select * from axelar_amplifier_verifier_axelar_dojo_1 aavad
// where aavad.chain_and_poll_id = 'sui/8'
func (repo *AmplifierIndexerRepository) SelectVerifierVoteList(chainAndPollID string) ([]model.AxelarAmplifierVerifierVote, error) {
	ctx, cancel := context.WithTimeout(context.Background(), repo.sqlTimeout)
	defer cancel()

	aav := make([]model.AxelarAmplifierVerifierVote, 0)
	err := repo.NewSelect().
		Model(&aav).
		Where("chain_and_poll_id = ?", chainAndPollID).
		Scan(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to select %s votes", repo.IndexName)
	}

	return aav, nil
}

func (repo *AmplifierIndexerRepository) DeleteOldValidatorExtensionVoteList(chainID, retentionPeriod string) (
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
		Model((*model.AxelarAmplifierVerifierVote)(nil)).
		ModelTableExpr(partitionTableName).
		Where("timestamp < ?", cutoffTime).
		Exec(ctx)
	if err != nil {
		return 0, err
	}

	rowsAffected, _ := res.RowsAffected()
	return rowsAffected, nil
}

// vi.moniker,
// COUNT(CASE WHEN status = 0 THEN 1 END) as "did_not_vote",
// STRING_AGG(CASE WHEN status = 0 THEN chain_and_poll_id END, ', ') AS "did_not_vote_polls",

// COUNT(CASE WHEN status = 1 THEN 1 END) as "failed_on_chain",
// STRING_AGG(CASE WHEN status = 1 THEN chain_and_poll_id END, ', ') AS "failed_on_chain_polls",

// COUNT(CASE WHEN status = 2 THEN 1 END) as "not_found",
// STRING_AGG(CASE WHEN status = 2 THEN chain_and_poll_id END, ', ') AS "not_found_polls",

// COUNT(CASE WHEN status = 3 THEN 1 END) as "succeeded_on_chain",
// STRING_AGG(CASE WHEN status = 3 THEN chain_and_poll_id END, ', ') AS "succeeded_on_chain_polls"
func (repo *AmplifierIndexerRepository) SelectPollVoteStatus(chainID string) ([]model.RecentVote, error) {
	ctx, cancel := context.WithTimeout(context.Background(), repo.sqlTimeout)
	defer cancel()

	partitionTableName := dbhelper.MakePartitionTableName(repo.IndexName, chainID)
	voteList := make([]model.RecentVote, 0)
	query := fmt.Sprintf(`
	SELECT 
		vi.moniker,
		COUNT(CASE WHEN status = 0 THEN 1 END) as "did_not_vote",
		COUNT(CASE WHEN status = 1 THEN 1 END) as "failed_on_chain",
		COUNT(CASE WHEN status = 2 THEN 1 END) as "not_found",
		COUNT(CASE WHEN status = 3 THEN 1 END) as "succeeded_on_chain"
	FROM %s idx
	JOIN meta.verifier_info vi ON idx.verifier_id  = vi.id
	WHERE idx.created_at < NOW() + INTERVAL '1 minutes'
	GROUP BY vi.moniker;
	`, partitionTableName)
	err := repo.NewRaw(query).Scan(ctx, &voteList)
	if err != nil {
		return nil, err
	}

	return voteList, nil
}
