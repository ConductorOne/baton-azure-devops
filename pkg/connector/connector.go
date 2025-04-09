package connector

import (
	"context"
	"io"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
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
		newSecurityNamespaceBuilder(d.client),
		newGroupBuilder(d.client),
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
	}, nil
}

// Validate is called to ensure that the connector is properly configured. It should exercise any API credentials
// to be sure that they are valid.
func (d *Connector) Validate(_ context.Context) (annotations.Annotations, error) {
	return nil, nil
}

// New returns a new instance of the connector.
func New(ctx context.Context, personalAccessToken, organizationUrl string, syncGrantSources bool, securityNamespaces []string) (*Connector, error) {
	l := ctxzap.Extract(ctx)

	azureDevOpsClient, err := client.New(ctx, personalAccessToken, organizationUrl, syncGrantSources, securityNamespaces)
	if err != nil {
		l.Error("error creating Azure DevOps client", zap.Error(err))
		return nil, err
	}

	return &Connector{
		client: azureDevOpsClient,
	}, nil
}
