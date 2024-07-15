# TimeTrack

TimeTrack is a simple time tracking tool that allows you to track your time spent on different tasks.

It is a multi-user server application that stores the time tracking data in a PostgreSQL database and communicates with
(pseudo-) external services to provide additional information about the users.

It was created as a test assignment for a job application in July 2024.

## Features

- **Track time**

  Start and stop timers for tasks <mark>quickly</mark> and <mark>safely</mark>, even with many users at the same time.

- **Generate reports**

  Generate reports for the time spent on tasks in a specific time frame.

- **Manage tasks**

  Create, view, update, and delete tasks.

- **Manage users**

  Register users using their national ID number, view all users with <mark>filtering</mark> options and
  <mark>pagination</mark>, and update or delete user information as needed.

  When registering a user, the application will try to fetch additional information about the user from a (pseudo-)
  <mark>external service</mark>.

  The application uses <mark>token-based authentication</mark> to ensure that only registered users can access the
  application. For simplicity, the application uses user IDs as tokens.

## Architecture

The application provides three executables:

- [`cmd/server`](cmd/server): The main entry point for the server application.
- [`cmd/database-up`](cmd/database-up): A helper tool to migrate the database schema using
  [migrate](https://github.com/golang-migrate/migrate).
- [`cmd/peopleinfoserver`](cmd/peopleinfoserver): A mock server that provides additional information about users. It
  implements an [OpenAPI specification](api/peopleinfo/v1/openapi.yaml) provided as part of the assignment.

### REST API

The application provides a REST API with the following endpoints:

- **Authentication.** Implemented in the [`auth`](internal/auth) package.

  - `POST /auth`: Authenticate a user using OAuth 2.0 password grant. For simplicity, the national ID number is used
    both as the username and password. The response contains an access token that can be used to authenticate further.

- **Users.** Implemented in the [`user`](internal/user) package.

  - `GET /users`: List all users. Supports filtering by national ID number and pagination.
  - `POST /users`: Register a new user.
  - `GET /users/{id}`: Get information about a specific user.
  - `PUT /users/{id}`: Update information about a specific user.
  - `DELETE /users/{id}`: Delete a specific user.

- **Tasks.** Implemented in the [`task`](internal/task) package.

  - `GET /tasks`: List all tasks. Supports pagination.
  - `POST /tasks`: Create a new task.
  - `GET /tasks/{id}`: Get information about a specific task.
  - `PUT /tasks/{id}`: Update information about a specific task.
  - `DELETE /tasks/{id}`: Delete a specific task.

- **Time tracking.** Implemented in the [`tracking`](internal/tracking) package.

  - `POST /tasks/{id}/start`: Start a timer for a specific task with authenticated user.
  - `POST /tasks/{id}/stop`: Stop the timer for a specific task with authenticated user.

- **Time reporting.** Implemented in the [`reporting`](internal/reporting) package.

  - `POST /users/{id}/report`: Generate a report for the time spent on tasks by a specific user in a specific time frame.

The API is documented in [`api/timetrack/v1/openapi.yaml`](api/timetrack/v1/openapi.yaml) and implemented using
[1.22 net/http](https://pkg.go.dev/net/http). The server uses
[oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) to generate
[timetrackapi.ServerInterface](api/timetrackapi/v1/timetrackapi.go) interface with the API handlers. To improve the
architecture, each module implements the relevant API handlers in a struct in `handler.go` file. These structs are then
embedded in the main `Handler` struct in [`app/api/handler.go`](internal/app/api/handler.go) and used in the
[`app/api/server.go`](internal/app/api/server.go) to create the server.

### Database

TBA.

## Usage

### Manual

```sh
go run ./cmd/peopleinfoserver &
go run ./cmd/database-up
go run ./cmd/server
```

### Docker Compose

```sh
docker compose up
```

## Testing

### Manual

```sh
# Spin up and migrate a new test database.
CUSTOM_DATABASE_PORT=5433 docker compose -p timetrack-test up database database-up -d

# Set environment variables from example.env.
...

# Run tests.
make test

# Tear down the test database.
docker compose -p timetrack-test down -v
```

## Demo

TBA.
