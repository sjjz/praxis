package lib

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

func IsValidMoodType(v MoodType) bool {
	switch v {
	case MoodTypeEnergy, MoodTypeFogHeaviness, MoodTypeStress, MoodTypeMotivation:
		return true
	default:
		return false
	}
}

func IsValidMealTag(v MealTag) bool {
	switch v {
	case MealTagBreakfast, MealTagLunch, MealTagDinner, MealTagSnack, MealTagOther:
		return true
	default:
		return false
	}
}

func ParseTimestamp(v string) (time.Time, error) {
	if strings.TrimSpace(v) == "" {
		return time.Time{}, fmt.Errorf("timestamp is required")
	}
	ts, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return time.Time{}, fmt.Errorf("timestamp must be RFC3339: %w", err)
	}
	return ts.UTC(), nil
}

func ParseLimit(v string, dflt int) (int, error) {
	if v == "" {
		return dflt, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("limit must be an integer")
	}
	if n < 1 || n > 200 {
		return 0, fmt.Errorf("limit must be between 1 and 200")
	}
	return n, nil
}

type Cursor struct {
	Timestamp time.Time
	ID        uuid.UUID
}

func EncodeCursor(ts time.Time, id uuid.UUID) string {
	raw := ts.UTC().Format(time.RFC3339Nano) + "|" + id.String()
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

func DecodeCursor(v string) (*Cursor, error) {
	if strings.TrimSpace(v) == "" {
		return nil, nil
	}
	data, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		return nil, fmt.Errorf("cursor is invalid")
	}
	parts := strings.Split(string(data), "|")
	if len(parts) != 2 {
		return nil, fmt.Errorf("cursor is invalid")
	}
	ts, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return nil, fmt.Errorf("cursor timestamp is invalid")
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return nil, fmt.Errorf("cursor id is invalid")
	}
	return &Cursor{Timestamp: ts.UTC(), ID: id}, nil
}

func AtLeastOneMacro(req CreateNutritionEntryRequest) bool {
	return req.Calories != nil || req.ProteinG != nil || req.FiberG != nil || req.AddedSugarG != nil || req.CarbsG != nil
}

func ValidateNonNegative(name string, v *float64) error {
	if v == nil {
		return nil
	}
	if *v < 0 {
		return fmt.Errorf("%s must be >= 0", name)
	}
	return nil
}
