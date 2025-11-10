import postgres from "https://deno.land/x/postgresjs@v3.4.7/mod.js";

type Sql = postgres.Sql;
export const getItemByIDQuery = `-- name: GetItemByID :one
SELECT id, name, description, price, quantity, created_at, updated_at FROM items
WHERE id = $1
LIMIT 1`;

export interface GetItemByIDArgs {
    id: string;
}

export interface GetItemByIDRow {
    id: string;
    name: string;
    description: string | null;
    price: string;
    quantity: number;
    createdAt: Date;
    updatedAt: Date;
}

export async function getItemByID(sql: Sql, args: GetItemByIDArgs): Promise<GetItemByIDRow | null> {
    const rows = await sql.unsafe(getItemByIDQuery, [args.id]).values();
    if (rows.length !== 1) {
        return null;
    }
    const row = rows[0];
    return {
        id: row[0],
        name: row[1],
        description: row[2],
        price: row[3],
        quantity: row[4],
        createdAt: row[5],
        updatedAt: row[6]
    };
}

export const listItemsQuery = `-- name: ListItems :many
SELECT id, name, description, price, quantity, created_at, updated_at FROM items
ORDER BY created_at DESC
LIMIT $1 OFFSET $2`;

export interface ListItemsArgs {
    limit: string;
    offset: string;
}

export interface ListItemsRow {
    id: string;
    name: string;
    description: string | null;
    price: string;
    quantity: number;
    createdAt: Date;
    updatedAt: Date;
}

export async function listItems(sql: Sql, args: ListItemsArgs): Promise<ListItemsRow[]> {
    return (await sql.unsafe(listItemsQuery, [args.limit, args.offset]).values()).map(row => ({
        id: row[0],
        name: row[1],
        description: row[2],
        price: row[3],
        quantity: row[4],
        createdAt: row[5],
        updatedAt: row[6]
    }));
}

export const countItemsQuery = `-- name: CountItems :one
SELECT COUNT(*) FROM items`;

export interface CountItemsRow {
    count: string;
}

export async function countItems(sql: Sql): Promise<CountItemsRow | null> {
    const rows = await sql.unsafe(countItemsQuery, []).values();
    if (rows.length !== 1) {
        return null;
    }
    const row = rows[0];
    return {
        count: row[0]
    };
}

export const searchItemsByNameQuery = `-- name: SearchItemsByName :many
SELECT id, name, description, price, quantity, created_at, updated_at FROM items
WHERE name ILIKE '%' || $1 || '%'
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`;

export interface SearchItemsByNameArgs {
    : string | null;
    limit: string;
    offset: string;
}

export interface SearchItemsByNameRow {
    id: string;
    name: string;
    description: string | null;
    price: string;
    quantity: number;
    createdAt: Date;
    updatedAt: Date;
}

export async function searchItemsByName(sql: Sql, args: SearchItemsByNameArgs): Promise<SearchItemsByNameRow[]> {
    return (await sql.unsafe(searchItemsByNameQuery, [args., args.limit, args.offset]).values()).map(row => ({
        id: row[0],
        name: row[1],
        description: row[2],
        price: row[3],
        quantity: row[4],
        createdAt: row[5],
        updatedAt: row[6]
    }));
}

export const createItemQuery = `-- name: CreateItem :one
INSERT INTO items (
    name,
    description,
    price,
    quantity
) VALUES (
    $1, $2, $3, $4
) RETURNING id, name, description, price, quantity, created_at, updated_at`;

export interface CreateItemArgs {
    name: string;
    description: string | null;
    price: string;
    quantity: number;
}

export interface CreateItemRow {
    id: string;
    name: string;
    description: string | null;
    price: string;
    quantity: number;
    createdAt: Date;
    updatedAt: Date;
}

export async function createItem(sql: Sql, args: CreateItemArgs): Promise<CreateItemRow | null> {
    const rows = await sql.unsafe(createItemQuery, [args.name, args.description, args.price, args.quantity]).values();
    if (rows.length !== 1) {
        return null;
    }
    const row = rows[0];
    return {
        id: row[0],
        name: row[1],
        description: row[2],
        price: row[3],
        quantity: row[4],
        createdAt: row[5],
        updatedAt: row[6]
    };
}

export const updateItemQuery = `-- name: UpdateItem :one
UPDATE items
SET
    name = COALESCE($2, name),
    description = COALESCE($3, description),
    price = COALESCE($4, price),
    quantity = COALESCE($5, quantity)
WHERE id = $1
RETURNING id, name, description, price, quantity, created_at, updated_at`;

export interface UpdateItemArgs {
    id: string;
    name: string;
    description: string | null;
    price: string;
    quantity: number;
}

export interface UpdateItemRow {
    id: string;
    name: string;
    description: string | null;
    price: string;
    quantity: number;
    createdAt: Date;
    updatedAt: Date;
}

export async function updateItem(sql: Sql, args: UpdateItemArgs): Promise<UpdateItemRow | null> {
    const rows = await sql.unsafe(updateItemQuery, [args.id, args.name, args.description, args.price, args.quantity]).values();
    if (rows.length !== 1) {
        return null;
    }
    const row = rows[0];
    return {
        id: row[0],
        name: row[1],
        description: row[2],
        price: row[3],
        quantity: row[4],
        createdAt: row[5],
        updatedAt: row[6]
    };
}

export const updateItemQuantityQuery = `-- name: UpdateItemQuantity :one
UPDATE items
SET quantity = quantity + $2
WHERE id = $1
RETURNING id, name, description, price, quantity, created_at, updated_at`;

export interface UpdateItemQuantityArgs {
    id: string;
    quantity: number;
}

export interface UpdateItemQuantityRow {
    id: string;
    name: string;
    description: string | null;
    price: string;
    quantity: number;
    createdAt: Date;
    updatedAt: Date;
}

export async function updateItemQuantity(sql: Sql, args: UpdateItemQuantityArgs): Promise<UpdateItemQuantityRow | null> {
    const rows = await sql.unsafe(updateItemQuantityQuery, [args.id, args.quantity]).values();
    if (rows.length !== 1) {
        return null;
    }
    const row = rows[0];
    return {
        id: row[0],
        name: row[1],
        description: row[2],
        price: row[3],
        quantity: row[4],
        createdAt: row[5],
        updatedAt: row[6]
    };
}

export const deleteItemQuery = `-- name: DeleteItem :exec
DELETE FROM items
WHERE id = $1`;

export interface DeleteItemArgs {
    id: string;
}

export async function deleteItem(sql: Sql, args: DeleteItemArgs): Promise<void> {
    await sql.unsafe(deleteItemQuery, [args.id]);
}

export const getLowStockItemsQuery = `-- name: GetLowStockItems :many
SELECT id, name, description, price, quantity, created_at, updated_at FROM items
WHERE quantity < $1
ORDER BY quantity ASC`;

export interface GetLowStockItemsArgs {
    quantity: number;
}

export interface GetLowStockItemsRow {
    id: string;
    name: string;
    description: string | null;
    price: string;
    quantity: number;
    createdAt: Date;
    updatedAt: Date;
}

export async function getLowStockItems(sql: Sql, args: GetLowStockItemsArgs): Promise<GetLowStockItemsRow[]> {
    return (await sql.unsafe(getLowStockItemsQuery, [args.quantity]).values()).map(row => ({
        id: row[0],
        name: row[1],
        description: row[2],
        price: row[3],
        quantity: row[4],
        createdAt: row[5],
        updatedAt: row[6]
    }));
}

