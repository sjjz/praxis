CREATE TYPE mood_type_enum AS ENUM ('energy', 'fog_heaviness', 'stress', 'motivation');
CREATE TYPE meal_tag_enum AS ENUM ('breakfast', 'lunch', 'dinner', 'snack', 'other');

CREATE TABLE users (
    id UUID PRIMARY KEY,
    email TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE user_settings (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    timezone TEXT NOT NULL DEFAULT 'America/Los_Angeles'
);

CREATE TABLE mood_checkins (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    mood_type mood_type_enum NOT NULL,
    quality SMALLINT NOT NULL CHECK (quality BETWEEN 1 AND 5),
    note TEXT,
    timestamp_utc TIMESTAMPTZ NOT NULL
);

CREATE TABLE nutrition_entries (
    id UUID PRIMARY KEY,
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
