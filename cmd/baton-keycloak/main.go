package main

// this connector should allow Conductor One to sync users and groups from Keycloak
// it should also allow for entitlement provisioning for users and groups
// this is mainly so that when someone at Weaviate tries to access a cluster, C1 can add them to the right group on a JIT basis.
import (
	"context"
	"fmt"
	"os"

	"github.com/spiros-spiros/baton-keycloak/pkg/connector"
	"github.com/spiros-spiros/baton-keycloak/pkg/keycloak"
	"github.com/conductorone/baton-sdk/pkg/cli"
)

func main() {
	ctx := context.Background()
	// these need to be secrets spiros - don't forget to make the secrets
	cfg := &config{
		ServerURL:    os.Getenv("KEYCLOAK_SERVER_URL"),
		Realm:        os.Getenv("KEYCLOAK_REALM"),
		ClientID:     os.Getenv("KEYCLOAK_CLIENT_ID"),
		ClientSecret: os.Getenv("KEYCLOAK_CLIENT_SECRET"),
	}

	if err := run(ctx, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

type config struct {
	ServerURL    string
	Realm        string
	ClientID     string
	ClientSecret string
}

func run(ctx context.Context, cfg *config) error {
	client := keycloak.NewClient(cfg.ServerURL, cfg.Realm, cfg.ClientID, cfg.ClientSecret)
	if err := client.Connect(ctx); err != nil {
		return err
	}

	c := connector.New(client)
	return cli.Start(ctx, c)
}

