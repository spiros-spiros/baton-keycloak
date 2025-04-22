package connector

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/connectorrunner"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/spiros-spiros/baton-keycloak/pkg/keycloak"
	"go.uber.org/zap"
)

var configRequiredFlags = []string{"keycloak-server-url", "keycloak-realm", "keycloak-client-id", "keycloak-client-secret"}

type Config struct {
	ServerURL    string `mapstructure:"keycloak-server-url"`
	Realm        string `mapstructure:"keycloak-realm"`
	ClientID     string `mapstructure:"keycloak-client-id"`
	ClientSecret string `mapstructure:"keycloak-client-secret"`
}

func (c *Config) Validate() error {
	for _, f := range configRequiredFlags {
		if viper.GetString(f) == "" {
			return fmt.Errorf("required flag %s is not set", f)
		}
	}
	return nil
}

func (c *Config) Load() error {
	err := viper.Unmarshal(c)
	if err != nil {
		return err
	}
	return nil
}

func NewConnector(ctx context.Context, cfg *Config) (*Connector, error) {
	client := keycloak.NewClient(cfg.ServerURL, cfg.Realm, cfg.ClientID, cfg.ClientSecret)
	if err := client.Connect(ctx); err != nil {
		return nil, err
	}

	return New(client), nil
}

func RegisterCmd(parent *cobra.Command) {
	config := &Config{}
	cmd := &cobra.Command{
		Use:   "keycloak",
		Short: "Keycloak connector",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			logger := ctxzap.Extract(ctx)

			err := config.Load()
			if err != nil {
				return err
			}

			err = config.Validate()
			if err != nil {
				return err
			}

			connector, err := NewConnector(ctx, config)
			if err != nil {
				return err
			}

			server, err := connectorbuilder.NewConnector(ctx, connector)
			if err != nil {
				logger.Error("error creating connector", zap.Error(err))
				return err
			}

			runner, err := connectorrunner.NewConnectorRunner(ctx, server)
			if err != nil {
				logger.Error("error creating connector runner", zap.Error(err))
				return err
			}

			if err := runner.Run(ctx); err != nil {
				logger.Error("error running connector", zap.Error(err))
				return err
			}

			return nil
		},
	}

	cmd.Flags().String("keycloak-server-url", "", "The URL of the Keycloak server")
	cmd.Flags().String("keycloak-realm", "", "The Keycloak realm to connect to")
	cmd.Flags().String("keycloak-client-id", "", "The Keycloak client ID")
	cmd.Flags().String("keycloak-client-secret", "", "The Keycloak client secret")

	for _, f := range configRequiredFlags {
		err := viper.BindPFlag(f, cmd.Flags().Lookup(f))
		if err != nil {
			panic(err)
		}
	}

	parent.AddCommand(cmd)
}
