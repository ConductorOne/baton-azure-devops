package connector

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-azure-devops/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
)

type repositoryBuilder struct {
	resourceType *v2.ResourceType
	client       *client.AzureDevOpsClient
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
	var entitlements []*v2.Entitlement

	namespaces, err := o.client.ListSecurityNamespaces(ctx, []string{gitRepositoriesSecurityNamespace})
	if err != nil {
		return nil, "", nil, err
	}

	for _, namespace := range namespaces {
		readPermissionName := getPermissionName(resource.DisplayName, *namespace.Name, "read")
		readOptions := []entitlement.EntitlementOption{
			entitlement.WithGrantableTo(userResourceType, groupResourceType, teamResourceType),
			entitlement.WithDescription(fmt.Sprintf("Read permission in %s security namespace at %s repository level", *namespace.Name, resource.DisplayName)),
			entitlement.WithDisplayName(readPermissionName),
		}

		writePermissionName := getPermissionName(resource.DisplayName, *namespace.Name, "write")
		writeOptions := []entitlement.EntitlementOption{
			entitlement.WithGrantableTo(userResourceType, groupResourceType, teamResourceType),
			entitlement.WithDescription(fmt.Sprintf("Write permission in %s security namespace at %s repository level", *namespace.Name, resource.DisplayName)),
			entitlement.WithDisplayName(writePermissionName),
		}

		entitlements = append(entitlements, entitlement.NewPermissionEntitlement(resource, readPermissionName, readOptions...))
		entitlements = append(entitlements, entitlement.NewPermissionEntitlement(resource, writePermissionName, writeOptions...))
	}
	return entitlements, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *repositoryBuilder) Grants(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	var grants []*v2.Grant

	userMap, err := o.client.GetUsersMap(ctx)
	if err != nil {
		return nil, "", nil, err
	}

	namespaces, err := o.client.ListSecurityNamespaces(ctx, []string{gitRepositoriesSecurityNamespace})
	if err != nil {
		return nil, "", nil, err
	}

	for _, namespace := range namespaces {
		readPermissionBit := *namespace.ReadPermission
		writePermissionBit := *namespace.WritePermission
		bitMask := *namespace.SystemBitMask

		ACLs, err := o.client.ListAccessControlsBySecurityNamespace(ctx, *namespace.NamespaceId, parseGitTokenBySecurityNamespace(namespace.NamespaceId.String(), resource))
		if err != nil {
			return nil, "", nil, err
		}
		for _, acl := range ACLs {
			for _, ace := range *acl.AcesDictionary {
				grantResource, err := getIdentityResourceByDescriptor(o.client, ctx, *ace.Descriptor, userMap)
				if err != nil {
					continue
				}
				var basicGrantOptions []grant.GrantOption

				if o.client.SyncGrantSources {
					basicGrantOptions = append(basicGrantOptions, grant.WithAnnotation(&v2.GrantExpandable{
						EntitlementIds: []string{
							fmt.Sprintf("team:%s:member", grantResource.Id.Resource),
							fmt.Sprintf("group:%s:member", grantResource.Id.Resource),
							fmt.Sprintf("group:%s:admin", grantResource.Id.Resource),
						},
						Shallow: true,
					}))
				}
				// check allow for read and write using bit and bitmask properly
				if *ace.Allow&readPermissionBit != bitMask {
					// add grant for read
					grants = append(grants, grant.NewGrant(
						resource,
						getPermissionName(resource.DisplayName, *namespace.Name, "read"),
						grantResource,
						basicGrantOptions...,
					))
				}

				if *ace.Allow&writePermissionBit != bitMask {
					// add grant for write
					grants = append(grants, grant.NewGrant(
						resource,
						getPermissionName(resource.DisplayName, *namespace.Name, "write"),
						grantResource,
						basicGrantOptions...,
					))
				}
			}
		}
	}

	return grants, "", nil, nil
}

func parseGitTokenBySecurityNamespace(securityNamespace string, repositoryResource *v2.Resource) string {
	if securityNamespace == gitRepositoriesSecurityNamespace {
		return fmt.Sprintf("repoV2/%s/%s", repositoryResource.ParentResourceId.Resource, repositoryResource.Id.Resource)
	}
	return ""
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

func newRepositoryBuilder(c *client.AzureDevOpsClient) *repositoryBuilder {
	return &repositoryBuilder{
		resourceType: repositoryResourceType,
		client:       c,
	}
}
