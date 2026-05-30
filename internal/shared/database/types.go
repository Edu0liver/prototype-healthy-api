package database

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

// JSONStringArray is a []string persisted as JSONB. Avoids Postgres array
// driver quirks while keeping a clean Go type.
type JSONStringArray []string

// Value implements driver.Valuer.
func (a JSONStringArray) Value() (driver.Value, error) {
	if a == nil {
		return "[]", nil
	}
	return json.Marshal(a)
}

// Scan implements sql.Scanner.
func (a *JSONStringArray) Scan(src any) error {
	if src == nil {
		*a = nil
		return nil
	}
	var data []byte
	switch v := src.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return errors.New("database: JSONStringArray: unsupported scan type")
	}
	if len(data) == 0 {
		*a = nil
		return nil
	}
	return json.Unmarshal(data, a)
}

// GormDataType tells GORM the column type.
func (JSONStringArray) GormDataType() string { return "jsonb" }

// JSONMap is a map[string]any persisted as JSONB.
type JSONMap map[string]any

// Value implements driver.Valuer.
func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return "{}", nil
	}
	return json.Marshal(m)
}

// Scan implements sql.Scanner.
func (m *JSONMap) Scan(src any) error {
	if src == nil {
		*m = nil
		return nil
	}
	var data []byte
	switch v := src.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return errors.New("database: JSONMap: unsupported scan type")
	}
	if len(data) == 0 {
		*m = nil
		return nil
	}
	return json.Unmarshal(data, m)
}

// GormDataType tells GORM the column type.
func (JSONMap) GormDataType() string { return "jsonb" }

// Vector is a pgvector embedding ([]float32) with text encoding `[v1,v2,...]`.
type Vector []float32

// Value implements driver.Valuer (pgvector accepts the text representation).
func (v Vector) Value() (driver.Value, error) {
	if v == nil {
		return nil, nil
	}
	var b strings.Builder
	b.WriteByte('[')
	for i, f := range v {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.FormatFloat(float64(f), 'f', -1, 32))
	}
	b.WriteByte(']')
	return b.String(), nil
}

// Scan implements sql.Scanner for the pgvector text representation.
func (v *Vector) Scan(src any) error {
	if src == nil {
		*v = nil
		return nil
	}
	var s string
	switch t := src.(type) {
	case []byte:
		s = string(t)
	case string:
		s = t
	default:
		return errors.New("database: Vector: unsupported scan type")
	}
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	if s == "" {
		*v = Vector{}
		return nil
	}
	parts := strings.Split(s, ",")
	out := make(Vector, len(parts))
	for i, p := range parts {
		f, err := strconv.ParseFloat(strings.TrimSpace(p), 32)
		if err != nil {
			return err
		}
		out[i] = float32(f)
	}
	*v = out
	return nil
}

// GormDataType tells GORM the column type.
func (Vector) GormDataType() string { return "vector(1536)" }
