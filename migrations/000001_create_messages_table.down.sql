CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    body TEXT NOT NULL,
    recipient TEXT NOT NULL,
    channel TEXT NOT NULL,
    priority INTEGER NOT NULL,
    status TEXT NOT NULL,
    attempts INTEGER NOT NULL,
    max_attempts INTEGER NOT NULL,
    provider_msg_id TEXT NOT NULL,
    failure_reason TEXT NOT NULL,
    idempotency_key TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);