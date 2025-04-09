package connector

import (
	"context"

	"github.com/conductorone/baton-azure-devops/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/graph"
)

type groupBuilder struct {
	resourceType *v2.ResourceType
	client       *client.AzureDevOpsClient
}

func (o *groupBuilder) ResourceType(_ context.Context) *v2.ResourceType {
	return groupResourceType
}

func (o *groupBuilder) List(ctx context.Context, _ *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	var resources []*v2.Resource

	groups, nextPageToken, err := o.client.ListGroups(ctx, pToken.Token)
	if err != nil {
		return nil, "", nil, err
	}

	for _, group := range groups {
		groupCopy := &group
		groupResource, err := parseIntoGroupResource(groupCopy)
		if err != nil {
			return nil, "", nil, err
		}
		resources = append(resources, groupResource)
	}

	return resources, nextPageToken, nil, nil
}

// Entitlements always returns an empty slice for users.
func (o *groupBuilder) Entitlements(_ context.Context, _ *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *groupBuilder) Grants(_ context.Context, _ *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func parseIntoGroupResource(group *graph.GraphGroup) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"group_id":     *group.Domain,
		"display_name": *group.DisplayName,
		"description":  *group.Description,
		"url":          *group.Url,
	}

	groupTraits := []resource.GroupTraitOption{
		resource.WithGroupProfile(profile),
	}

	ret, err := resource.NewGroupResource(
		*group.DisplayName,
		groupResourceType,
		*group.Descriptor,
		groupTraits,
	)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func newGroupBuilder(c *client.AzureDevOpsClient) *groupBuilder {
	return &groupBuilder{
		resourceType: groupResourceType,
		client:       c,
	}
}
