-- name: AddRiotAccount :one
INSERT INTO riot_accounts (user_id, puuid, game_name, tag_line)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListRiotAccountsByUser :many
SELECT * FROM riot_accounts
WHERE user_id = $1
ORDER BY created_at;

-- name: DeleteRiotAccount :execrows
DELETE FROM riot_accounts
WHERE id = $1 AND user_id = $2;

-- name: UpdateRiotAccountName :exec
UPDATE riot_accounts
SET game_name = $2, tag_line = $3
WHERE puuid = $1;
