package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ SolTransactionsModel = (*customSolTransactionsModel)(nil)

type (
	// SolTransactionsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSolTransactionsModel.
	SolTransactionsModel interface {
		solTransactionsModel
		withSession(session sqlx.Session) SolTransactionsModel
	}

	customSolTransactionsModel struct {
		*defaultSolTransactionsModel
	}
)

// NewSolTransactionsModel returns a model for the database table.
func NewSolTransactionsModel(conn sqlx.SqlConn) SolTransactionsModel {
	return &customSolTransactionsModel{
		defaultSolTransactionsModel: newSolTransactionsModel(conn),
	}
}

func (m *customSolTransactionsModel) withSession(session sqlx.Session) SolTransactionsModel {
	return NewSolTransactionsModel(sqlx.NewSqlConnFromSession(session))
}
