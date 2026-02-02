CREATE TABLE IF NOT EXISTS users (
  id            BIGSERIAL PRIMARY KEY,
  login         TEXT NOT NULL UNIQUE,
  password_hash BYTEA NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS accounts (
  user_id   BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  current   NUMERIC(20,4) NOT NULL DEFAULT 0,
  withdrawn NUMERIC(20,4) NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS orders (
  number          TEXT PRIMARY KEY,
  user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  status          TEXT NOT NULL,
  accrual         NUMERIC(20,4),
  accrual_applied BOOLEAN NOT NULL DEFAULT FALSE,
  uploaded_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_orders_user_uploaded_at ON orders(user_id, uploaded_at DESC);
CREATE INDEX IF NOT EXISTS idx_orders_status_uploaded_at ON orders(status, uploaded_at ASC);

CREATE TABLE IF NOT EXISTS withdrawals (
  id           BIGSERIAL PRIMARY KEY,
  user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  order_number TEXT NOT NULL UNIQUE,
  sum          NUMERIC(20,4) NOT NULL,
  processed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_withdrawals_user_processed_at ON withdrawals(user_id, processed_at DESC);

