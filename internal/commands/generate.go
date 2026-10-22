package commands

import (
	"github.com/spf13/cobra"
)

func NewGenerateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate components for your microservice",
		Long: `Generate components for your microservice following clean architecture.

The recommended order when building a feature is:
  1. entity    - Define the domain model (and its repository interface)
  2. interface - Optionally declare another domain port that operates on an entity
  3. usecase   - Implement the business logic that orchestrates entities
  4. ep        - Add an entry-point that drives the use case (HTTP, message consumer, ...)
  5. da        - Add a driven adapter that implements outbound contracts (database, api, publisher, cache)

Each entry-point and driven adapter requires you to specify its kind explicitly,
so there is no ambiguity between, for example, an HTTP handler and a message consumer.

All generate commands validate that you're in an initialized project (checks for go.mod).`,
		Aliases: []string{"gen", "g"},
		Example: `  # 1. Domain entity
  scaffold-go generate entity User --fields Name:string,Email:string

  # 2. Domain interface (optional port)
  scaffold-go generate interface ProductReceiver --entity Product --operation ProcessProduct

  # 3. Domain use case
  scaffold-go generate usecase CreateUser --dependencies userRepo:UserRepository

  # 4. Entry-point (HTTP handler)
  scaffold-go generate ep http UserHandler \
    --usecases createUser:CreateUserUseCase \
    --routes POST:/users:CreateUser

  # 4. Entry-point (message consumer)
  scaffold-go generate ep consumer OrderConsumer --usecases processOrder:ProcessOrderUseCase

  # 5. Driven adapter (database)
  scaffold-go generate da database UserRepo --type postgres --entity User --fields Name:string,Email:string

  # 5. Driven adapter (external API)
  scaffold-go generate da api PaymentClient --url https://api.payment.com

  # Use short alias
  scaffold-go gen entity Product --fields Name:string,Price:float64`,
	}

	// Define common flags that are inherited by all subcommands
	cmd.PersistentFlags().StringP("output-path", "o", ".", "Output directory path for generated files (relative to project root)")
	cmd.PersistentFlags().Bool("force", false, "Force overwrite existing files without prompting")

	cmd.AddCommand(NewGenerateEntityCommand())
	cmd.AddCommand(NewGenerateInterfaceCommand())
	cmd.AddCommand(NewGenerateUseCaseCommand())
	cmd.AddCommand(NewGenerateEntryPointCommand())
	cmd.AddCommand(NewGenerateDrivenAdapterCommand())

	return cmd
}
