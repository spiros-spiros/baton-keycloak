package connector

import (
	"context"
	"io"

	"github.com/spiros-spiros/baton-keycloak/pkg/keycloak"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
)

type Connector struct {
	client *keycloak.Client
}

// ResourceSyncers returns a ResourceSyncer for each resource type that should be synced from the upstream service.
func (c *Connector) ResourceSyncers(ctx context.Context) []connectorbuilder.ResourceSyncer {
	return []connectorbuilder.ResourceSyncer{
		newUserSyncer(c.client),
		newGroupSyncer(c.client),
	}
}

// Asset takes an input AssetRef and attempts to fetch it using the connector's authenticated http client
// crossing my fingers that this is not needed tbh.
func (c *Connector) Asset(ctx context.Context, asset *v2.AssetRef) (string, io.ReadCloser, error) {
	return "", nil, nil
}

// Metadata returns metadata about the connector for C1 in the logs and whatnot. It will also display in the UI. Sadly emojis are not supported.
func (c *Connector) Metadata(ctx context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "Keycloak",
		Description: "Connector syncing users and groups from Keycloak",
	}, nil
}

// Validate is called to ensure that the connector is properly configured. It should test API credentials
func (c *Connector) Validate(ctx context.Context) (annotations.Annotations, error) {
	return nil, nil
}

func (c *Connector) Close() error {
	return c.client.Close()
}

// Actually create a Keycloak connector.
func New(ctx context.Context, keycloakServerURL string, keycloakRealm string, keycloakClientID string, keycloakClientSecret string) (*Connector, error) {
	l := ctxzap.Extract(ctx)
	client, err := client.New(ctx, client.NewClient(ctx, keycloakServerURL, keycloakRealm, keycloakClientID, keycloakClientSecret))
	if err != nil {
		l.Error("error creating Keycloak client for some reason", zap.Error(err))
		return nil, err
	}

	return &Connector{
		client: client,
	}, nil
}