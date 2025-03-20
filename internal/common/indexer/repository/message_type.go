package repository

import (
	"context"

	"github.com/cosmostation/cvms/internal/common/indexer/model"
	"github.com/pkg/errors"
)

const MESSAGE_TYPE_TABLE_NAME = "message_type"

func (repo *MetaRepository) GetMessageTypeListByChainInfoID(chainInfoID int64) ([]model.MessageType, error) {
	ctx := context.Background()
	defer ctx.Done()

	modelList := make([]model.MessageType, 0)
	err := repo.
		NewSelect().
		Model(&modelList).
		Where("chain_info_id = ?", chainInfoID).
		Scan(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query validator_info list by chain_info_id")
	}

	return modelList, nil
}

func (repo *MetaRepository) InsertMessageTypeList(messageTypeList []model.MessageType) error {
	ctx := context.Background()
	defer ctx.Done()

	_, err := repo.NewInsert().
		Model(&messageTypeList).
		ExcludeColumn("id").
		Exec(ctx)
	if err != nil {
		return errors.Wrapf(err, "failed to insert new message types: %v", messageTypeList)
	}

	return nil
}
