package connector

import (
	"context"
	"io"

	"github.com/conductorone/baton-azure-devops/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type Connector struct {
	client *client.AzureDevOpsClient
}

// ResourceSyncers returns a ResourceSyncer for each resource type that should be synced from the upstream service.
func (d *Connector) ResourceSyncers(_ context.Context) []connectorbuilder.ResourceSyncer {
	return []connectorbuilder.ResourceSyncer{
		newUserBuilder(d.client),
		newProjectBuilder(d.client),
		newTeamBuilder(d.client),
		newGroupBuilder(d.client),
		newRepositoryBuilder(d.client),
	}
}

// Asset takes an input AssetRef and attempts to fetch it using the connector's authenticated http client
// It streams a response, always starting with a metadata object, following by chunked payloads for the asset.
func (d *Connector) Asset(_ context.Context, _ *v2.AssetRef) (string, io.ReadCloser, error) {
	return "", nil, nil
}

// Metadata returns metadata about the connector.
func (d *Connector) Metadata(_ context.Context) (*v2.ConnectorMetadata, error) {
	return &v2.ConnectorMetadata{
		DisplayName: "Azure Dev Ops Connector",
		Description: "Connector to sync users, security namespaces, projects, teams and groups",
		AccountCreationSchema: &v2.ConnectorAccountCreationSchema{
			FieldMap: map[string]*v2.ConnectorAccountCreationSchema_Field{
				"principal_name": {
					DisplayName: "Principal Name",
					Required:    true,
					Description: "The Entra ID principal name of the user (e.g., their email).",
					Field: &v2.ConnectorAccountCreationSchema_Field_StringField{
						StringField: &v2.ConnectorAccountCreationSchema_StringField{},
					},
					Placeholder: "user@example.com",
					Order:       1,
				},
				"license_type": {
					DisplayName: "License Type",
					Required:    true,
					Description: "The type of license to assign to the user. Must be one of: express, stakeholder, Visual Studio Subscriber.",
					Field: &v2.ConnectorAccountCreationSchema_Field_StringField{
						StringField: &v2.ConnectorAccountCreationSchema_StringField{},
					},
					Placeholder: "express",
					Order:       2,
				},
			},
		},
	}, nil
}

// Validate is called to ensure that the connector is properly configured. It should exercise any API credentials
// to be sure that they are valid.
func (d *Connector) Validate(_ context.Context) (annotations.Annotations, error) {
	return nil, nil
}

// New returns a new instance of the connector.
func New(ctx context.Context, personalAccessToken, organizationUrl string, syncGrantSources bool) (*Connector, error) {
	l := ctxzap.Extract(ctx)

	azureDevOpsClient, err := client.New(ctx, personalAccessToken, organizationUrl, syncGrantSources)
	if err != nil {
		l.Error("error creating Azure DevOps client", zap.Error(err))
		return nil, err
	}

	return &Connector{
		client: azureDevOpsClient,
	}, nil
}
