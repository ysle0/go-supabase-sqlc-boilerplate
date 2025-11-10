import postgres from "https://deno.land/x/postgresjs@v3.4.7/mod.js";

type Sql = postgres.Sql;
export const getUserByIDQuery = `-- name: GetUserByID :one
SELECT id, public_id, email, username, display_name, created_at, updated_at, deleted_at FROM users
WHERE id = $1 AND deleted_at IS NULL
LIMIT 1`;

export interface GetUserByIDArgs {
    id: string;
}

export interface GetUserByIDRow {
    id: string;
    publicId: string;
    email: string;
    username: string;
    displayName: string | null;
    createdAt: Date;
    updatedAt: Date;
    deletedAt: Date | null;
}

export async function getUserByID(sql: Sql, args: GetUserByIDArgs): Promise<GetUserByIDRow | null> {
    const rows = await sql.unsafe(getUserByIDQuery, [args.id]).values();
    if (rows.length !== 1) {
        return null;
    }
    const row = rows[0];
    return {
        id: row[0],
        publicId: row[1],
        email: row[2],
        username: row[3],
        displayName: row[4],
        createdAt: row[5],
        updatedAt: row[6],
        deletedAt: row[7]
    };
}

export const getUserByPublicIDQuery = `-- name: GetUserByPublicID :one
SELECT id, public_id, email, username, display_name, created_at, updated_at, deleted_at FROM users
WHERE public_id = $1 AND deleted_at IS NULL
LIMIT 1`;

export interface GetUserByPublicIDArgs {
    publicId: string;
}

export interface GetUserByPublicIDRow {
    id: string;
    publicId: string;
    email: string;
    username: string;
    displayName: string | null;
    createdAt: Date;
    updatedAt: Date;
    deletedAt: Date | null;
}

export async function getUserByPublicID(sql: Sql, args: GetUserByPublicIDArgs): Promise<GetUserByPublicIDRow | null> {
    const rows = await sql.unsafe(getUserByPublicIDQuery, [args.publicId]).values();
    if (rows.length !== 1) {
        return null;
    }
    const row = rows[0];
    return {
        id: row[0],
        publicId: row[1],
        email: row[2],
        username: row[3],
        displayName: row[4],
        createdAt: row[5],
        updatedAt: row[6],
        deletedAt: row[7]
    };
}

export const getUserByEmailQuery = `-- name: GetUserByEmail :one
SELECT id, public_id, email, username, display_name, created_at, updated_at, deleted_at FROM users
WHERE email = $1 AND deleted_at IS NULL
LIMIT 1`;

export interface GetUserByEmailArgs {
    email: string;
}

export interface GetUserByEmailRow {
    id: string;
    publicId: string;
    email: string;
    username: string;
    displayName: string | null;
    createdAt: Date;
    updatedAt: Date;
    deletedAt: Date | null;
}

export async function getUserByEmail(sql: Sql, args: GetUserByEmailArgs): Promise<GetUserByEmailRow | null> {
    const rows = await sql.unsafe(getUserByEmailQuery, [args.email]).values();
    if (rows.length !== 1) {
        return null;
    }
    const row = rows[0];
    return {
        id: row[0],
        publicId: row[1],
        email: row[2],
        username: row[3],
        displayName: row[4],
        createdAt: row[5],
        updatedAt: row[6],
        deletedAt: row[7]
    };
}

export const getUserByUsernameQuery = `-- name: GetUserByUsername :one
SELECT id, public_id, email, username, display_name, created_at, updated_at, deleted_at FROM users
WHERE username = $1 AND deleted_at IS NULL
LIMIT 1`;

export interface GetUserByUsernameArgs {
    username: string;
}

export interface GetUserByUsernameRow {
    id: string;
    publicId: string;
    email: string;
    username: string;
    displayName: string | null;
    createdAt: Date;
    updatedAt: Date;
    deletedAt: Date | null;
}

export async function getUserByUsername(sql: Sql, args: GetUserByUsernameArgs): Promise<GetUserByUsernameRow | null> {
    const rows = await sql.unsafe(getUserByUsernameQuery, [args.username]).values();
    if (rows.length !== 1) {
        return null;
    }
    const row = rows[0];
    return {
        id: row[0],
        publicId: row[1],
        email: row[2],
        username: row[3],
        displayName: row[4],
        createdAt: row[5],
        updatedAt: row[6],
        deletedAt: row[7]
    };
}

export const listUsersQuery = `-- name: ListUsers :many
SELECT id, public_id, email, username, display_name, created_at, updated_at, deleted_at FROM users
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2`;

export interface ListUsersArgs {
    limit: string;
    offset: string;
}

export interface ListUsersRow {
    id: string;
    publicId: string;
    email: string;
    username: string;
    displayName: string | null;
    createdAt: Date;
    updatedAt: Date;
    deletedAt: Date | null;
}

export async function listUsers(sql: Sql, args: ListUsersArgs): Promise<ListUsersRow[]> {
    return (await sql.unsafe(listUsersQuery, [args.limit, args.offset]).values()).map(row => ({
        id: row[0],
        publicId: row[1],
        email: row[2],
        username: row[3],
        displayName: row[4],
        createdAt: row[5],
        updatedAt: row[6],
        deletedAt: row[7]
    }));
}

export const countUsersQuery = `-- name: CountUsers :one
SELECT COUNT(*) FROM users
WHERE deleted_at IS NULL`;

export interface CountUsersRow {
    count: string;
}

export async function countUsers(sql: Sql): Promise<CountUsersRow | null> {
    const rows = await sql.unsafe(countUsersQuery, []).values();
    if (rows.length !== 1) {
        return null;
    }
    const row = rows[0];
    return {
        count: row[0]
    };
}

export const createUserQuery = `-- name: CreateUser :one
INSERT INTO users (
    email,
    username,
    display_name
) VALUES (
    $1, $2, $3
) RETURNING id, public_id, email, username, display_name, created_at, updated_at, deleted_at`;

export interface CreateUserArgs {
    email: string;
    username: string;
    displayName: string | null;
}

export interface CreateUserRow {
    id: string;
    publicId: string;
    email: string;
    username: string;
    displayName: string | null;
    createdAt: Date;
    updatedAt: Date;
    deletedAt: Date | null;
}

export async function createUser(sql: Sql, args: CreateUserArgs): Promise<CreateUserRow | null> {
    const rows = await sql.unsafe(createUserQuery, [args.email, args.username, args.displayName]).values();
    if (rows.length !== 1) {
        return null;
    }
    const row = rows[0];
    return {
        id: row[0],
        publicId: row[1],
        email: row[2],
        username: row[3],
        displayName: row[4],
        createdAt: row[5],
        updatedAt: row[6],
        deletedAt: row[7]
    };
}

export const updateUserQuery = `-- name: UpdateUser :one
UPDATE users
SET
    email = COALESCE($2, email),
    username = COALESCE($3, username),
    display_name = COALESCE($4, display_name)
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, public_id, email, username, display_name, created_at, updated_at, deleted_at`;

export interface UpdateUserArgs {
    id: string;
    email: string;
    username: string;
    displayName: string | null;
}

export interface UpdateUserRow {
    id: string;
    publicId: string;
    email: string;
    username: string;
    displayName: string | null;
    createdAt: Date;
    updatedAt: Date;
    deletedAt: Date | null;
}

export async function updateUser(sql: Sql, args: UpdateUserArgs): Promise<UpdateUserRow | null> {
    const rows = await sql.unsafe(updateUserQuery, [args.id, args.email, args.username, args.displayName]).values();
    if (rows.length !== 1) {
        return null;
    }
    const row = rows[0];
    return {
        id: row[0],
        publicId: row[1],
        email: row[2],
        username: row[3],
        displayName: row[4],
        createdAt: row[5],
        updatedAt: row[6],
        deletedAt: row[7]
    };
}

export const softDeleteUserQuery = `-- name: SoftDeleteUser :exec
UPDATE users
SET deleted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL`;

export interface SoftDeleteUserArgs {
    id: string;
}

export async function softDeleteUser(sql: Sql, args: SoftDeleteUserArgs): Promise<void> {
    await sql.unsafe(softDeleteUserQuery, [args.id]);
}

export const hardDeleteUserQuery = `-- name: HardDeleteUser :exec
DELETE FROM users
WHERE id = $1`;

export interface HardDeleteUserArgs {
    id: string;
}

export async function hardDeleteUser(sql: Sql, args: HardDeleteUserArgs): Promise<void> {
    await sql.unsafe(hardDeleteUserQuery, [args.id]);
}

