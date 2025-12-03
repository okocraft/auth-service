-- name: InsertRefreshToken :exec
INSERT INTO users_refresh_tokens (user_id, jti, login_id, created_at)
VALUES (?, ?, ?, ?);

-- name: InsertAccessToken :exec
INSERT INTO users_access_tokens (refresh_token_id, jti, created_at)
VALUES (?, ?, ?);

-- name: GetUserIDAndRefreshTokenIDByJTI :one
SELECT id, user_id
FROM users_refresh_tokens
WHERE jti = ?;

-- name: DeleteRefreshTokensByLoginID :exec
DELETE
FROM users_refresh_tokens
WHERE login_id = ?;

-- name: DeleteAccessTokensByLoginID :exec
DELETE
FROM users_access_tokens
WHERE users_access_tokens.refresh_token_id IN (SELECT users_refresh_tokens.id
                                               FROM users_refresh_tokens
                                               WHERE users_refresh_tokens.login_id = ?);

-- name: GetUserIDByAccessTokenJTI :one
SELECT user_id
FROM users_refresh_tokens
WHERE id = (SELECT refresh_token_id
            FROM users_access_tokens
            WHERE users_access_tokens.jti = ?);

-- name: DeleteExpiredAccessTokens :execrows
DELETE
FROM users_access_tokens
WHERE created_at < ?;

-- name: DeleteExpiredRefreshTokens :execrows
DELETE
FROM users_refresh_tokens
WHERE created_at < ?;
