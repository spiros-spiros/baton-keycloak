package main

// this connector should allow Conductor One to sync users and groups from Keycloak
// it should also allow for entitlement provisioning for users and groups
// this is mainly so that when someone at Weaviate tries to access a cluster, C1 can add them to the right group on a JIT basis.
import (
	"context"
	"os"

	"github.com/spiros-spiros/baton-keycloak/pkg/connector"
	"github.com/conductorone/baton-sdk/pkg/logging"
	"github.com/spf13/cobra"
)

func main() {
	ctx := context.Background()

	cmd := &cobra.Command{
		Use:   "baton-keycloak",
		Short: "Baton Keycloak connector",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			logger := logging.FromContext(ctx)

			connector.RegisterCmd(cmd)

			err := cmd.Execute()
			if err != nil {
				logger.Error().Err(err).Msg("error running connector")
				return err
			}

			return nil
		},
	}

	cmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug logging")
	cmd.PersistentFlags().StringP("output", "o", "", "Output file path")
	cmd.PersistentFlags().StringP("format", "f", "json", "Output format (json, yaml)")

	err := cmd.ExecuteContext(ctx)
	if err != nil {
		os.Exit(1)
	}
}
