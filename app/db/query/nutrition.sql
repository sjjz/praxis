-- name: CreateNutritionEntry :one
INSERT INTO nutrition_entries (
  id, user_id, timestamp_utc, meal_tag, calories, protein_g, fiber_g, added_sugar_g, carbs_g
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetNutritionEntry :one
SELECT * FROM nutrition_entries
WHERE id = $1 AND user_id = $2;

-- name: DeleteNutritionEntry :execrows
DELETE FROM nutrition_entries
WHERE id = $1 AND user_id = $2;

-- name: ListNutritionEntries :many
SELECT * FROM nutrition_entries
WHERE user_id = $1
  AND ($2::timestamptz IS NULL OR timestamp_utc >= $2)
  AND ($3::timestamptz IS NULL OR timestamp_utc <= $3)
  AND ($4::meal_tag_enum IS NULL OR meal_tag = $4)
  AND ($5::timestamptz IS NULL OR (timestamp_utc, id) < ($5, $6))
ORDER BY timestamp_utc DESC, id DESC
LIMIT $7;
