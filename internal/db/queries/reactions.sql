-- name: AddReaction :exec
INSERT INTO reactions (user_id, reaction)
VALUES (?, ?);

-- name: IncrementReactionCount :exec
INSERT INTO reaction_counts (reaction, count)
VALUES (?, 1)
ON CONFLICT(reaction) DO UPDATE SET count = count + 1;

-- name: GetReactionCounts :many
SELECT reaction, count FROM reaction_counts ORDER BY count DESC;
