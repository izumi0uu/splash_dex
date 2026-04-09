package block

import (
	"database/sql"
	"time"
)

func NullableInt64(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: true}
}

func NullableString(v string) sql.NullString {
	if v == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: v, Valid: true}
}

func NullableTime(v time.Time) sql.NullTime {
	return sql.NullTime{Time: v, Valid: true}
}

func TrimmedNullableString(msg string) sql.NullString {
	const maxErrMsgLen = 512

	if len(msg) <= maxErrMsgLen {
		return NullableString(msg)
	}

	return TrimmedNullableString(msg[:maxErrMsgLen])
}
