package model

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ SolSwapsModel = (*customSolSwapsModel)(nil)

type (
	// SolSwapsModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSolSwapsModel.
	SolSwapsModel interface {
		solSwapsModel
		withSession(session sqlx.Session) SolSwapsModel
	}

	customSolSwapsModel struct {
		*defaultSolSwapsModel
	}
)

// NewSolSwapsModel returns a model for the database table.
func NewSolSwapsModel(conn sqlx.SqlConn) SolSwapsModel {
	return &customSolSwapsModel{
		defaultSolSwapsModel: newSolSwapsModel(conn),
	}
}

func (m *customSolSwapsModel) withSession(session sqlx.Session) SolSwapsModel {
	return NewSolSwapsModel(sqlx.NewSqlConnFromSession(session))
}
