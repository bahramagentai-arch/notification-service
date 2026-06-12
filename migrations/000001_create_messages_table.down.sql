DROP TRIGGER IF EXISTS trg_messages_updated_at ON messages;
DROP FUNCTION IF EXISTS set_updated_at();
DROP TABLE IF EXISTS messages;
DROP TYPE IF EXISTS message_channel;
DROP TYPE IF EXISTS message_status;
