CREATE TABLE IF NOT EXISTS api_keys
(
  id         BIGSERIAL PRIMARY KEY,
  created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
  user_id    BIGINT NOT NULL,
  key_name   TEXT NOT NULL,
  key_hash   INT UNIQUE NOT NULL,
  activated  BOOL DEFAULT TRUE,
  CONSTRAINT user_id FOREIGN KEY (id)
  REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS api_key_usage
(
  id            BIGSERIAL PRIMARY KEY,
  key_id        BIGINT NOT NULL,
  bucket_start  TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(), --TODO: double check this
  usage_count   INT DEFAULT 0
  CONSTRAINT key_id FOREIGN KEY (id)
  REFERENCES api_keys(id)
);


