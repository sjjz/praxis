-- name: UserTimezone :one
SELECT timezone FROM user_settings WHERE user_id = $1;

-- name: DailyMoodAverages :many
SELECT
  (timestamp_utc AT TIME ZONE $2)::date AS date_local,
  mood_type,
  AVG(quality)::double precision AS avg_quality
FROM mood_checkins
WHERE user_id = $1
  AND (timestamp_utc AT TIME ZONE $2)::date BETWEEN $3::date AND $4::date
GROUP BY 1, 2
ORDER BY 1 ASC;

-- name: DailyMacroTotals :many
SELECT
  (timestamp_utc AT TIME ZONE $2)::date AS date_local,
  SUM(calories)::double precision AS calories,
  SUM(protein_g)::double precision AS protein_g,
  SUM(fiber_g)::double precision AS fiber_g,
  SUM(added_sugar_g)::double precision AS added_sugar_g,
  SUM(carbs_g)::double precision AS carbs_g
FROM nutrition_entries
WHERE user_id = $1
  AND (timestamp_utc AT TIME ZONE $2)::date BETWEEN $3::date AND $4::date
GROUP BY 1
ORDER BY 1 ASC;
