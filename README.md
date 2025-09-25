# Go API Key Manager

## Requirements
- Go version >1.24. (Older version will probably work but I have not tested them.)
- Docker for using ephemeral db
- GNU Make 4.3 (Preinstalled on most Linux distros.) 
- curl for using the api
- ~SQL migrate for deploying sql schema to database~ I have disabled migrate for convenience of test setup.

## Basic Usage
This project uses make to enable easy usage. Details on make commands can be found in the Makefile. 
Also the command `make help` will display descriptions of the make commands. Here are some commands.
`make run/api` will create an ephemeral Postgres DB and start the API

`make test/data` will run the tests for the data and CRUD layer. 
These tests test against the docker database and clean the database tables before each test to isolate test logic.

`make test/api` will run the tests for endpoints.
These tests setup a mock client that make url requests to the multiplexer (router).
These tests **do not** clean the database tables before each test.

`make test/api` will run all tests. data tests followed by api tests.

## Example Usage
First start by running
`make run/api` or `sudo make run/api` if docker needs permissions.
When you see 
```
{"level":"INFO","time":"2025-09-23T15:17:30Z","message":"starting server","properties":{"addr":":4000","env":"development"}}
```
The api is ready to receive on localhost:4000

**Note**
A user is required for auth to use ApiKey endpoints
For testing a default admin user and token is created with the credentials
```
username: admin@admin.com
password: password
token: 6UTN57UN5XXWRUB2XCYZPEWE44
```
So instead of creating your own user you can just run `TOKEN=6UTN57UN5XXWRUB2XCYZPEWE44`.

### create user
To create a user you can run the following commands
```
BODY='{"name":"testUser","email":"testEmail@gmail.com","password":"testpass"}'
curl -X POST localhost:4000/v1/users -H "Content-Type: application/json" -d "$BODY"
```

The response should look like this
```
{
        "user": {
                "id": 1,
                "created_at": "2025-09-23T15:23:32Z",
                "name": "testUser",
                "email": "testEmail@gmail.com",
                "activated": true
        }
}
```

Then get a token by using the commands
```
BODY='{"email": "testEmail@gmail.com","password":"testpass"}'
curl -X POST localhost:4000/v1/tokens/authentication -H "Content-Type: application/json" -d "$BODY"
```

The response should generate a token and look like this
```
{
        "authentication_token": {
                "token": "NCDA5BAD3DF746ZIFKWTZB424Y",
                "expiry": "2025-09-24T11:26:38.203904074-04:00"
        }
}
```
This token will be used to make Api Key requests.
To save to token run `TOKEN=<your token value here>`


### API Key Endpoints

To create a key use
```
BODY='{ "keyName": "testKey" }'
curl -X POST localhost:4000/v1/keys \
-H "Content-Type: application/json" \
-H "Authorization: Bearer ${TOKEN}" \
-d "$BODY"
```

Your response should look something like
```
{
        "apiKey": {
                "id": 2,
                "user_id": 1,
                "key_name": "testKey",
                "key_value": "LSpLUw3YCq+6Sv1DOyXcWUx/oJiugc+6ni+EeOa5ub+NC0OgeX7IeDuYa/a5a80i",
                "uses": 0
        }
}
```
To save your API key for future commands run `APIKEY=<your key_value here>`

To validate you key use /v1/keys/validate
```
BODY="{ \"key\": \"${APIKEY}\" }"
curl -X POST localhost:4000/v1/keys/validate \
-H "Content-Type: application/json" \
-H "Authorization: Bearer ${TOKEN}" \
-d "$BODY"
```
The response should look something like
```
{
        "apiKey": {
                "id": 2,
                "user_id": 1,
                "key_name": "testKey",
                "key_value": "LSpLUw3YCq+6Sv1DOyXcWUx/oJiugc+6ni+EeOa5ub+NC0OgeX7IeDuYa/a5a80i",
                "uses": 1
        }
}
```

To delete an API you can use /v1/keys/id where id is the id retruned by apikey
```
curl -X DELETE localhost:4000/v1/keys/<key_id> \
-H "Authorization: Bearer ${TOKEN}"
```
The response should look something like
```
{
        "message": "API key succesfully deactivated."
}
```

To get the details on all your keys you can use
```
curl -X GET localhost:4000/v1/keys \
-H "Authorization: Bearer ${TOKEN}"
```
you can also provide pagination data for the query params "page", "page_size", and "sort". 
Accepted sort values are "id" for id ascending and "-id" for id descending

```
curl -X GET "localhost:4000/v1/keys?page=2&page_size=2&sort=-id" -H "Authorization: Bearer ${TOKEN}"
```

This will return something like this:
```
        "ApiKeys": [
                {
                        "id": 2,
                        "user_id": 1,
                        "key_name": "testKey1",
                        "key_value": "",
                        "uses": 2 
                },
                {
                        "id": 3,
                        "user_id": 1,
                        "key_name": "testKey2",
                        "key_value": "",
                        "uses": 5
                },
                {
                        "id": 4,
                        "user_id": 1,
                        "key_name": "testKey3",
                        "key_value": "",
                        "uses": 0
                }
        ],
        "metadata": {
                "current_page": 1,
                "page_size": 20,
                "first_page": 1,
                "last_page": 1,
                "total_records": 3
        }
```
The metadata field provides client pagination data

## Database Models

## Production Deployment

## Benchmarks

## Future Enhancements


## Greenlight
The base of this project is forked from a template by Alex Edwards and further enhanced by Chris Kaiser

Greenlight is a simple but flexible project by Alex Edwards (https://github.com/codeaucafe/greenlight) 
that I have been using as a template for REST projects. 
The project sets up an idiomatic folder and code structure. 
Has features such as:
- multiplexer and route table in /cmd/api/routes.go
- CRUD's and endpoints for Users, Tokens, and Permissions  
- A smtp mailer interface to verify user email addresses
- Database configuration
- SQL Migration setup
- Rate Limiter
- CORS Configuration

I am working to add additional features where this project is lacking to perfect a template to use for REST projects. 
Mainly a test suite and benchmarking suite are not included in this project that I want to add. 

