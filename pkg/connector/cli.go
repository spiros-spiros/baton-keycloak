package connector

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/spiros-spiros/baton-keycloak/pkg/keycloak"
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
			logger := logging.NewLogger()

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

			builder, err := connectorbuilder.NewConnector(ctx, connector)
			if err != nil {
				logger.Error().Err(err).Msg("error creating connector builder")
				return err
			}

			err = builder.Run(ctx)
			if err != nil {
				logger.Error().Err(err).Msg("error running connector")
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
