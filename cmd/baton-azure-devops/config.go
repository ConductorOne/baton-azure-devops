package main

import (
	"github.com/conductorone/baton-sdk/pkg/field"
	"github.com/spf13/viper"
)

var (
	bearerTokenField = field.StringField(
		"personal-access-token",
		field.WithDescription("The bearer token used to authenticate the request for Azure Dev Ops"),
		field.WithRequired(true),
	)
	organizationUrlField = field.StringField(
		"organization-url",
		field.WithDescription("The organization ids used to sync data for Azure Dev Ops"),
		field.WithRequired(true),
	)
	syncGrantSourcesField = field.BoolField(
		"sync-grant-sources",
		field.WithDefaultValue(false),
		field.WithDescription("Sync grant sources. If this is not set, grant sources will not be synced."),
	)
	// ConfigurationFields defines the external configuration required for the
	// connector to run. Note: these fields can be marked as optional or
	// required.
	ConfigurationFields = []field.SchemaField{bearerTokenField, organizationUrlField, syncGrantSourcesField}

	// FieldRelationships defines relationships between the fields listed in
	// ConfigurationFields that can be automatically validated. For example, a
	// username and password can be required together, or an access token can be
	// marked as mutually exclusive from the username password pair.
	FieldRelationships = []field.SchemaFieldRelationship{}
)

// ValidateConfig is run after the configuration is loaded, and should return an
// error if it isn't valid. Implementing this function is optional, it only
// needs to perform extra validations that cannot be encoded with configuration
// parameters.
func ValidateConfig(v *viper.Viper) error {
	return nil
}
