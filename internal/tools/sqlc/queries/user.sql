-- name: GetUserIDBySub :one
SELECT user_id
FROM users_sub
WHERE sub = ?;

-- name: InsertSubForUserID :execrows
INSERT IGNORE INTO users_sub (user_id, sub, created_at)
VALUES (?, ?, ?);

-- name: DeleteUserSubBySub :exec
DELETE
FROM users_sub
WHERE sub = ?;

-- name: GetUserIDByLoginKey :one
SELECT user_id
FROM users_login_key
WHERE login_key = ?;

-- name: InsertLoginKeyForUserID :exec
INSERT INTO users_login_key (user_id, login_key, created_at)
VALUES (?, ?, ?);

-- name: DeleteLoginKey :exec
DELETE
FROM users_login_key
WHERE user_id = ?;
