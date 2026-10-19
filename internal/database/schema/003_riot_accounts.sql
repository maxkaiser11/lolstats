-- +goose Up
CREATE TABLE riot_accounts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  puuid TEXT NOT NULL,
  game_name TEXT NOT NULL,
  tag_line TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  UNIQUE (user_id, puuid)
);

CREATE INDEX riot_accounts_user_id_idx ON riot_accounts(user_id);

-- +goose Down
DROP TABLE riot_accounts;
