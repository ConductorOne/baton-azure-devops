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

type teamBuilder struct {
	resourceType *v2.ResourceType
	client       *client.AzureDevOpsClient
}

func (o *teamBuilder) ResourceType(_ context.Context) *v2.ResourceType {
	return teamResourceType
}

func (o *teamBuilder) List(ctx context.Context, _ *v2.ResourceId, _ *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	var resources []*v2.Resource

	teams, err := o.client.ListTeams(ctx)
	if err != nil {
		return nil, "", nil, err
	}

	for _, team := range teams {
		teamCopy := &team
		teamResource, err := parseIntoTeamResource(teamCopy)
		if err != nil {
			return nil, "", nil, err
		}
		resources = append(resources, teamResource)
	}

	return resources, "", nil, nil
}

// Entitlements always returns an empty slice for users.
func (o *teamBuilder) Entitlements(_ context.Context, _ *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *teamBuilder) Grants(_ context.Context, _ *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func parseIntoTeamResource(team *core.WebApiTeam) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"team_id":      team.Id.String(),
		"display_name": *team.Name,
		"project_name": *team.ProjectName,
		"description":  *team.Description,
		"url":          *team.Url,
	}

	groupTraits := []resource.GroupTraitOption{
		resource.WithGroupProfile(profile),
	}
	parentResourceId, err := resource.NewResourceID(projectResourceType, *team.ProjectId)

	ret, err := resource.NewGroupResource(
		*team.Name,
		teamResourceType,
		team.Id.String(),
		groupTraits,
		resource.WithParentResourceID(parentResourceId),
	)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func newTeamBuilder(c *client.AzureDevOpsClient) *teamBuilder {
	return &teamBuilder{
		resourceType: teamResourceType,
		client:       c,
	}
}
