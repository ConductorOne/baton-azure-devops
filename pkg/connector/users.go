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

type userBuilder struct {
	resourceType *v2.ResourceType
	client       *client.AzureDevOpsClient
}

func (o *userBuilder) ResourceType(_ context.Context) *v2.ResourceType {
	return userResourceType
}

// List returns all the users from the database as resource objects.
// Users include a UserTrait because they are the 'shape' of a standard user.
func (o *userBuilder) List(ctx context.Context, _ *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	var resources []*v2.Resource

	users, nextPageToken, err := o.client.ListUsers(ctx, pToken.Token)
	if err != nil {
		return nil, "", nil, err
	}

	for _, user := range users {
		userCopy := &user
		organizationResource, err := parseIntoUserResource(userCopy)
		if err != nil {
			return nil, "", nil, err
		}
		resources = append(resources, organizationResource)
	}

	return resources, nextPageToken, nil, nil
}

// Entitlements always returns an empty slice for users.
func (o *userBuilder) Entitlements(_ context.Context, _ *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *userBuilder) Grants(_ context.Context, _ *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func parseIntoUserResource(user *graph.GraphUser) (*v2.Resource, error) {
	var userStatus = v2.UserTrait_Status_STATUS_ENABLED

	profile := map[string]interface{}{
		"user_descriptor": *user.Descriptor,
		"username":        *user.DisplayName,
		"email":           *user.MailAddress,
	}
	userTraits := []resource.UserTraitOption{
		resource.WithUserProfile(profile),
		resource.WithStatus(userStatus),
		resource.WithEmail(*user.MailAddress, true),
		resource.WithUserLogin(*user.DisplayName),
	}

	userResource, err := resource.NewUserResource(
		*user.DisplayName,
		userResourceType,
		*user.Descriptor,
		userTraits,
	)
	if err != nil {
		return nil, err
	}

	return userResource, nil
}

func newUserBuilder(c *client.AzureDevOpsClient) *userBuilder {
	return &userBuilder{
		resourceType: userResourceType,
		client:       c,
	}
}
