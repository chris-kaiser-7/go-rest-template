CREATE INDEX IF NOT EXISTS ix_users ON tokens (user_id); 
CREATE INDEX IF NOT EXISTS ix_users ON users (email); 
CREATE INDEX IF NOT EXISTS ix_key_hash ON api_keys (hash); 
CREATE INDEX IF NOT EXISTS ix_apikey_id_bucket ON api_key_usage (key_id, bucket_start DESC); 
