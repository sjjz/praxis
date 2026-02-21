package lib

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMoodTypeValidation(t *testing.T) {
	if !IsValidMoodType(MoodTypeEnergy) {
		t.Fatalf("expected mood type to be valid")
	}
	if IsValidMoodType("invalid") {
		t.Fatalf("expected invalid mood type to fail")
	}
}

func TestMealTagValidation(t *testing.T) {
	if !IsValidMealTag(MealTagBreakfast) {
		t.Fatalf("expected meal tag to be valid")
	}
	if IsValidMealTag("not-a-tag") {
		t.Fatalf("expected invalid meal tag to fail")
	}
}

func TestCursorRoundTrip(t *testing.T) {
	id := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)
	encoded := EncodeCursor(now, id)
	decoded, err := DecodeCursor(encoded)
	if err != nil {
		t.Fatalf("decode cursor failed: %v", err)
	}
	if decoded.ID != id {
		t.Fatalf("id mismatch")
	}
	if !decoded.Timestamp.Equal(now) {
		t.Fatalf("timestamp mismatch")
	}
}

func TestAtLeastOneMacro(t *testing.T) {
	req := CreateNutritionEntryRequest{}
	if AtLeastOneMacro(req) {
		t.Fatalf("expected false for empty macros")
	}
	v := 12.0
	req.ProteinG = &v
	if !AtLeastOneMacro(req) {
		t.Fatalf("expected true when one macro exists")
	}
}
