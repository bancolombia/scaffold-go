package commands

import (
	"github.com/gabheadz/scaffold-go/internal/validator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	cfgFile           string
	verbose           bool
	logger            *zap.Logger
	validatorInstance *validator.Validator
)

func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "scaffold-go",
		Short: "A Go CLI tool for scaffolding microservice projects",
		Long: `scaffold-go is a code generation tool that scaffolds Go microservice projects 
following clean architecture principles.

The tool generates projects with three distinct layers:
  • Domain Layer: Pure business logic with no external dependencies
  • EntryPoints Layer: External interfaces (HTTP, messaging)
  • DrivenAdapters Layer: Infrastructure implementations (databases, APIs, caching)

The dependency rule is strictly enforced: dependencies point inward toward the domain.`,
		Example: `  # Initialize a new project
  scaffold-go init my-service --module github.com/user/my-service

  # Generate a domain entity
  scaffold-go generate entity User --fields Name:string,Email:string

  # Generate an HTTP handler entry-point
  scaffold-go generate ep http UserHandler --usecases createUser:CreateUserUseCase

  # Generate a database driven adapter
  scaffold-go generate da database UserRepo --type postgres --entity User --fields Name:string,Email:string

  # Get help for any command
  scaffold-go help generate
  scaffold-go generate entity --help`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initLogger()
		},
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path (default: $HOME/.scaffold-go.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging with detailed debug information")

	rootCmd.AddCommand(NewInitCommand())
	rootCmd.AddCommand(NewGenerateCommand())
	rootCmd.AddCommand(NewVersionCommand())

	cobra.OnInitialize(initConfig)

	return rootCmd
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath("$HOME")
		viper.SetConfigName(".scaffold-go")
	}

	viper.AutomaticEnv()
	viper.ReadInConfig()
}

func initLogger() {
	var err error
	var config zap.Config

	if verbose {
		// Development mode with detailed output
		config = zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		// Production mode with structured JSON output
		config = zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	logger, err = config.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel),
	)
	if err != nil {
		panic(err)
	}

	// Add global context fields
	logger = logger.With(
		zap.String("tool", "scaffold-go"),
		zap.String("version", Version),
	)
}

func GetLogger() *zap.Logger {
	if logger == nil {
		initLogger()
	}
	return logger
}

func GetValidator() *validator.Validator {
	if validatorInstance == nil {
		validatorInstance = validator.New()
	}
	return validatorInstance
}
