package connector

import (
	"context"

	"github.com/conductorone/baton-azure-devops/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
)

type projectBuilder struct {
	resourceType *v2.ResourceType
	client       *client.AzureDevOpsClient
}

func (o *projectBuilder) ResourceType(_ context.Context) *v2.ResourceType {
	return projectResourceType
}

func (o *projectBuilder) List(ctx context.Context, _ *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	var resources []*v2.Resource

	projects, nextPageToken, err := o.client.ListProjects(ctx, pToken.Token)
	if err != nil {
		return nil, "", nil, err
	}

	for _, project := range projects {
		projectCopy := &project
		projectResource, err := parseIntoProjectResource(projectCopy)
		if err != nil {
			return nil, "", nil, err
		}
		resources = append(resources, projectResource)
	}

	return resources, nextPageToken, nil, nil
}

// Entitlements always returns an empty slice for users.
func (o *projectBuilder) Entitlements(_ context.Context, _ *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *projectBuilder) Grants(_ context.Context, _ *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func parseIntoProjectResource(project *core.TeamProjectReference) (*v2.Resource, error) {
	userResource, err := resource.NewResource(
		*project.Name,
		projectResourceType,
		project.Id.String(),
	)
	if err != nil {
		return nil, err
	}

	return userResource, nil
}

func newProjectBuilder(c *client.AzureDevOpsClient) *projectBuilder {
	return &projectBuilder{
		resourceType: projectResourceType,
		client:       c,
	}
}
