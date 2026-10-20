-- name: InsertMatch :exec
INSERT INTO matches (match_id, queue_id, game_duration, game_creation)
VALUES ($1, $2, $3, $4)
ON CONFLICT (match_id) DO NOTHING;

-- name: InsertParticipant :exec
INSERT INTO match_participants (match_id, puuid, champion_name, kills, deaths, assists, win)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (match_id, puuid) DO NOTHING;

-- name: ExistingMatchIDs :many
SELECT match_id FROM matches
WHERE match_id = ANY(sqlc.arg(match_ids)::text[]);

-- name: RecentGamesByPUUID :many
SELECT m.match_id, m.queue_id, m.game_duration, m.game_creation,
       p.champion_name, p.kills, p.deaths, p.assists, p.win
FROM match_participants p
JOIN matches m ON m.match_id = p.match_id
WHERE p.puuid = sqlc.arg(puuid)
ORDER BY m.game_creation DESC
LIMIT sqlc.arg(max_games);

-- name: ChampionStatsByPUUID :many
SELECT p.champion_name,
       COUNT(*)                         AS games,
       COUNT(*) FILTER (WHERE p.win)    AS wins,
       AVG(p.kills)::float8             AS avg_kills,
       AVG(p.deaths)::float8            AS avg_deaths,
       AVG(p.assists)::float8           AS avg_assists
FROM match_participants p
WHERE p.puuid = sqlc.arg(puuid)
GROUP BY p.champion_name
ORDER BY games DESC, p.champion_name;
