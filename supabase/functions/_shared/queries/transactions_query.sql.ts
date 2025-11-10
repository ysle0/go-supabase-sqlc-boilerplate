import postgres from "https://deno.land/x/postgresjs@v3.4.7/mod.js";

type Sql = postgres.Sql;
export const getTransactionByIDQuery = `-- name: GetTransactionByID :one
SELECT id, user_id, item_id, transaction_type, quantity, amount, notes, created_at FROM transactions
WHERE id = $1
LIMIT 1`;

export interface GetTransactionByIDArgs {
    id: string;
}

export interface GetTransactionByIDRow {
    id: string;
    userId: string;
    itemId: string;
    transactionType: string;
    quantity: number;
    amount: string;
    notes: string | null;
    createdAt: Date;
}

export async function getTransactionByID(sql: Sql, args: GetTransactionByIDArgs): Promise<GetTransactionByIDRow | null> {
    const rows = await sql.unsafe(getTransactionByIDQuery, [args.id]).values();
    if (rows.length !== 1) {
        return null;
    }
    const row = rows[0];
    return {
        id: row[0],
        userId: row[1],
        itemId: row[2],
        transactionType: row[3],
        quantity: row[4],
        amount: row[5],
        notes: row[6],
        createdAt: row[7]
    };
}

export const listTransactionsQuery = `-- name: ListTransactions :many
SELECT id, user_id, item_id, transaction_type, quantity, amount, notes, created_at FROM transactions
ORDER BY created_at DESC
LIMIT $1 OFFSET $2`;

export interface ListTransactionsArgs {
    limit: string;
    offset: string;
}

export interface ListTransactionsRow {
    id: string;
    userId: string;
    itemId: string;
    transactionType: string;
    quantity: number;
    amount: string;
    notes: string | null;
    createdAt: Date;
}

export async function listTransactions(sql: Sql, args: ListTransactionsArgs): Promise<ListTransactionsRow[]> {
    return (await sql.unsafe(listTransactionsQuery, [args.limit, args.offset]).values()).map(row => ({
        id: row[0],
        userId: row[1],
        itemId: row[2],
        transactionType: row[3],
        quantity: row[4],
        amount: row[5],
        notes: row[6],
        createdAt: row[7]
    }));
}

export const listTransactionsByUserIDQuery = `-- name: ListTransactionsByUserID :many
SELECT id, user_id, item_id, transaction_type, quantity, amount, notes, created_at FROM transactions
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`;

export interface ListTransactionsByUserIDArgs {
    userId: string;
    limit: string;
    offset: string;
}

export interface ListTransactionsByUserIDRow {
    id: string;
    userId: string;
    itemId: string;
    transactionType: string;
    quantity: number;
    amount: string;
    notes: string | null;
    createdAt: Date;
}

export async function listTransactionsByUserID(sql: Sql, args: ListTransactionsByUserIDArgs): Promise<ListTransactionsByUserIDRow[]> {
    return (await sql.unsafe(listTransactionsByUserIDQuery, [args.userId, args.limit, args.offset]).values()).map(row => ({
        id: row[0],
        userId: row[1],
        itemId: row[2],
        transactionType: row[3],
        quantity: row[4],
        amount: row[5],
        notes: row[6],
        createdAt: row[7]
    }));
}

export const listTransactionsByItemIDQuery = `-- name: ListTransactionsByItemID :many
SELECT id, user_id, item_id, transaction_type, quantity, amount, notes, created_at FROM transactions
WHERE item_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`;

export interface ListTransactionsByItemIDArgs {
    itemId: string;
    limit: string;
    offset: string;
}

export interface ListTransactionsByItemIDRow {
    id: string;
    userId: string;
    itemId: string;
    transactionType: string;
    quantity: number;
    amount: string;
    notes: string | null;
    createdAt: Date;
}

export async function listTransactionsByItemID(sql: Sql, args: ListTransactionsByItemIDArgs): Promise<ListTransactionsByItemIDRow[]> {
    return (await sql.unsafe(listTransactionsByItemIDQuery, [args.itemId, args.limit, args.offset]).values()).map(row => ({
        id: row[0],
        userId: row[1],
        itemId: row[2],
        transactionType: row[3],
        quantity: row[4],
        amount: row[5],
        notes: row[6],
        createdAt: row[7]
    }));
}

export const countTransactionsQuery = `-- name: CountTransactions :one
SELECT COUNT(*) FROM transactions`;

export interface CountTransactionsRow {
    count: string;
}

export async function countTransactions(sql: Sql): Promise<CountTransactionsRow | null> {
    const rows = await sql.unsafe(countTransactionsQuery, []).values();
    if (rows.length !== 1) {
        return null;
    }
    const row = rows[0];
    return {
        count: row[0]
    };
}

export const countTransactionsByUserIDQuery = `-- name: CountTransactionsByUserID :one
SELECT COUNT(*) FROM transactions
WHERE user_id = $1`;

export interface CountTransactionsByUserIDArgs {
    userId: string;
}

export interface CountTransactionsByUserIDRow {
    count: string;
}

export async function countTransactionsByUserID(sql: Sql, args: CountTransactionsByUserIDArgs): Promise<CountTransactionsByUserIDRow | null> {
    const rows = await sql.unsafe(countTransactionsByUserIDQuery, [args.userId]).values();
    if (rows.length !== 1) {
        return null;
    }
    const row = rows[0];
    return {
        count: row[0]
    };
}

export const createTransactionQuery = `-- name: CreateTransaction :one
INSERT INTO transactions (
    user_id,
    item_id,
    transaction_type,
    quantity,
    amount,
    notes
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING id, user_id, item_id, transaction_type, quantity, amount, notes, created_at`;

export interface CreateTransactionArgs {
    userId: string;
    itemId: string;
    transactionType: string;
    quantity: number;
    amount: string;
    notes: string | null;
}

export interface CreateTransactionRow {
    id: string;
    userId: string;
    itemId: string;
    transactionType: string;
    quantity: number;
    amount: string;
    notes: string | null;
    createdAt: Date;
}

export async function createTransaction(sql: Sql, args: CreateTransactionArgs): Promise<CreateTransactionRow | null> {
    const rows = await sql.unsafe(createTransactionQuery, [args.userId, args.itemId, args.transactionType, args.quantity, args.amount, args.notes]).values();
    if (rows.length !== 1) {
        return null;
    }
    const row = rows[0];
    return {
        id: row[0],
        userId: row[1],
        itemId: row[2],
        transactionType: row[3],
        quantity: row[4],
        amount: row[5],
        notes: row[6],
        createdAt: row[7]
    };
}

export const getTransactionsByDateRangeQuery = `-- name: GetTransactionsByDateRange :many
SELECT id, user_id, item_id, transaction_type, quantity, amount, notes, created_at FROM transactions
WHERE created_at >= $1 AND created_at <= $2
ORDER BY created_at DESC`;

export interface GetTransactionsByDateRangeArgs {
    createdAt: Date;
    createdAt: Date;
}

export interface GetTransactionsByDateRangeRow {
    id: string;
    userId: string;
    itemId: string;
    transactionType: string;
    quantity: number;
    amount: string;
    notes: string | null;
    createdAt: Date;
}

export async function getTransactionsByDateRange(sql: Sql, args: GetTransactionsByDateRangeArgs): Promise<GetTransactionsByDateRangeRow[]> {
    return (await sql.unsafe(getTransactionsByDateRangeQuery, [args.createdAt, args.createdAt]).values()).map(row => ({
        id: row[0],
        userId: row[1],
        itemId: row[2],
        transactionType: row[3],
        quantity: row[4],
        amount: row[5],
        notes: row[6],
        createdAt: row[7]
    }));
}

export const getUserTransactionSummaryQuery = `-- name: GetUserTransactionSummary :one
SELECT
    user_id,
    COUNT(*) as total_transactions,
    SUM(amount) as total_amount,
    SUM(CASE WHEN transaction_type = 'purchase' THEN 1 ELSE 0 END) as purchase_count,
    SUM(CASE WHEN transaction_type = 'refund' THEN 1 ELSE 0 END) as refund_count
FROM transactions
WHERE user_id = $1
GROUP BY user_id`;

export interface GetUserTransactionSummaryArgs {
    userId: string;
}

export interface GetUserTransactionSummaryRow {
    userId: string;
    totalTransactions: string;
    totalAmount: string;
    purchaseCount: string;
    refundCount: string;
}

export async function getUserTransactionSummary(sql: Sql, args: GetUserTransactionSummaryArgs): Promise<GetUserTransactionSummaryRow | null> {
    const rows = await sql.unsafe(getUserTransactionSummaryQuery, [args.userId]).values();
    if (rows.length !== 1) {
        return null;
    }
    const row = rows[0];
    return {
        userId: row[0],
        totalTransactions: row[1],
        totalAmount: row[2],
        purchaseCount: row[3],
        refundCount: row[4]
    };
}

