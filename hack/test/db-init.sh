#!/usr/bin/env bash

# INSERT INTO users (name, email, password_hash, activated)
# VALUES (
# 	'default user',
# 	'defaultemail@gmail.com',
# 	E'\\x243261243132244261505345396b423956744a6b4a5433464c67656c2e695933676c6453414e696b4731504b686c414f6f4a4535424c4174562f4c4f', 
# 	true
# );
#
# 6UTN57UN5XXWRUB2XCYZPEWE44

echo "
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS tokens;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS users_permissions;
DROP TABLE IF EXISTS api_key_usage;
DROP TABLE IF EXISTS api_keys;

CREATE TABLE IF NOT EXISTS users
(
	id            BIGSERIAL PRIMARY KEY,
	created_at    TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
	name          TEXT                        NOT NULL,
	email         CITEXT UNIQUE               NOT NULL,
	password_hash BYTEA                       NOT NULL,
	activated     BOOL                        NOT NULL,
	version       INTEGER                     NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS tokens
(
	hash    BYTEA PRIMARY KEY,
	user_id BIGINT                      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	expiry  TIMESTAMP(0) WITH TIME ZONE NOT NULL,
	scope   TEXT                        NOT NULL
);

CREATE TABLE IF NOT EXISTS permissions
(
	id   BIGSERIAL PRIMARY KEY,
	code TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS users_permissions
(
	user_id       BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	permission_id BIGINT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
	PRIMARY KEY (user_id, permission_id)
);

INSERT INTO permissions (code)
VALUES ('keys:get'), ('keys:add'), ('keys:delete'), ('keys:admin');

CREATE TABLE IF NOT EXISTS api_keys
(
  id         BIGSERIAL PRIMARY KEY,
  created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
  user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  key_name   TEXT NOT NULL,
  key_hash   BYTEA UNIQUE NOT NULL,
  activated  BOOL DEFAULT TRUE
);


CREATE TABLE IF NOT EXISTS api_key_usage
(
  id            BIGSERIAL PRIMARY KEY,
  key_id        BIGINT NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
  bucket_start  TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
  usage_count   INT DEFAULT 1
);

CREATE INDEX IF NOT EXISTS ix_users ON tokens (user_id); 
CREATE INDEX IF NOT EXISTS ix_users ON users (email); 
CREATE INDEX IF NOT EXISTS ix_key_hash ON api_keys (key_hash); 
CREATE INDEX IF NOT EXISTS ix_apikey_id_bucket ON api_key_usage (key_id, bucket_start DESC); 
"
