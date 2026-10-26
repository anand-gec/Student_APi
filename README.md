# Student API

A Go REST API for managing student records with SQLite persistence. Admins can register and sign in to receive a JWT, which is required for all student-management endpoints.

## Features

- Create, list, retrieve, update, and delete students
- Register and log in admins; passwords are stored as bcrypt hashes
- Generate a time-limited password-reset token and reset an admin password
- Protect student endpoints with JWT bearer authentication
- Serve an embedded OpenAPI specification and Swagger UI
- Handle interrupt and termination signals for server shutdown

## Requirements

- Go 1.27.0 or later, as specified in [`go.mod`](./go.mod)
- A writable directory and file path for the SQLite database

## Configuration

The application reads YAML configuration from the path supplied by `-config` or the `CONFIG_PATH` environment variable. `CONFIG_PATH` takes precedence when both are set. The YAML fields are:

| Field | Description |
| --- | --- |
| `env` | Environment name, included in startup logs |
| `storage_path` | SQLite database file path |
| `http_server.address` | HTTP listen address, such as `localhost:8082` |
| `jwt_secret` | Secret used to sign and validate JWTs; set a private, sufficiently random value |

For example:

```yaml
env: "dev"
storage_path: "storage/students.db"
http_server:
  address: "localhost:8082"
jwt_secret: "replace-with-a-private-random-secret"
```

The checked-in [`config/local.yaml`](./config/local.yaml) currently sets `storage_path` to `storage/storage.go`, which is a Go source file, not a database file. Before starting the service with that config, change it to a writable database path such as `storage/students.db`. The `storage` directory already exists in this repository.

`ENV` and `JWT_SECRET` can override the corresponding YAML values. The storage path and listen address are read from YAML.

## Run

From the repository root, after correcting the database path as described above:

```sh
go run ./cmd/students-api -config config/local.yaml
```

Or specify the config path in the environment:

```sh
CONFIG_PATH=config/local.yaml go run ./cmd/students-api
```

The [`Makefile`](./Makefile) also provides a `run` target:

```sh
make run
```

With the example address, the API listens at `http://localhost:8082`. The application handles interrupt and termination signals and attempts a bounded shutdown.

On startup, SQLite creates the `students`, `admins`, and `password_resets` tables if they do not already exist.

## API

All request and response bodies use JSON. Student fields are `name` (string), `email` (string), and `age` (integer); all are required on create and update. Student responses include the generated `id`.

### Admin endpoints

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/api/admin/register` | Register an admin |
| `POST` | `/api/admin/login` | Verify credentials and return a JWT |
| `POST` | `/api/admin/forgot-password` | Generate a password-reset token |
| `POST` | `/api/admin/reset-password` | Set a new password using a reset token |

Register and login request:

```json
{
  "email": "admin@example.com",
  "password": "secret123"
}
```

Successful login returns HTTP `200` with a token:

```json
{
  "status": "success",
  "token": "<jwt-token>"
}
```

Registration returns HTTP `201` with the new `Admin_id`. Duplicate email addresses or storage errors return an error.

To request a reset token:

```json
{
  "email": "admin@example.com"
}
```

The intended token lifetime is 15 minutes, and the token is returned in the response as `reset_token`. No email is sent by this application. Note that the current SQLite implementation creates a `password_resets` table but attempts to insert into `password_reset` (singular), so requesting a reset token currently fails until those table names are made consistent. To reset the password once that issue is corrected, submit:

```json
{
  "token": "<reset-token>",
  "new_password": "newsecret123"
}
```

The reset endpoint requires a password of at least six characters.

### Student endpoints

Every student endpoint requires a valid admin JWT.

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/api/students` | Create a student |
| `GET` | `/api/students` | List all students |
| `GET` | `/api/students/{id}` | Get a student by ID |
| `PUT` | `/api/students/{id}` | Replace all student fields |
| `DELETE` | `/api/students/{id}` | Delete a student by ID |
| `DELETE` | `/api/students` | Delete all students |

Create and update request:

```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "age": 21
}
```

New student creation returns HTTP `201` with the generated `id`. Updates return HTTP `200`; deleting a student or deleting all students returns HTTP `200` with a status message. Invalid IDs and invalid or missing student fields return HTTP `400`. Updating a nonexistent student returns HTTP `404`.

### Authentication

Pass the login token in the `Authorization` header using the bearer scheme:

```http
Authorization: Bearer <jwt-token>
```

Tokens are signed with the configured JWT secret and expire after 24 hours. Missing or invalid tokens return HTTP `401`.

## Swagger UI and OpenAPI

With the service running, open [`http://localhost:8082/swagger/`](http://localhost:8082/swagger/). The embedded OpenAPI document is available at [`http://localhost:8082/swagger/openapi.yaml`](http://localhost:8082/swagger/openapi.yaml). Swagger UI assets are loaded from a CDN, so viewing the interactive UI requires an internet connection.

## Project layout

```text
.
├── cmd/students-api/       # Application entry point and HTTP routes
├── config/                 # Local YAML configuration
├── internal/config/        # Configuration loading and JWT helpers
├── internal/http/          # Swagger UI, OpenAPI, handlers, and middleware
├── internal/storage/       # Storage interface and SQLite implementation
├── internal/types/         # Student and admin data types
├── internal/utils/         # JSON response helpers
├── Makefile
├── go.mod
└── README.md
```

## Development

Run the Go test suite with:

```sh
go test ./...
```

Use a private JWT secret and avoid exposing password-reset tokens. The current reset endpoint returns the token directly for local testing, and admin registration is publicly reachable; review these behaviors before exposing the API to untrusted networks.
