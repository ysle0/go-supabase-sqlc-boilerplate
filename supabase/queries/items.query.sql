-- name: GetItemByID :one
SELECT * FROM items
WHERE id = $1
LIMIT 1;

-- name: ListItems :many
SELECT * FROM items
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountItems :one
SELECT COUNT(*) FROM items;

-- name: SearchItemsByName :many
SELECT * FROM items
WHERE name ILIKE '%' || $1 || '%'
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateItem :one
INSERT INTO items (
    name,
    description,
    price,
    quantity
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: UpdateItem :one
UPDATE items
SET
    name = COALESCE($2, name),
    description = COALESCE($3, description),
    price = COALESCE($4, price),
    quantity = COALESCE($5, quantity)
WHERE id = $1
RETURNING *;

-- name: UpdateItemQuantity :one
UPDATE items
SET quantity = quantity + $2
WHERE id = $1
RETURNING *;

-- name: DeleteItem :exec
DELETE FROM items
WHERE id = $1;

-- name: GetLowStockItems :many
SELECT * FROM items
WHERE quantity < $1
ORDER BY quantity ASC;
