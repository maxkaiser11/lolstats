-- +goose Up
CREATE TABLE matches (
  match_id TEXT PRIMARY KEY,
  queue_id INT NOT NULL,
  game_duration INT NOT NULL,
  game_creation TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT now()
);


CREATE TABLE match_participants (
  match_id      TEXT NOT NULL REFERENCES matches(match_id) ON DELETE CASCADE,
  puuid         TEXT NOT NULL,
  champion_name TEXT NOT NULL,
  kills         INT NOT NULL,
  deaths        INT NOT NULL,
  assists       INT NOT NULL,
  win           BOOLEAN NOT NULL,
  PRIMARY KEY (match_id, puuid)
);

CREATE INDEX match_participants_puuid_idx ON match_participants(puuid);

-- +goose Down
DROP TABLE match_participants;
DROP TABLE matches;
