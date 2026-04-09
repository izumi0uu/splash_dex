package model

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ SolBlocksModel = (*customSolBlocksModel)(nil)

type (
	// SolBlocksModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSolBlocksModel.
	SolBlocksModel interface {
		solBlocksModel
		Upsert(ctx context.Context, data *SolBlocks) error
		withSession(session sqlx.Session) SolBlocksModel
	}

	customSolBlocksModel struct {
		*defaultSolBlocksModel
	}
)

// NewSolBlocksModel returns a model for the database table.
func NewSolBlocksModel(conn sqlx.SqlConn) SolBlocksModel {
	return &customSolBlocksModel{
		defaultSolBlocksModel: newSolBlocksModel(conn),
	}
}

func (m *customSolBlocksModel) withSession(session sqlx.Session) SolBlocksModel {
	return NewSolBlocksModel(sqlx.NewSqlConnFromSession(session))
}

const upsertSolBlocksStmt = `
INSERT INTO sol_blocks (
	chain_id,
	slot,
	block_height,
	parent_slot,
	block_hash,
	previous_block_hash,
	block_time,
	tx_count,
	status,
	err_msg
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
	block_height = VALUES(block_height),
	parent_slot = VALUES(parent_slot),
	block_hash = VALUES(block_hash),
	previous_block_hash = VALUES(previous_block_hash),
	block_time = VALUES(block_time),
	tx_count = VALUES(tx_count),
	status = VALUES(status),
	err_msg = VALUES(err_msg),
	updated_at = CURRENT_TIMESTAMP
`

func (m *customSolBlocksModel) Upsert(ctx context.Context, data *SolBlocks) error {
	_, err := m.conn.ExecCtx(ctx, upsertSolBlocksStmt,
		data.ChainId,
		data.Slot,
		data.BlockHeight,
		data.ParentSlot,
		data.BlockHash,
		data.PreviousBlockHash,
		data.BlockTime,
		data.TxCount,
		data.Status,
		data.ErrMsg,
	)
	return err
}
