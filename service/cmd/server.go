package cmd

import (
	"fmt"
	"os"

	"github.com/nickbryan/go-template/service/app"
	"github.com/nickbryan/go-template/service/infrastructure/postgres"
	"github.com/nickbryan/go-template/service/transport/rest"
	"github.com/nickbryan/go-template/service/transport/rest/customers"
	"github.com/nickbryan/go-template/service/transport/rest/health"
	"github.com/spf13/cobra"
)

const localEnv = "local"

//nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "server",
	Short: "Start the api server.",
	Long:  "Start the gotemplate api server for our example project.",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		defaultEnv, cleanup, er := app.NewDefaultEnvironment()
		if er != nil {
			return fmt.Errorf("unable to initialise default environment: %w", er)
		}
		defer func() {
			err = cleanup()
		}()

		// You can use the migration commands in the make file for local development.
		// Loading the migrations with live reloading would become cumbersome.
		if os.Getenv("APP_ENV") != localEnv {
			if err := app.Migrate(defaultEnv.Logger(), defaultEnv.Config().DatabaseURL); err != nil {
				return fmt.Errorf("migration failed: %w", err)
			}
		}

		s := rest.NewServer(defaultEnv)

		customerRepo := postgres.NewCustomerRepository(defaultEnv.DB())

		s.RegisterHandlers(
			health.NewCheckHandler(),
			customers.NewCreateHandler(customerRepo),
		)

		return s.Start()
	},
}
