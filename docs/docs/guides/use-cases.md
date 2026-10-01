---
title: Generate use cases
description: Create application services, Fx providers, and main.go wiring.
---

# Generate use cases

A use case orchestrates domain ports. The implementation lives in
`internal/usecase` (not under `internal/domain`).

Source:
[internal/commands/generate_usecase.go](https://github.com/bancolombia/scaffold-go/blob/master/internal/commands/generate_usecase.go),
[internal/generator/usecase.go](https://github.com/bancolombia/scaffold-go/blob/master/internal/generator/usecase.go).

## Command

```bash
scaffold-go generate usecase [name] [--dependencies name:Type,...]
```

| Flag | Short | Required | Format |
|------|-------|----------|--------|
| `--dependencies` | `-d` | no | `camelName:PascalType` |

```bash
scaffold-go generate usecase CreateUser --dependencies userRepo:IUserRepository
scaffold-go gen usecase ProcessOrder -d orderRepo:IOrderRepository,payment:PaymentService
scaffold-go generate usecase ValidateEmail
```

Dependencies are imported as `domain.{Type}`. Use the interface names from
entity generation (`IUserRepository`), not adapter structs.

## Files written

For `CreateUser`:

| File | Role |
|------|------|
| `internal/usecase/createuser_usecase.go` | Interface + impl |
| `internal/usecase/createuser_usecase_test.go` | Tests |
| `internal/usecase/module.go` | Created or updated with `NewCreateUserUseCase` |

If `module.go` already exists, the CLI **adds** `New{Name}UseCase` to
`fx.Provide` via
[GenericModuleManager](https://github.com/bancolombia/scaffold-go/blob/master/internal/manager/generic_module.go).

## Generated shape

From
[usecase.go.tmpl](https://github.com/bancolombia/scaffold-go/blob/master/internal/template/templates/domain/usecase/usecase.go.tmpl):

```go
package usecase

import (
    "context"
    "github.com/user/blog-service/internal/domain"
    "go.uber.org/zap"
)

type CreateUserUseCase interface {
    Execute(ctx context.Context, input any) (any, error)
}

type createuserImpl struct {
    userRepo domain.IUserRepository
    logger   *zap.Logger
}

func NewCreateUserUseCase(userRepo domain.IUserRepository, logger *zap.Logger) CreateUserUseCase {
    return &createuserImpl{userRepo: userRepo, logger: logger}
}

func (uc *createuserImpl) Execute(ctx context.Context, input any) (any, error) {
    uc.logger.Info("executing CreateUser")
    // TODO: Implement use case logic
    var result any
    return result, nil
}
```

`InputType` and `OutputType` are currently always `any` in the generator.
Replace them with real DTOs or domain types.

`extractEntityType` maps `IUserRepository` / `UserRepository` → `User` for
template metadata; you still write the `Execute` body.

## main.go wiring

After writing files, the command:

1. Parses the module path from `go.mod`
2. `AddImport(modulePath + "/internal/usecase")`
3. `AddFxModule("usecase.Module")`

Source: [MainFileManager](https://github.com/bancolombia/scaffold-go/blob/master/internal/manager/main_file.go).

Fx still needs a provider for `domain.IUserRepository`. That comes from a
[database adapter](./database-adapters.md) (or a fake in tests). Until then,
`fx.New` may fail at startup with a missing dependency.

## What to do next

1. Change `Execute` signature if you do not want `any`.
2. Call repository methods; map errors to domain errors.
3. Generate an [HTTP handler](./http-entry-points.md) with
   `--usecases createUser:CreateUserUseCase`.
4. Keep use cases free of Fiber/`sql` imports.

## Related pages

| Page | Relationship |
|------|-------------|
| [Entities](./entities.md) | Ports the use case depends on |
| [HTTP entry-points](./http-entry-points.md) | Injects `CreateUserUseCase` |
| [Dependency injection](./dependency-injection.md) | `usecase.Module` in `fx.New` |
