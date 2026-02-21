CREATE EXTENSION IF NOT EXISTS pgcrypto;

DO $$ BEGIN
    CREATE TYPE mood_type_enum AS ENUM ('energy', 'fog_heaviness', 'stress', 'motivation');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE meal_tag_enum AS ENUM ('breakfast', 'lunch', 'dinner', 'snack', 'other');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS user_settings (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    timezone TEXT NOT NULL DEFAULT 'America/Los_Angeles'
);

CREATE TABLE IF NOT EXISTS mood_checkins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    mood_type mood_type_enum NOT NULL,
    quality SMALLINT NOT NULL CHECK (quality BETWEEN 1 AND 5),
    note TEXT,
    timestamp_utc TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS nutrition_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    timestamp_utc TIMESTAMPTZ NOT NULL,
    meal_tag meal_tag_enum,
    calories NUMERIC CHECK (calories >= 0),
    protein_g NUMERIC CHECK (protein_g >= 0),
    fiber_g NUMERIC CHECK (fiber_g >= 0),
    added_sugar_g NUMERIC CHECK (added_sugar_g >= 0),
    carbs_g NUMERIC CHECK (carbs_g >= 0),
    CHECK (
        calories IS NOT NULL OR protein_g IS NOT NULL OR fiber_g IS NOT NULL OR
        added_sugar_g IS NOT NULL OR carbs_g IS NOT NULL
    )
);

CREATE INDEX IF NOT EXISTS idx_mood_checkins_user_ts
    ON mood_checkins (user_id, timestamp_utc DESC);

CREATE INDEX IF NOT EXISTS idx_mood_checkins_user_type_ts
    ON mood_checkins (user_id, mood_type, timestamp_utc DESC);

CREATE INDEX IF NOT EXISTS idx_nutrition_entries_user_ts
    ON nutrition_entries (user_id, timestamp_utc DESC);

INSERT INTO users (id, email)
VALUES ('00000000-0000-0000-0000-000000000001', NULL)
ON CONFLICT (id) DO NOTHING;

INSERT INTO user_settings (user_id, timezone)
VALUES ('00000000-0000-0000-0000-000000000001', 'America/Los_Angeles')
ON CONFLICT (user_id) DO NOTHING;
