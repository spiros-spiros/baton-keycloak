package main

// this connector should allow Conductor One to sync users and groups from Keycloak
// it should also allow for entitlement provisioning for users and groups
// this is mainly so that when someone at Weaviate tries to access a cluster, C1 can add them to the right group on a JIT basis.
import (
	"context"
	"fmt"
	"os"

	"github.com/conductorone/baton-sdk/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/field"
	"github.com/conductorone/baton-sdk/pkg/types"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/spf13/viper"
	connectorSchema "github.com/spiros-spiros/baton-keycloak/pkg/connector"
	"go.uber.org/zap"
)

var (
	apiUrlField               = field.StringField("api_url", field.WithDescription("The URL of the Keycloak server"), field.WithRequired(true))
	realmField                = field.StringField("realm", field.WithDescription("The realm to connect to"), field.WithRequired(true))
	keycloakclientField       = field.StringField("keycloak_client_id", field.WithDescription("The client ID to use for authentication"), field.WithRequired(true))
	keycloakclientSecretField = field.StringField("keycloak_client_secret", field.WithDescription("The client secret to use for authentication"), field.WithRequired(true))
	batonClientIDField        = field.StringField("baton_client_id", field.WithDescription("The Baton client ID"), field.WithRequired(true))
	batonClientSecretField    = field.StringField("baton_client_secret", field.WithDescription("The Baton client secret"), field.WithRequired(true))
)

var configuration = field.NewConfiguration([]field.SchemaField{
	apiUrlField,
	realmField,
	keycloakclientField,
	keycloakclientSecretField,
	batonClientIDField,
	batonClientSecretField,
})

var version = "dev"

func main() {
	ctx := context.Background()

	_, cmd, err := config.DefineConfiguration(
		ctx,
		"baton-keycloak",
		getConnector,
		configuration,
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	cmd.Version = version

	err = cmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func getConnector(ctx context.Context, v *viper.Viper) (types.ConnectorServer, error) {
	l := ctxzap.Extract(ctx)
	if err := ValidateConfig(v); err != nil {
		return nil, err
	}

	keycloakServerURL := v.GetString(apiUrlField.FieldName)
	keycloakRealm := v.GetString(realmField.FieldName)
	keycloakClientID := v.GetString(keycloakclientField.FieldName)
	keycloakClientSecret := v.GetString(keycloakclientSecretField.FieldName)

	cb, err := connectorSchema.New(ctx, keycloakServerURL, keycloakRealm, keycloakClientID, keycloakClientSecret)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}
	connector, err := connectorbuilder.NewConnector(ctx, cb)
	if err != nil {
		l.Error("error creating connector", zap.Error(err))
		return nil, err
	}
	return connector, nil
}
