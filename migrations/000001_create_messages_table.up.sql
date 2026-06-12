CREATE TYPE message_status AS ENUM (
    'PENDING',
    'SENDING',
    'SENT',
    'FAILED'
);

CREATE TYPE message_channel AS ENUM (
    'SMS'
);


CREATE TABLE messages
(
    id              TEXT PRIMARY KEY,
    idempotency_key TEXT            NOT NULL UNIQUE,
    body            TEXT            NOT NULL,
    recipient       TEXT            NOT NULL,
    channel         message_channel NOT NULL DEFAULT 'SMS',
    priority        SMALLINT        NOT NULL DEFAULT 0,
    status          message_status  NOT NULL DEFAULT 'PENDING',
    attempts        SMALLINT        NOT NULL DEFAULT 0,
    max_attempts    SMALLINT        NOT NULL DEFAULT 3,
    provider_msg_id TEXT            NOT NULL DEFAULT '',
    failure_reason  TEXT            NOT NULL DEFAULT '',
    created_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE INDEX idx_messages_status_priority
    ON messages (status, priority DESC, created_at ASC);

CREATE INDEX idx_messages_idempotency_key
    ON messages (idempotency_key);

CREATE
OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    NEW.updated_at
= NOW();
RETURN NEW;
END;
$$;

CREATE TRIGGER trg_messages_updated_at
    BEFORE UPDATE
    ON messages
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
