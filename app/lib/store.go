package lib

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}
	cfg.MaxConns = 10
	return pgxpool.NewWithConfig(ctx, cfg)
}

func (s *Store) EnsureUser(ctx context.Context, userID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO users (id, email) VALUES ($1, NULL)
		ON CONFLICT (id) DO NOTHING
	`, userID)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(ctx, `
		INSERT INTO user_settings (user_id, timezone) VALUES ($1, 'America/Los_Angeles')
		ON CONFLICT (user_id) DO NOTHING
	`, userID)
	return err
}

func (s *Store) UserTimezone(ctx context.Context, userID uuid.UUID) (string, error) {
	var tz string
	err := s.db.QueryRow(ctx, `
		SELECT timezone FROM user_settings WHERE user_id = $1
	`, userID).Scan(&tz)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "America/Los_Angeles", nil
		}
		return "", err
	}
	return tz, nil
}

func (s *Store) CreateMoodCheckin(ctx context.Context, userID uuid.UUID, req CreateMoodCheckinRequest) (*MoodCheckin, error) {
	id := uuid.New()
	row := s.db.QueryRow(ctx, `
		INSERT INTO mood_checkins (id, user_id, mood_type, quality, note, timestamp_utc)
		VALUES ($1, $2, $3::mood_type_enum, $4, $5, $6)
		RETURNING id, user_id, mood_type::text, quality, note, timestamp_utc
	`, id, userID, string(req.MoodType), req.Quality, req.Note, req.Timestamp)
	return scanMood(row)
}

type ListMoodFilters struct {
	From     *time.Time
	To       *time.Time
	MoodType *MoodType
	Cursor   *Cursor
	Limit    int
}

func (s *Store) ListMoodCheckins(ctx context.Context, userID uuid.UUID, f ListMoodFilters) ([]MoodCheckin, string, error) {
	var moodType any
	if f.MoodType != nil {
		moodType = string(*f.MoodType)
	}
	var cursorTS any
	cursorID := uuid.Nil
	if f.Cursor != nil {
		cursorTS = f.Cursor.Timestamp
		cursorID = f.Cursor.ID
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, user_id, mood_type::text, quality, note, timestamp_utc
		FROM mood_checkins
		WHERE user_id = $1
		  AND ($2::timestamptz IS NULL OR timestamp_utc >= $2)
		  AND ($3::timestamptz IS NULL OR timestamp_utc <= $3)
		  AND ($4::mood_type_enum IS NULL OR mood_type = $4)
		  AND ($5::timestamptz IS NULL OR (timestamp_utc, id) < ($5, $6))
		ORDER BY timestamp_utc DESC, id DESC
		LIMIT $7
	`, userID, f.From, f.To, moodType, cursorTS, cursorID, f.Limit)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	out := make([]MoodCheckin, 0, f.Limit)
	for rows.Next() {
		item, scanErr := scanMood(rows)
		if scanErr != nil {
			return nil, "", scanErr
		}
		out = append(out, *item)
	}
	if rows.Err() != nil {
		return nil, "", rows.Err()
	}

	var next string
	if len(out) == f.Limit {
		last := out[len(out)-1]
		lastID, _ := uuid.Parse(last.ID)
		next = EncodeCursor(last.TimestampUTC, lastID)
	}

	return out, next, nil
}

func (s *Store) GetMoodCheckin(ctx context.Context, userID, id uuid.UUID) (*MoodCheckin, error) {
	row := s.db.QueryRow(ctx, `
		SELECT id, user_id, mood_type::text, quality, note, timestamp_utc
		FROM mood_checkins
		WHERE user_id = $1 AND id = $2
	`, userID, id)
	return scanMood(row)
}

func (s *Store) UpdateMoodCheckin(ctx context.Context, userID, id uuid.UUID, req UpdateMoodCheckinRequest) (*MoodCheckin, error) {
	current, err := s.GetMoodCheckin(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	nextMoodType := current.MoodType
	nextQuality := current.Quality
	nextNote := current.Note
	nextTimestamp := current.TimestampUTC

	if req.MoodType != nil {
		nextMoodType = *req.MoodType
	}
	if req.Quality != nil {
		nextQuality = *req.Quality
	}
	if req.Note != nil {
		nextNote = req.Note
	}
	if req.Timestamp != nil {
		ts, parseErr := ParseTimestamp(*req.Timestamp)
		if parseErr != nil {
			return nil, parseErr
		}
		nextTimestamp = ts
	}

	row := s.db.QueryRow(ctx, `
		UPDATE mood_checkins
		SET mood_type = $3::mood_type_enum, quality = $4, note = $5, timestamp_utc = $6
		WHERE user_id = $1 AND id = $2
		RETURNING id, user_id, mood_type::text, quality, note, timestamp_utc
	`, userID, id, string(nextMoodType), nextQuality, nextNote, nextTimestamp)
	return scanMood(row)
}

func (s *Store) DeleteMoodCheckin(ctx context.Context, userID, id uuid.UUID) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM mood_checkins WHERE user_id = $1 AND id = $2`, userID, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func scanMood(row pgx.Row) (*MoodCheckin, error) {
	var out MoodCheckin
	var moodType string
	var note sql.NullString
	err := row.Scan(&out.ID, &out.UserID, &moodType, &out.Quality, &note, &out.TimestampUTC)
	if err != nil {
		return nil, err
	}
	out.MoodType = MoodType(moodType)
	if note.Valid {
		out.Note = &note.String
	}
	return &out, nil
}

func (s *Store) CreateNutritionEntry(ctx context.Context, userID uuid.UUID, req CreateNutritionEntryRequest) (*NutritionEntry, error) {
	id := uuid.New()
	row := s.db.QueryRow(ctx, `
		INSERT INTO nutrition_entries (
			id, user_id, timestamp_utc, meal_tag, calories, protein_g, fiber_g, added_sugar_g, carbs_g
		)
		VALUES ($1, $2, $3, $4::meal_tag_enum, $5, $6, $7, $8, $9)
		RETURNING id, user_id, timestamp_utc, meal_tag::text, calories::double precision,
		          protein_g::double precision, fiber_g::double precision,
		          added_sugar_g::double precision, carbs_g::double precision
	`, id, userID, req.Timestamp, mealTagParam(req.MealTag), req.Calories, req.ProteinG, req.FiberG, req.AddedSugarG, req.CarbsG)
	return scanNutrition(row)
}

type ListNutritionFilters struct {
	From    *time.Time
	To      *time.Time
	MealTag *MealTag
	Cursor  *Cursor
	Limit   int
}

func (s *Store) ListNutritionEntries(ctx context.Context, userID uuid.UUID, f ListNutritionFilters) ([]NutritionEntry, string, error) {
	var mealTag any
	if f.MealTag != nil {
		mealTag = string(*f.MealTag)
	}
	var cursorTS any
	cursorID := uuid.Nil
	if f.Cursor != nil {
		cursorTS = f.Cursor.Timestamp
		cursorID = f.Cursor.ID
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, user_id, timestamp_utc, meal_tag::text, calories::double precision,
		       protein_g::double precision, fiber_g::double precision,
		       added_sugar_g::double precision, carbs_g::double precision
		FROM nutrition_entries
		WHERE user_id = $1
		  AND ($2::timestamptz IS NULL OR timestamp_utc >= $2)
		  AND ($3::timestamptz IS NULL OR timestamp_utc <= $3)
		  AND ($4::meal_tag_enum IS NULL OR meal_tag = $4)
		  AND ($5::timestamptz IS NULL OR (timestamp_utc, id) < ($5, $6))
		ORDER BY timestamp_utc DESC, id DESC
		LIMIT $7
	`, userID, f.From, f.To, mealTag, cursorTS, cursorID, f.Limit)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	out := make([]NutritionEntry, 0, f.Limit)
	for rows.Next() {
		item, scanErr := scanNutrition(rows)
		if scanErr != nil {
			return nil, "", scanErr
		}
		out = append(out, *item)
	}
	if rows.Err() != nil {
		return nil, "", rows.Err()
	}
	var next string
	if len(out) == f.Limit {
		last := out[len(out)-1]
		lastID, _ := uuid.Parse(last.ID)
		next = EncodeCursor(last.TimestampUTC, lastID)
	}
	return out, next, nil
}

func (s *Store) GetNutritionEntry(ctx context.Context, userID, id uuid.UUID) (*NutritionEntry, error) {
	row := s.db.QueryRow(ctx, `
		SELECT id, user_id, timestamp_utc, meal_tag::text, calories::double precision,
		       protein_g::double precision, fiber_g::double precision,
		       added_sugar_g::double precision, carbs_g::double precision
		FROM nutrition_entries
		WHERE user_id = $1 AND id = $2
	`, userID, id)
	return scanNutrition(row)
}

func (s *Store) UpdateNutritionEntry(ctx context.Context, userID, id uuid.UUID, req UpdateNutritionEntryRequest) (*NutritionEntry, error) {
	current, err := s.GetNutritionEntry(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	ts := current.TimestampUTC
	mealTag := current.MealTag
	calories := current.Calories
	protein := current.ProteinG
	fiber := current.FiberG
	addedSugar := current.AddedSugarG
	carbs := current.CarbsG

	if req.Timestamp != nil {
		parsed, parseErr := ParseTimestamp(*req.Timestamp)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid timestamp: %w", parseErr)
		}
		ts = parsed
	}
	if req.MealTag != nil {
		mealTag = req.MealTag
	}
	if req.Calories != nil {
		calories = req.Calories
	}
	if req.ProteinG != nil {
		protein = req.ProteinG
	}
	if req.FiberG != nil {
		fiber = req.FiberG
	}
	if req.AddedSugarG != nil {
		addedSugar = req.AddedSugarG
	}
	if req.CarbsG != nil {
		carbs = req.CarbsG
	}

	row := s.db.QueryRow(ctx, `
		UPDATE nutrition_entries
		SET timestamp_utc = $3,
		    meal_tag = $4::meal_tag_enum,
		    calories = $5,
		    protein_g = $6,
		    fiber_g = $7,
		    added_sugar_g = $8,
		    carbs_g = $9
		WHERE user_id = $1 AND id = $2
		RETURNING id, user_id, timestamp_utc, meal_tag::text, calories::double precision,
		          protein_g::double precision, fiber_g::double precision,
		          added_sugar_g::double precision, carbs_g::double precision
	`, userID, id, ts, mealTagParam(mealTag), calories, protein, fiber, addedSugar, carbs)
	return scanNutrition(row)
}

func (s *Store) DeleteNutritionEntry(ctx context.Context, userID, id uuid.UUID) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM nutrition_entries WHERE user_id = $1 AND id = $2`, userID, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func scanNutrition(row pgx.Row) (*NutritionEntry, error) {
	var out NutritionEntry
	var mealTag sql.NullString
	var calories, protein, fiber, addedSugar, carbs sql.NullFloat64
	err := row.Scan(
		&out.ID, &out.UserID, &out.TimestampUTC, &mealTag,
		&calories, &protein, &fiber, &addedSugar, &carbs,
	)
	if err != nil {
		return nil, err
	}
	if mealTag.Valid {
		tag := MealTag(mealTag.String)
		out.MealTag = &tag
	}
	if calories.Valid {
		out.Calories = &calories.Float64
	}
	if protein.Valid {
		out.ProteinG = &protein.Float64
	}
	if fiber.Valid {
		out.FiberG = &fiber.Float64
	}
	if addedSugar.Valid {
		out.AddedSugarG = &addedSugar.Float64
	}
	if carbs.Valid {
		out.CarbsG = &carbs.Float64
	}
	return &out, nil
}

func (s *Store) DailySummaries(ctx context.Context, userID uuid.UUID, from, to time.Time, timezone string) ([]DailySummary, error) {
	rows, err := s.db.Query(ctx, `
		WITH days AS (
		  SELECT generate_series($2::date, $3::date, interval '1 day')::date AS d
		),
		mood AS (
		  SELECT (timestamp_utc AT TIME ZONE $4)::date AS d,
		         mood_type::text AS mood_type,
		         AVG(quality)::double precision AS avg_quality
		  FROM mood_checkins
		  WHERE user_id = $1
		    AND (timestamp_utc AT TIME ZONE $4)::date BETWEEN $2::date AND $3::date
		  GROUP BY 1,2
		),
		macro AS (
		  SELECT (timestamp_utc AT TIME ZONE $4)::date AS d,
		         SUM(calories)::double precision AS calories,
		         SUM(protein_g)::double precision AS protein_g,
		         SUM(fiber_g)::double precision AS fiber_g,
		         SUM(added_sugar_g)::double precision AS added_sugar_g,
		         SUM(carbs_g)::double precision AS carbs_g
		  FROM nutrition_entries
		  WHERE user_id = $1
		    AND (timestamp_utc AT TIME ZONE $4)::date BETWEEN $2::date AND $3::date
		  GROUP BY 1
		)
		SELECT days.d::text,
		       COALESCE(mood.mood_type, ''),
		       mood.avg_quality,
		       macro.calories, macro.protein_g, macro.fiber_g, macro.added_sugar_g, macro.carbs_g
		FROM days
		LEFT JOIN mood ON mood.d = days.d
		LEFT JOIN macro ON macro.d = days.d
		ORDER BY days.d ASC
	`, userID, from.Format("2006-01-02"), to.Format("2006-01-02"), timezone)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summaryMap := map[string]*DailySummary{}
	order := make([]string, 0)
	for rows.Next() {
		var dateLocal, moodType string
		var moodAvg, calories, protein, fiber, addedSugar, carbs sql.NullFloat64
		if err := rows.Scan(&dateLocal, &moodType, &moodAvg, &calories, &protein, &fiber, &addedSugar, &carbs); err != nil {
			return nil, err
		}
		item, ok := summaryMap[dateLocal]
		if !ok {
			item = &DailySummary{
				DateLocal:  dateLocal,
				Timezone:   timezone,
				MoodByType: map[MoodType]float64{},
			}
			summaryMap[dateLocal] = item
			order = append(order, dateLocal)
			if calories.Valid {
				item.Calories = &calories.Float64
			}
			if protein.Valid {
				item.ProteinG = &protein.Float64
			}
			if fiber.Valid {
				item.FiberG = &fiber.Float64
			}
			if addedSugar.Valid {
				item.AddedSugarG = &addedSugar.Float64
			}
			if carbs.Valid {
				item.CarbsG = &carbs.Float64
			}
		}
		if moodType != "" && moodAvg.Valid {
			item.MoodByType[MoodType(moodType)] = moodAvg.Float64
		}
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	out := make([]DailySummary, 0, len(order))
	for _, date := range order {
		out = append(out, *summaryMap[date])
	}
	return out, nil
}

func mealTagParam(tag *MealTag) any {
	if tag == nil {
		return nil
	}
	return string(*tag)
}
