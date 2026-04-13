package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"splash.xyz/dex/pkg/constants"
)

var _ SolBlocksModel = (*customSolBlocksModel)(nil)

type (
	// SolBlocksModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSolBlocksModel.
	SolBlocksModel interface {
		solBlocksModel
		Upsert(ctx context.Context, data *SolBlocks) error
		FindFirstFailBlock(ctx context.Context, chainId int64) (*SolBlocks, error)
		FindProcessingSlots(ctx context.Context, chainId int64, fromSlot uint64, limit int) ([]*SolBlocks, error)
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
const queryFirstFailBlockStmt = `
select %s from %s
where chain_id = ? and status = ?
order by slot asc
limit 1
`

const queryProcessingSlotsStmt = `
select %s from %s
where chain_id = ? and status = ? and slot >= ?
order by slot desc
limit ?
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

func (m *customSolBlocksModel) FindFirstFailBlock(ctx context.Context, chainId int64) (*SolBlocks, error) {
	var resp SolBlocks
	query := fmt.Sprintf(queryFirstFailBlockStmt, solBlocksRows, m.table)
	err := m.conn.QueryRowCtx(ctx, &resp, query, chainId, constants.BlockFailed)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customSolBlocksModel) FindProcessingSlots(ctx context.Context, chainId int64, fromSlot uint64, limit int) ([]*SolBlocks, error) {
	var resp []*SolBlocks
	query := fmt.Sprintf(queryProcessingSlotsStmt, solBlocksRows, m.table)
	err := m.conn.QueryRowsCtx(ctx, &resp, query, chainId, constants.BlockPending, fromSlot, limit)
	switch err {
	case nil:
		return resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
