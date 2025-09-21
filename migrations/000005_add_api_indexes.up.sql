CREATE INDEX IF NOT EXISTS ix_apikey_id_ts ON api_key_usage (key_id, bucket_start DESC); 
