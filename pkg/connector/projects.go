package connector

import (
	"context"
	"fmt"
	"strings"

	"github.com/conductorone/baton-azure-devops/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
)

const (
	projectSecurityNamespace             = "52d39943-cb85-4d7f-8fa8-c6baac873819"
	taggingSecurityNamespace             = "bb50f182-8e5e-40b8-bc21-e8752a1e7ae2"
	versionControlItemsSecurityNamespace = "a39371cf-0841-4c16-bbd3-276e341bc052"
	analyticsViewsSecurityNamespace      = "d34d3680-dfe5-4cc6-a949-7d9c68f73cba"
	buildSecurityNamespace               = "33344d9c-fc72-4d6f-aba5-fa317101a7e9"
	gitRepositoriesSecurityNamespace     = "2e9eb7ed-3c0a-47d4-87c1-0ffdd275fd87"
	metaTaskSecurityNamespace            = "f6a4de49-dbe2-4704-86dc-f8ec1a294436"
	releaseManagementSecurityNamespace   = "c788c23e-1b46-4162-8f5e-d7585343b5de"
)

var (
	securityNamespaces = []string{
		projectSecurityNamespace,
		taggingSecurityNamespace,
		versionControlItemsSecurityNamespace,
		analyticsViewsSecurityNamespace,
		buildSecurityNamespace,
		gitRepositoriesSecurityNamespace,
		metaTaskSecurityNamespace,
		releaseManagementSecurityNamespace,
	}
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

func (o *projectBuilder) Entitlements(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var entitlements []*v2.Entitlement

	namespaces, err := o.client.ListSecurityNamespaces(ctx, securityNamespaces)
	if err != nil {
		return nil, "", nil, err
	}

	for _, namespace := range namespaces {
		readPermissionName := getPermissionName(resource.DisplayName, *namespace.Name, "read")
		readOptions := []entitlement.EntitlementOption{
			entitlement.WithGrantableTo(userResourceType, groupResourceType, teamResourceType),
			entitlement.WithDescription(fmt.Sprintf("Read permission in %s security namespace at %s project level", *namespace.Name, resource.DisplayName)),
			entitlement.WithDisplayName(readPermissionName),
		}

		writePermissionName := getPermissionName(resource.DisplayName, *namespace.Name, "write")
		writeOptions := []entitlement.EntitlementOption{
			entitlement.WithGrantableTo(userResourceType, groupResourceType, teamResourceType),
			entitlement.WithDescription(fmt.Sprintf("Write permission in %s security namespace at %s project level", *namespace.Name, resource.DisplayName)),
			entitlement.WithDisplayName(writePermissionName),
		}

		entitlements = append(entitlements, entitlement.NewPermissionEntitlement(resource, readPermissionName, readOptions...))
		entitlements = append(entitlements, entitlement.NewPermissionEntitlement(resource, writePermissionName, writeOptions...))
	}
	return entitlements, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *projectBuilder) Grants(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	var grants []*v2.Grant

	userMap, err := o.client.GetUsersMap(ctx)
	if err != nil {
		return nil, "", nil, err
	}

	namespaces, err := o.client.ListSecurityNamespaces(ctx, securityNamespaces)
	if err != nil {
		return nil, "", nil, err
	}

	for _, namespace := range namespaces {
		readPermissionBit := *namespace.ReadPermission
		writePermissionBit := *namespace.WritePermission
		bitMask := *namespace.SystemBitMask

		ACLs, err := o.client.ListAccessControlsBySecurityNamespace(ctx, *namespace.NamespaceId, parseTokenBySecurityNamespace(namespace.NamespaceId.String(), resource))
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
				effectiveAllow := *ace.Allow
				if ace.ExtendedInfo.EffectiveAllow != nil {
					effectiveAllow = *ace.ExtendedInfo.EffectiveAllow
				}
				if effectiveAllow&readPermissionBit != bitMask {
					// add grant for read
					grants = append(grants, grant.NewGrant(
						resource,
						getPermissionName(resource.DisplayName, *namespace.Name, "read"),
						grantResource,
						basicGrantOptions...,
					))
				}

				if effectiveAllow&writePermissionBit != bitMask {
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

func parseTokenBySecurityNamespace(securityNamespace string, projectResource *v2.Resource) string {
	switch securityNamespace {
	case projectSecurityNamespace:
		return fmt.Sprintf("$PROJECT:vstfs:///Classification/TeamProject/%s", projectResource.Id.Resource)
	case taggingSecurityNamespace:
		return fmt.Sprintf("/%s", projectResource.Id.Resource)
	case versionControlItemsSecurityNamespace:
		return fmt.Sprintf("$/%s", projectResource.DisplayName)
	case analyticsViewsSecurityNamespace:
		return fmt.Sprintf("$/Shared/%s", projectResource.Id.Resource)
	case buildSecurityNamespace:
		return projectResource.Id.Resource
	case gitRepositoriesSecurityNamespace:
		return fmt.Sprintf("repoV2/%s", projectResource.Id.Resource)
	case metaTaskSecurityNamespace:
		return projectResource.Id.Resource
	case releaseManagementSecurityNamespace:
		return projectResource.Id.Resource
	}
	return ""
}

func getPermissionName(projectName, namespace, action string) string {
	return strings.Join([]string{projectName, namespace, action}, "_")
}

func parseIntoProjectResource(project *core.TeamProjectReference) (*v2.Resource, error) {
	userResource, err := resource.NewResource(
		*project.Name,
		projectResourceType,
		project.Id.String(),
		resource.WithAnnotation(
			&v2.ChildResourceType{ResourceTypeId: repositoryResourceType.Id},
		),
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
