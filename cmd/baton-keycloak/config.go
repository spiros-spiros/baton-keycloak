package main

import (
	"github.com/conductorone/baton-sdk/pkg/field"
	"github.com/spf13/viper"
)

var (
	keycloakServerURLField = field.StringField(
		"keycloak-server-url",
		field.WithDescription("The URL of the Keycloak server."),
		field.WithDefaultValue("https://keycloak.com/"),
		field.WithRequired(true),
	)
	keycloakRealmField = field.StringField(
		"keycloak-realm",
		field.WithDescription("The realm of the Keycloak server."),
		field.WithRequired(true),
	)
	keycloakClientIDField = field.StringField(
		"keycloak-client-id",
		field.WithDescription("The client ID you made."),
		field.WithRequired(true),
	)
	keycloakClientSecretField = field.StringField(
		"keycloak-client-secret",
		field.WithDescription("The client secret for the client you made."),
		field.WithRequired(true),
	)

	ConfigurationFields = []field.SchemaField{keycloakServerURLField, keycloakRealmField, keycloakClientIDField, keycloakClientSecretField}

	FieldRelationships = []field.SchemaFieldRelationship{}
)

// ValidateConfig is run after the configuration is loaded, and should return an
// error if it isn't valid. Implementing this function is optional, it only
// needs to perform extra validations that cannot be encoded with configuration
// parameters.
func ValidateConfig(v *viper.Viper) error {
	return nil
}