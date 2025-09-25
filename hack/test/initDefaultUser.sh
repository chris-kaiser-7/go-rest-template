#!/usr/bin/env bash

echo "INSERT INTO users (name, email, password_hash, activated)
VALUES (
	'admin',
	'admin@admin.com',
	'\\x243261243132244261505345396b423956744a6b4a5433464c67656c2e695933676c6453414e696b4731504b686c414f6f4a4535424c4174562f4c4f', 
	true
);

INSERT INTO tokens (hash, user_id, expiry, scope)
VALUES (
	'\\xa83f22476a615b9d400c6a79d4386e668984bdf6d6589c2028f1b67b300ec0ea',
	1,
	TIMESTAMP '2030-12-31 12:00:00',
	'authentication'
);

INSERT INTO users_permissions (user_id, permission_id)
VALUES (1, 1),(1, 2),(1, 3),(1, 4);"
