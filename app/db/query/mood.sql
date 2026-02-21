-- name: CreateMoodCheckin :one
INSERT INTO mood_checkins (id, user_id, mood_type, quality, note, timestamp_utc)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetMoodCheckin :one
SELECT * FROM mood_checkins
WHERE id = $1 AND user_id = $2;

-- name: DeleteMoodCheckin :execrows
DELETE FROM mood_checkins
WHERE id = $1 AND user_id = $2;

-- name: ListMoodCheckins :many
SELECT * FROM mood_checkins
WHERE user_id = $1
  AND ($2::timestamptz IS NULL OR timestamp_utc >= $2)
  AND ($3::timestamptz IS NULL OR timestamp_utc <= $3)
  AND ($4::mood_type_enum IS NULL OR mood_type = $4)
  AND ($5::timestamptz IS NULL OR (timestamp_utc, id) < ($5, $6))
ORDER BY timestamp_utc DESC, id DESC
LIMIT $7;
