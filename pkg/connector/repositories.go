package connector

import (
	"context"

	"github.com/conductorone/baton-azure-devops/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
)

type repositoryBuilder struct {
	resourceType *v2.ResourceType
	client       *client.AzureDevOpsClient
	connector    *Connector
}

func (o *repositoryBuilder) ResourceType(_ context.Context) *v2.ResourceType {
	return repositoryResourceType
}

func (o *repositoryBuilder) List(ctx context.Context, parent *v2.ResourceId, _ *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	var resources []*v2.Resource

	if parent != nil {
		repositories, err := o.client.ListRepositories(ctx, parent.Resource)
		if err != nil {
			return nil, "", nil, err
		}

		for _, repository := range repositories {
			repositoryCopy := &repository
			repositoryResource, err := parseIntoRepositoryResource(repositoryCopy)
			if err != nil {
				return nil, "", nil, err
			}
			resources = append(resources, repositoryResource)
		}
	}

	return resources, "", nil, nil
}

func (o *repositoryBuilder) Entitlements(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	namespaces, err := o.client.ListSecurityNamespaces(ctx, []string{gitRepositoriesSecurityNamespace})
	if err != nil {
		return nil, "", nil, err
	}

	return getEntitlementsFromSecurityNamespaces(namespaces, resource), "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *repositoryBuilder) Grants(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	err := o.connector.loadUsers(ctx)
	if err != nil {
		return nil, "", nil, err
	}

	namespaces, err := o.client.ListSecurityNamespaces(ctx, []string{gitRepositoriesSecurityNamespace})
	if err != nil {
		return nil, "", nil, err
	}

	grants, err := getGrantsFromSecurityNamespaces(ctx, o.client, o.connector.users, namespaces, resource)
	if err != nil {
		return nil, "", nil, err
	}

	return grants, "", nil, nil
}

func parseIntoRepositoryResource(repository *git.GitRepository) (*v2.Resource, error) {
	userResource, err := resource.NewResource(
		*repository.Name,
		repositoryResourceType,
		repository.Id.String(),
		resource.WithParentResourceID(
			&v2.ResourceId{
				ResourceType: projectResourceType.Id,
				Resource:     repository.Project.Id.String(),
			}),
	)
	if err != nil {
		return nil, err
	}

	return userResource, nil
}

func newRepositoryBuilder(c *client.AzureDevOpsClient, d *Connector) *repositoryBuilder {
	return &repositoryBuilder{
		resourceType: repositoryResourceType,
		client:       c,
		connector:    d,
	}
}
