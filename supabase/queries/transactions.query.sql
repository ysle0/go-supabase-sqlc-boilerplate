-- name: GetTransactionByID :one
SELECT * FROM transactions
WHERE id = $1
LIMIT 1;

-- name: ListTransactions :many
SELECT * FROM transactions
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListTransactionsByUserID :many
SELECT * FROM transactions
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListTransactionsByItemID :many
SELECT * FROM transactions
WHERE item_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountTransactions :one
SELECT COUNT(*) FROM transactions;

-- name: CountTransactionsByUserID :one
SELECT COUNT(*) FROM transactions
WHERE user_id = $1;

-- name: CreateTransaction :one
INSERT INTO transactions (
    user_id,
    item_id,
    transaction_type,
    quantity,
    amount,
    notes
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetTransactionsByDateRange :many
SELECT * FROM transactions
WHERE created_at >= $1 AND created_at <= $2
ORDER BY created_at DESC;

-- name: GetUserTransactionSummary :one
SELECT
    user_id,
    COUNT(*) as total_transactions,
    SUM(amount) as total_amount,
    SUM(CASE WHEN transaction_type = 'purchase' THEN 1 ELSE 0 END) as purchase_count,
    SUM(CASE WHEN transaction_type = 'refund' THEN 1 ELSE 0 END) as refund_count
FROM transactions
WHERE user_id = $1
GROUP BY user_id;
