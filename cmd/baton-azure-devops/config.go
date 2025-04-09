package main

import (
	"github.com/conductorone/baton-sdk/pkg/field"
	"github.com/spf13/viper"
)

var (
	bearerToken = field.StringField(
		"personal-access-token",
		field.WithDescription("The bearer token used to authenticate the request for Azure Dev Ops"),
		field.WithRequired(true),
	)
	organizations = field.StringField(
		"organization-url",
		field.WithDescription("The organization ids used to sync data for Azure Dev Ops"),
		field.WithRequired(true),
	)
	userSubjectTypes = field.StringSliceField(
		"user-subject-types",
		field.WithDescription("A comma separated list of user subject subtypes to reduce the retrieved results, e.g. msa’, ‘aad’, ‘svc’ (service identity), ‘imp’ (imported identity), etc."),
		field.WithRequired(false),
	)
	// ConfigurationFields defines the external configuration required for the
	// connector to run. Note: these fields can be marked as optional or
	// required.
	ConfigurationFields = []field.SchemaField{bearerToken, organizations, userSubjectTypes}

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
