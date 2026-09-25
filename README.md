# scaffold-go

A Go CLI tool for scaffolding microservice projects following clean architecture principles.

## Features

- 🏗️ **Clean Architecture**: Generates projects with Domain, EntryPoints, and DrivenAdapters layers
- 🚀 **Quick Start**: Initialize projects with a single command
- 🔧 **Layered component generation**: `entity` → `usecase` → `ep` (entry-points) → `da` (driven adapters)
- 📦 **Dependency Injection**: Uses uber-go/fx for DI
- 📝 **Configuration**: Viper-based configuration management
- 📊 **Logging**: Structured logging with Zap
- 🗄️ **Multiple Databases**: Support for PostgreSQL, MongoDB, and DynamoDB

## Installation

```bash
go install github.com/gabheadz/scaffold-go@latest
```

Or build from source:

```bash
git clone https://github.com/gabheadz/scaffold-go.git
cd scaffold-go
go build -o scaffold-go cmd/main.go
```

## Quick Start

### 1. Initialize a new project

```bash
scaffold-go init my-service --module github.com/user/my-service
cd my-service
go mod tidy
```

### 2. Generate a domain entity

```bash
scaffold-go generate entity User --fields Name:string,Email:string,Age:int
```

### 3. Generate a use case

```bash
scaffold-go generate usecase CreateUser --dependencies userRepo:UserRepository
```

### 4. Generate an HTTP entry-point

```bash
scaffold-go generate ep http User \
  --usecases createUser:CreateUserUseCase \
  --routes POST:/users:CreateUser
```

### 5. Generate a database driven adapter

```bash
scaffold-go generate da database UserRepo \
  --type postgres \
  --entity User \
  --fields Name:string,Email:string,Age:int
```

## Commands

### Global Flags

All commands support these global flags:

```bash
  -v, --verbose         Enable verbose output with detailed debug logging
      --config string   Config file (default is $HOME/.scaffold-go.yaml)
```

**Verbose Mode**: When enabled with `-v` or `--verbose`, the CLI outputs detailed debug information including:
- Step-by-step execution progress
- Template processing details
- File operations (create, write, format)
- Validation checks
- Structured context fields (timestamps, component names, file paths)

**Production Mode** (default): Outputs structured JSON logs with INFO level and above, suitable for log aggregation systems.

### init

Initialize a new microservice project.

```bash
scaffold-go init [project-name] [flags]

Flags:
  -m, --module string   Go module path (e.g., github.com/user/project)
  -o, --output string   Output directory (default ".")
```

### generate entity

Generate a domain entity with fields and repository interface.

```bash
scaffold-go generate entity [name] --fields Field1:type1,Field2:type2

Flags:
  -f, --fields strings   Entity fields in format Name:type (comma-separated)
```

### generate usecase

Generate a domain use case with interface and implementation.

```bash
scaffold-go generate usecase [name] --dependencies dep1:Type1,dep2:Type2

Flags:
  -d, --dependencies strings   Dependencies in format name:Type (comma-separated)
```

### generate ep

Group command for **entry-points** (the components that drive the application
from the outside). You must pick a kind:

```bash
scaffold-go generate ep <kind> [name] [flags]

Kinds:
  http        HTTP handler using the Fiber framework
  consumer    Message queue consumer
```

#### generate ep http

```bash
scaffold-go generate ep http [name] \
  --usecases uc1:Type1 \
  --routes METHOD:path:handler

Flags:
  -u, --usecases strings   Use cases in format name:Type (comma-separated)
  -r, --routes strings     Routes in format METHOD:path:handler (comma-separated)
      --enable-cors           Enable CORS middleware (default true)
      --enable-rate-limiter   Enable rate-limiter middleware (default true)
      --enable-error-handler  Enable custom error handler (default true)
```

#### generate ep consumer

Broker-aware message consumer. The CLI emits the consumer file, a per-component
fx module, and (on first use per broker) the shared infrastructure module that
provides the broker client. `cmd/main.go` is patched automatically between the
`// scaffold-go:imports:begin/end` and `// scaffold-go:modules:begin/end`
markers so the new module is wired into `fx.New(...)` with no manual edits.

```bash
scaffold-go generate ep consumer [name] \
  --broker rabbitmq|ibmmq|kafka \
  --pattern <pattern> \
  --usecases uc1:Type1 \
  --queue <queue|topic|event>

Flags:
  -b, --broker string      Messaging broker (required): rabbitmq | ibmmq | kafka
  -p, --pattern string     Messaging pattern (required); see table below
  -u, --usecases strings   Use cases in format name:Type[:Package] (comma-separated)
  -q, --queue string       Queue, topic, event, or resource name; defaults to lowercased component name
```

| Broker      | Patterns                                  | Library                                                                       |
|-------------|-------------------------------------------|-------------------------------------------------------------------------------|
| `rabbitmq`  | `event`, `command`, `query`, `notification` | [bancolombia/reactive-commons-go](https://github.com/bancolombia/reactive-commons-go) |
| `ibmmq`     | `listener`, `requestreply`                | [bancolombia/commons-mq-go](https://github.com/bancolombia/commons-mq-go) (CGO + IBM MQ C client) |
| `kafka`     | placeholder                               | not implemented yet — the CLI returns a friendly error                        |

### generate da

Group command for **driven adapters** (the components the application uses to
talk to the outside world). You must pick a kind:

```bash
scaffold-go generate da <kind> [name] [flags]

Kinds:
  database    Database adapter (postgres, mongodb, dynamodb)
  api         External HTTP API client
  publisher   Message queue publisher
  cache       Cache adapter (Redis, in-memory)
```

#### generate da database

```bash
scaffold-go generate da database [name] \
  --type postgres \
  --entity EntityName \
  --fields Field1:type1,Field2:type2

Flags:
  -t, --type string      Database type (postgres, mongodb, dynamodb)
  -e, --entity string    Entity name
  -f, --fields strings   Entity fields in format Name:type (comma-separated)
```

#### generate da api

```bash
scaffold-go generate da api [name] --url https://api.example.com

Flags:
  -u, --url string   Base URL for the external API
```

#### generate da publisher

Broker-aware message publisher. Just like `generate ep consumer`, this emits the
publisher file plus a per-component fx module and (on first use per broker) the
shared infrastructure module, and patches `cmd/main.go` automatically so the
new publisher is wired into `fx.New(...)`.

```bash
scaffold-go generate da publisher [name] \
  --broker rabbitmq|ibmmq|kafka \
  --pattern <pattern> \
  --target <event|command|resource|queue|app-name>

Flags:
  -b, --broker string    Messaging broker (required): rabbitmq | ibmmq | kafka
  -p, --pattern string   Messaging pattern (required); see table below
  -t, --target string    Destination name (event/command/resource/queue/app-name)
```

| Broker      | Patterns                                  | Library                                                                       |
|-------------|-------------------------------------------|-------------------------------------------------------------------------------|
| `rabbitmq`  | `event`, `command`, `query`, `notification` | [bancolombia/reactive-commons-go](https://github.com/bancolombia/reactive-commons-go) |
| `ibmmq`     | `sender`, `requestreply`                  | [bancolombia/commons-mq-go](https://github.com/bancolombia/commons-mq-go) (CGO + IBM MQ C client) |
| `kafka`     | placeholder                               | not implemented yet — the CLI returns a friendly error                        |

#### generate da cache

```bash
scaffold-go generate da cache [name]
```

## Project Structure

Generated projects follow this structure:

```
my-service/
├── cmd/
│   └── main.go              # Application entry point with Fx DI
├── config/
│   ├── config.go            # Configuration struct and Viper setup
│   └── config.yaml          # Default configuration
├── internal/
│   ├── domain/              # Business logic (entities, use cases, interfaces)
│   ├── entrypoints/         # HTTP handlers, message consumers
│   └── drivenadapters/      # Database adapters, API clients
├── go.mod
└── README.md
```

## Architecture Principles

### Domain Layer
- Contains core business logic
- No external dependencies
- Defines interfaces for repositories and services
- Framework-agnostic

### EntryPoints Layer
- Handles external requests (HTTP, messages)
- Validates input
- Calls domain use cases
- Formats responses

### DrivenAdapters Layer
- Implements domain interfaces
- Handles infrastructure concerns
- Manages external connections (databases, APIs, caches)

## Examples

### Complete workflow

The recommended order when adding a new feature is:
**1.** entity → **2.** usecase → **3.** entry-point (`ep`) → **4.** driven-adapter (`da`).

```bash
# 1. Create project
scaffold-go init blog-service --module github.com/user/blog-service
cd blog-service

# 2. Generate Post entity (domain)
scaffold-go gen entity Post --fields Title:string,Content:string,AuthorID:string

# 3. Generate CreatePost use case (domain)
scaffold-go gen usecase CreatePost --dependencies postRepo:PostRepository

# 4. Generate Post HTTP entry-point
scaffold-go gen ep http Post \
  --usecases createPost:CreatePostUseCase \
  --routes POST:/posts:CreatePost,GET:/posts/:id:GetPost

# 5. Generate PostgreSQL driven adapter
scaffold-go gen da database PostRepo \
  --type postgres \
  --entity Post \
  --fields Title:string,Content:string,AuthorID:string

# 6. Build and run
go mod tidy
go build -o blog-service cmd/main.go
./blog-service
```

### Using verbose logging for debugging

```bash
# Initialize with verbose output to see detailed progress
scaffold-go init my-service --module github.com/user/my-service -v

# Generate entity with debug information
scaffold-go gen entity User --fields Name:string,Email:string -v

# Output shows:
# - Validation steps
# - Template processing
# - File formatting operations
# - Detailed success metrics
```

## Development

### Build

```bash
go build -o scaffold-go cmd/main.go
```

### Test

```bash
go test ./...
```

### Run

```bash
go run cmd/main.go [command]
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License
