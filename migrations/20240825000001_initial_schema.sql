-- +goose Up
-- Create users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE CHECK (length(name) >= 3 AND length(name) <= 20),
    email TEXT NOT NULL CHECK (email ~ '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
    password TEXT NOT NULL CHECK (length(password) >= 6),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create unique index on lowercase email to prevent duplicates
CREATE UNIQUE INDEX idx_users_email_lower ON users (LOWER(email));

-- Create abilities table
CREATE TABLE abilities (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    base_damage INT NOT NULL,
    ap_cost INT NOT NULL,
    range INT NOT NULL,
    aoe_radius INT NULL,
    per_turn_limit INT NULL,
    per_target_per_turn_limit INT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create matches table
CREATE TABLE matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status TEXT NOT NULL CHECK (status IN ('pending', 'active', 'ended')),
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMPTZ NULL,
    winner_user_id UUID NULL,
    end_reason TEXT NULL
);

-- Create match_participants table
CREATE TABLE match_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    match_id UUID NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    is_bot BOOLEAN NOT NULL DEFAULT FALSE,
    starting_hp INT NOT NULL,
    starting_ap INT NOT NULL,
    start_x INT NOT NULL,
    start_y INT NOT NULL
);

-- Create character_snapshots table
CREATE TABLE character_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    match_id UUID NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    turn_no INT NOT NULL,
    hp INT NOT NULL,
    ap INT NOT NULL,
    x INT NOT NULL,
    y INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create ability_snapshots table
CREATE TABLE ability_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    match_id UUID NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    turn_no INT NOT NULL,
    ability_id TEXT NOT NULL REFERENCES abilities(id),
    uses_this_turn INT NOT NULL DEFAULT 0,
    per_target_uses JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create match_actions table
CREATE TABLE match_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    match_id UUID NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    turn_no INT NOT NULL,
    action_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_matches_status ON matches(status);
CREATE INDEX idx_match_participants_match_id ON match_participants(match_id);
CREATE INDEX idx_character_snapshots_match_user_turn ON character_snapshots(match_id, user_id, turn_no);
CREATE INDEX idx_ability_snapshots_match_user_turn ON ability_snapshots(match_id, user_id, turn_no);

-- +goose Down
DROP TABLE IF EXISTS match_actions CASCADE;
DROP TABLE IF EXISTS ability_snapshots CASCADE;
DROP TABLE IF EXISTS character_snapshots CASCADE;
DROP TABLE IF EXISTS match_participants CASCADE;
DROP TABLE IF EXISTS matches CASCADE;
DROP TABLE IF EXISTS abilities CASCADE;
DROP TABLE IF EXISTS users CASCADE;
