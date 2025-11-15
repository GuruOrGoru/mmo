-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1
LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (
    username, password_hash
) VALUES (
    $1, $2
)
RETURNING *;

-- name: CreatePlayer :one
INSERT INTO players (
    user_id, name, color
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetPlayerById :one
SELECT * FROM players
WHERE user_id = $1
LIMIT 1;

-- name: UpdatePlayerHighScore :exec
UPDATE players
SET best_score = $1
WHERE id = $2;

-- name: GetTopScores :many
SELECT name, best_score
FROM players
ORDER BY best_score DESC
LIMIT $1
OFFSET $2;

-- name: GetPlayerByName :one
SELECT * FROM players
WHERE name LIKE $1
LIMIT 1;

-- name: GetPlayerRank :one
SELECT COUNT(*) + 1 AS "rank" FROM players
WHERE best_score > (
	SELECT best_score FROM players p2
	WHERE p2.id = $1
);
