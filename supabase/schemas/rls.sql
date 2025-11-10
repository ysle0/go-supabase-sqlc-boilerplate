-- Row Level Security (RLS) Policies
-- Example policies for Supabase

-- Enable RLS on tables
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE items ENABLE ROW LEVEL SECURITY;
ALTER TABLE transactions ENABLE ROW LEVEL SECURITY;

-- Users table policies
-- Users can view their own profile
CREATE POLICY "Users can view own profile"
    ON users FOR SELECT
    USING (auth.uid()::text = public_id::text);

-- Users can update their own profile
CREATE POLICY "Users can update own profile"
    ON users FOR UPDATE
    USING (auth.uid()::text = public_id::text);

-- Service role can do everything on users
CREATE POLICY "Service role full access to users"
    ON users
    USING (auth.jwt()->>'role' = 'service_role');

-- Items table policies
-- Anyone can view items (public catalog)
CREATE POLICY "Anyone can view items"
    ON items FOR SELECT
    TO authenticated, anon
    USING (true);

-- Only authenticated users can create items
CREATE POLICY "Authenticated users can create items"
    ON items FOR INSERT
    TO authenticated
    WITH CHECK (true);

-- Only authenticated users can update items
CREATE POLICY "Authenticated users can update items"
    ON items FOR UPDATE
    TO authenticated
    USING (true);

-- Service role can do everything on items
CREATE POLICY "Service role full access to items"
    ON items
    USING (auth.jwt()->>'role' = 'service_role');

-- Transactions table policies
-- Users can view their own transactions
CREATE POLICY "Users can view own transactions"
    ON transactions FOR SELECT
    USING (
        EXISTS (
            SELECT 1 FROM users
            WHERE users.id = transactions.user_id
            AND users.public_id::text = auth.uid()::text
        )
    );

-- Service role can do everything on transactions
CREATE POLICY "Service role full access to transactions"
    ON transactions
    USING (auth.jwt()->>'role' = 'service_role');

-- Note: In production, you should:
-- 1. Customize these policies based on your application's requirements
-- 2. Add more granular policies for different user roles
-- 3. Test policies thoroughly to prevent data leaks
-- 4. Consider performance impact of complex policies
