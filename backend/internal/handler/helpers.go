package handler

import (
	"database/sql"
	"fmt"
	"strings"

	"printmart/backend/internal/repository"
)

// scanRows mengubah baris SQL menjadi map untuk JSON, mempertahankan
// tipe numerik SQLite (INTEGER -> int64) agar paralel dengan output Node.
func scanRows(rows *sql.Rows) ([]map[string]any, error) {
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	vals := make([]any, len(cols))
	ptrs := make([]any, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}
	out := []map[string]any{}
	for rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		row := map[string]any{}
		for i, c := range cols {
			row[c] = vals[i]
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

type queryer interface {
	Query(query string, args ...any) (*sql.Rows, error)
}

func allRowsQ(q queryer, query string, args ...any) ([]map[string]any, error) {
	rs, err := q.Query(query, args...)
	if err != nil {
		return nil, err
	}
	return scanRows(rs)
}

func allRows(db *sql.DB, query string, args ...any) ([]map[string]any, error) {
	return allRowsQ(db, query, args...)
}

func allRowsTx(tx *sql.Tx, query string, args ...any) ([]map[string]any, error) {
	return allRowsQ(tx, query, args...)
}

func getRow(db *sql.DB, query string, args ...any) (map[string]any, error) {
	return row1(db, query, args...)
}

func getRowTx(tx *sql.Tx, query string, args ...any) (map[string]any, error) {
	return row1(tx, query, args...)
}

func row1(q queryer, query string, args ...any) (map[string]any, error) {
	rows, err := allRowsQ(q, query, args...)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return rows[0], nil
}

func toInt(v any) int64 {
	switch t := v.(type) {
	case int64:
		return t
	case int:
		return int64(t)
	case float64:
		return int64(t)
	}
	return 0
}

func toStr(v any) string {
	if v == nil {
		return ""
	}
	if b, ok := v.([]byte); ok {
		return string(b)
	}
	return fmt.Sprintf("%v", v)
}

func toBoolInt(v any) int64 {
	if b, ok := v.(bool); ok {
		if b {
			return 1
		}
		return 0
	}
	return toInt(v)
}

func num(v any) int64 {
	switch t := v.(type) {
	case float64:
		return int64(t)
	case float32:
		return int64(t)
	case int64:
		return t
	case int:
		return int64(t)
	}
	return 0
}

// truncateOrderId meniru id.slice(0, 8).toUpperCase()
func truncateOrderId(id string) string {
	r := []rune(id)
	if len(r) > 8 {
		r = r[:8]
	}
	return strings.ToUpper(string(r))
}

func newID() string { return repository.UUID() }