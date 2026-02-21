package lib

import "time"

type APIError struct {
	Error   string `json:"error"`
	Details any    `json:"details,omitempty"`
}

type MoodType string

const (
	MoodTypeEnergy       MoodType = "energy"
	MoodTypeFogHeaviness MoodType = "fog_heaviness"
	MoodTypeStress       MoodType = "stress"
	MoodTypeMotivation   MoodType = "motivation"
)

type MealTag string

const (
	MealTagBreakfast MealTag = "breakfast"
	MealTagLunch     MealTag = "lunch"
	MealTagDinner    MealTag = "dinner"
	MealTagSnack     MealTag = "snack"
	MealTagOther     MealTag = "other"
)

type MoodCheckin struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	MoodType     MoodType  `json:"mood_type"`
	Quality      int16     `json:"quality"`
	Note         *string   `json:"note,omitempty"`
	TimestampUTC time.Time `json:"timestamp_utc"`
}

type NutritionEntry struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	TimestampUTC time.Time `json:"timestamp_utc"`
	MealTag     *MealTag  `json:"meal_tag,omitempty"`
	Calories    *float64  `json:"calories,omitempty"`
	ProteinG    *float64  `json:"protein_g,omitempty"`
	FiberG      *float64  `json:"fiber_g,omitempty"`
	AddedSugarG *float64  `json:"added_sugar_g,omitempty"`
	CarbsG      *float64  `json:"carbs_g,omitempty"`
}

type CreateMoodCheckinRequest struct {
	MoodType  MoodType `json:"mood_type"`
	Quality   int16    `json:"quality"`
	Note      *string  `json:"note"`
	Timestamp string   `json:"timestamp"`
}

type UpdateMoodCheckinRequest struct {
	MoodType  *MoodType `json:"mood_type"`
	Quality   *int16    `json:"quality"`
	Note      *string   `json:"note"`
	Timestamp *string   `json:"timestamp"`
}

type CreateNutritionEntryRequest struct {
	Timestamp   string   `json:"timestamp"`
	MealTag     *MealTag `json:"meal_tag"`
	Calories    *float64 `json:"calories"`
	ProteinG    *float64 `json:"protein_g"`
	FiberG      *float64 `json:"fiber_g"`
	AddedSugarG *float64 `json:"added_sugar_g"`
	CarbsG      *float64 `json:"carbs_g"`
}

type UpdateNutritionEntryRequest struct {
	Timestamp   *string  `json:"timestamp"`
	MealTag     *MealTag `json:"meal_tag"`
	Calories    *float64 `json:"calories"`
	ProteinG    *float64 `json:"protein_g"`
	FiberG      *float64 `json:"fiber_g"`
	AddedSugarG *float64 `json:"added_sugar_g"`
	CarbsG      *float64 `json:"carbs_g"`
}

type MoodDaily struct {
	DateLocal  string            `json:"date_local"`
	MoodByType map[MoodType]float64 `json:"mood_by_type"`
}

type MacroDaily struct {
	DateLocal   string   `json:"date_local"`
	Calories    *float64 `json:"calories,omitempty"`
	ProteinG    *float64 `json:"protein_g,omitempty"`
	FiberG      *float64 `json:"fiber_g,omitempty"`
	AddedSugarG *float64 `json:"added_sugar_g,omitempty"`
	CarbsG      *float64 `json:"carbs_g,omitempty"`
}

type DailySummary struct {
	DateLocal   string             `json:"date_local"`
	Timezone    string             `json:"timezone"`
	MoodByType  map[MoodType]float64 `json:"mood_by_type"`
	Calories    *float64           `json:"calories,omitempty"`
	ProteinG    *float64           `json:"protein_g,omitempty"`
	FiberG      *float64           `json:"fiber_g,omitempty"`
	AddedSugarG *float64           `json:"added_sugar_g,omitempty"`
	CarbsG      *float64           `json:"carbs_g,omitempty"`
}

type TrendsResponse struct {
	Window string         `json:"window"`
	Points []DailySummary `json:"points"`
}
