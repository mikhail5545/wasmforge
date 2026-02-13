package pagination

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PageTokenPayload struct {
	CursorValue any    `json:"v"`
	LastID      string `json:"id"`
}

// EncodePageToken encodes the pagination token with the last seen cursor value and ID.
func EncodePageToken(val any, id uuid.UUID) string {
	// Ensure time is in UTC
	if t, ok := val.(time.Time); ok {
		val = t.UTC()
	}
	p := PageTokenPayload{
		CursorValue: val,
		LastID:      id.String(),
	}
	b, _ := json.Marshal(p)
	return base64.RawURLEncoding.EncodeToString(b)
}

func DecodePageToken(token string) (any, uuid.UUID, error) {
	if token == "" {
		return nil, uuid.Nil, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return nil, uuid.Nil, err
	}
	var p PageTokenPayload
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, uuid.Nil, err
	}
	id, err := uuid.Parse(p.LastID)
	if err != nil {
		return nil, uuid.Nil, err
	}
	return p.CursorValue, id, nil
}

func normalizeOrderDirection(dir string) string {
	if dir == "ASC" || dir == "asc" {
		return "ASC"
	}
	return "DESC"
}

type ApplyCursorParams struct {
	PageSize   int
	PageToken  string
	OrderField string
	OrderDir   string
}

var safeColumnName = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func ApplyCursor(db *gorm.DB, params ApplyCursorParams) (*gorm.DB, error) {
	if params.PageSize < 0 {
		return nil, errors.New("page_size must be non-negative")
	}
	if !safeColumnName.MatchString(params.OrderField) {
		return nil, fmt.Errorf("invalid order field: %s", params.OrderField)
	}
	params.OrderDir = normalizeOrderDirection(params.OrderDir)

	cursorVal, lastID, err := DecodePageToken(params.PageToken)
	if err != nil {
		return nil, fmt.Errorf("invalid page token: %w", err)
	}

	orderExpr := fmt.Sprintf("%s %s, id %s", params.OrderField, params.OrderDir, params.OrderDir)
	db = db.Order(orderExpr).Limit(params.PageSize + 1) // Fetch one extra to check for next page

	if cursorVal != nil && lastID != uuid.Nil {
		op := ">"
		if params.OrderDir == "DESC" {
			op = "<"
		}
		db = db.Where(fmt.Sprintf("(%s, id) %s (?, ?)", params.OrderField, op), cursorVal, lastID)
	}
	return db, nil
}
