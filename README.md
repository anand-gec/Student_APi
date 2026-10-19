# Student API

A lightweight Go REST API for managing student records with SQLite persistence and JWT-protected admin endpoints.

## Features

- Create, list, fetch by ID, and delete student records
- SQLite database storage
- Admin registration and login flow
- JWT-based authentication for student-related API routes
- JSON response helpers and request validation

## Tech Stack

- Go
- SQLite
- JWT (`golang-jwt`)
- `net/http` router
- `cleanenv` for configuration
- `validator` for request validation

## Project Structure

```text
.
├── cmd/
│   └── students-api/
│       └── main.go
├── config/
│   └── local.yaml
├── internal/
│   ├── config/
│   ├── http/
│   ├── storage/
│   ├── types/
│   └── utils/
├── storage/
│   └── storage.db
├── go.mod
├── go.sum
└── README.md
```

## Prerequisites

- Go 1.27+ (or a version compatible with the module definition in `go.mod`)
- A SQLite database file or a writable database path

## Configuration

The app loads configuration from a YAML file. Set the path through the `CONFIG_PATH` environment variable or by passing `-config` when running the service.

Example configuration file:

```yaml
env: "dev"
storage_path: "storage/storage.db"
http_server:
  address: "localhost:8082"
jwt_secret: "your-secret-key"
```

## Running the Application

```bash
go mod download
go run ./cmd/students-api
```

Or with config explicitly provided:

```bash
CONFIG_PATH=config/local.yaml go run ./cmd/students-api
```

The server starts on the address defined in the config file, for example:

```text
http://localhost:8082
```

## API Endpoints

### Admin Authentication

#### Register an admin

```http
POST /api/admin/register
Content-Type: application/json
```

Request body:

```json
{
  "email": "admin@example.com",
  "password": "secret123"
}
```

#### Login an admin

```http
POST /api/admin/login
Content-Type: application/json
```

Request body:

```json
{
  "email": "admin@example.com",
  "password": "secret123"
}
```

Response:

```json
{
  "status": "success",
  "token": "<jwt-token>"
}
```

### Student Routes

All student routes are protected with JWT-based admin authorization.

#### Create a student

```http
POST /api/students
Authorization: Bearer <jwt-token>
Content-Type: application/json
```

Request body:

```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "age": 21
}
```

#### Get all students

```http
GET /api/students
Authorization: Bearer <jwt-token>
```

#### Get a student by ID

```http
GET /api/students/{id}
Authorization: Bearer <jwt-token>
```

#### Delete a student by ID

```http
DELETE /api/students/{id}
Authorization: Bearer <jwt-token>
```

#### Delete all students

```http
DELETE /api/students
Authorization: Bearer <jwt-token>
```

## Authentication

Protected routes require an `Authorization` header in this format:

```http
Authorization: Bearer <jwt-token>
```

If the token is missing or invalid, the API responds with an unauthorized error.

## Notes

- The app uses SQLite tables named `students` and `admins`.
- Admin passwords are hashed before being stored in the database.
- The project is designed as a simple CRUD API for learning and internal use.
