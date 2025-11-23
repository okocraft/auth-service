-- name: InsertAccessLog :exec
INSERT INTO users_access_logs (user_id, action_type, login_id, ip, user_agent, created_at)
VALUES (?, ?, ?, ?, ?, ?);
