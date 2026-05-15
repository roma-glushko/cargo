package location

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var ErrUnknown = errors.New("unknown location")

var unlocodePattern = regexp.MustCompile(`^[A-Z]{2}[A-Z2-9]{3}$`)

// UNLocode is a United Nations location code (5 characters: 2 country + 3 location).
type UNLocode string

func NewUNLocode(code string) (UNLocode, error) {
	upper := strings.ToUpper(code)
	if !unlocodePattern.MatchString(upper) {
		return "", fmt.Errorf("invalid UN/LOCODE: %q", code)
	}
	return UNLocode(upper), nil
}

type Location struct {
	Code UNLocode
	Name string
}

type Repository interface {
	Find(ctx context.Context, code UNLocode) (*Location, error)
	FindAll(ctx context.Context) ([]*Location, error)
	Store(ctx context.Context, loc *Location) error
}
