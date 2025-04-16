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
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/security"
)

type securityNamespaceBuilder struct {
	resourceType *v2.ResourceType
	client       *client.AzureDevOpsClient
}

func (o *securityNamespaceBuilder) ResourceType(_ context.Context) *v2.ResourceType {
	return securityNamespaceResourceType
}

func (o *securityNamespaceBuilder) List(ctx context.Context, _ *v2.ResourceId, _ *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	var resources []*v2.Resource

	securityNamespaces, err := o.client.ListSecurityNamespaces(ctx)
	if err != nil {
		return nil, "", nil, err
	}

	for _, securityNamespace := range securityNamespaces {
		securityNamespaceCopy := &securityNamespace
		securityNamespaceResource, err := parseIntoSecurityNamespaceResource(securityNamespaceCopy)
		if err != nil {
			return nil, "", nil, err
		}
		resources = append(resources, securityNamespaceResource)
	}

	return resources, "", nil, nil
}

func (o *securityNamespaceBuilder) Entitlements(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)
	var entitlements []*v2.Entitlement

	namespaceUUID, err := uuid.Parse(resource.Id.Resource)
	if err != nil {
		return nil, "", nil, err
	}

	// get all projects
	projectsMap := make(map[string]string)
	nextPageToken := ""
	for {
		projects, nextPageToken, err := o.client.ListProjects(ctx, nextPageToken)
		if err != nil {
			l.Error(fmt.Sprintf("Error getting list of projects %v", err))
		}

		for _, project := range projects {
			projectsMap[project.Id.String()] = *project.Name
		}

		if nextPageToken == "" {
			break
		}
	}

	actions, err := o.client.ListActionsBySecurityNamespace(ctx, namespaceUUID)
	if err != nil {
		return nil, "", nil, err
	}

	for _, action := range actions {
		if action.Name != nil || action.DisplayName != nil {
			var actionName string
			var actionDisplayName string
			if action.Name != nil {
				actionName = *action.Name
				actionDisplayName = *action.Name
			}
			if action.DisplayName != nil {
				actionDisplayName = *action.DisplayName
			}

			entitlements = append(entitlements, parseIntoEntitlements(resource, actionDisplayName, actionName, "")...)

			for _, projectName := range projectsMap {
				entitlements = append(entitlements, parseIntoEntitlements(resource, actionDisplayName, actionName, fmt.Sprintf("[%s]", projectName))...)
			}
		}
	}

	return entitlements, "", nil, nil
}

func (o *securityNamespaceBuilder) Grants(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)

	var grants []*v2.Grant

	namespaceUUID, err := uuid.Parse(resource.Id.Resource)
	if err != nil {
		return nil, "", nil, err
	}

	userMap, err := o.client.GetUsersMap(ctx)
	if err != nil {
		return nil, "", nil, err
	}
	actions, err := o.client.ListActionsBySecurityNamespace(ctx, namespaceUUID)
	if err != nil {
		return nil, "", nil, err
	}
	acls, err := o.client.ListAccessControlsBySecurityNamespace(ctx, namespaceUUID)
	if err != nil {
		return nil, "", nil, err
	}

	for _, action := range actions {
		if action.Bit == nil || action.Name == nil {
			continue
		}
		for _, acl := range acls {
			for _, value := range *acl.AcesDictionary {
				// get user email to map with user
				parts := strings.Split(*value.Descriptor, `\`)

				if len(parts) == 2 {
					userPrincipalName := parts[1]
					if userMap[userPrincipalName] != "" {
						userResource := &v2.Resource{
							Id: &v2.ResourceId{
								ResourceType: userResourceType.Id,
								Resource:     userMap[userPrincipalName],
							},
						}
						grants = append(grants, parseIntoGrants(o.client.SyncGrantSources, userResource, resource, value, action)...)
					}
				} else {
					teamsMap, err := o.client.ListTeamIDs(ctx)
					if err != nil {
						l.Error(fmt.Sprintf("Failed to list teams %v", err))
						continue
					}
					groupIdentities, err := o.client.ListIdentities(ctx, "", *value.Descriptor)
					if err != nil {
						l.Error(fmt.Sprintf("Failed to list identities %v", err))
						continue
					}
					if len(groupIdentities) > 0 {
						groupIdentity := groupIdentities[0]
						if groupIdentity.Id != nil {
							var parentResource *v2.ResourceId

							// get parent name
							parts := strings.Split(*groupIdentity.ProviderDisplayName, `\`)
							if len(parts) == 2 {
								providerDisplayName := parts[0]
								parentResource = &v2.ResourceId{
									ResourceType: projectResourceType.Id,
									Resource:     providerDisplayName,
								}
							}

							if _, isTeam := teamsMap[groupIdentity.Id.String()]; isTeam {
								teamResource := &v2.Resource{
									Id: &v2.ResourceId{
										ResourceType: teamResourceType.Id,
										Resource:     groupIdentity.Id.String(),
									},
									ParentResourceId: parentResource,
								}
								grants = append(grants, parseIntoGrants(o.client.SyncGrantSources, teamResource, resource, value, action)...)
							} else {
								groupResource := &v2.Resource{
									Id: &v2.ResourceId{
										ResourceType: groupResourceType.Id,
										Resource:     groupIdentity.Id.String(),
									},
									ParentResourceId: parentResource,
								}
								grants = append(grants, parseIntoGrants(o.client.SyncGrantSources, groupResource, resource, value, action)...)
							}
						}
					}
				}
			}
		}
	}

	return grants, "", nil, nil
}

func parseIntoSecurityNamespaceResource(securityNamespace *security.SecurityNamespaceDescription) (*v2.Resource, error) {
	securityNamespaceResource, err := resource.NewResource(
		*securityNamespace.Name,
		securityNamespaceResourceType,
		securityNamespace.NamespaceId.String(),
	)
	if err != nil {
		return nil, err
	}

	return securityNamespaceResource, nil
}

func newSecurityNamespaceBuilder(c *client.AzureDevOpsClient) *securityNamespaceBuilder {
	return &securityNamespaceBuilder{
		resourceType: securityNamespaceResourceType,
		client:       c,
	}
}

func getEntitlementName(permission, actionName string, projectResource *v2.ResourceId) string {
	permissionEntitlementName := strings.Join([]string{permission, actionName}, "_")
	if projectResource != nil {
		permissionEntitlementName = strings.Join([]string{projectResource.Resource, permissionEntitlementName}, "_")
	}

	return permissionEntitlementName
}

func parseIntoEntitlements(resource *v2.Resource, actionDisplayName, actionName, projectName string) []*v2.Entitlement {
	var entitlements []*v2.Entitlement

	var projectResourceId *v2.ResourceId

	displayName := actionDisplayName

	permissionLevel := "organization"
	if projectName != "" {
		permissionLevel = fmt.Sprintf("project (%s)", projectName)
		displayName = fmt.Sprintf("%s(%s)", projectName, actionDisplayName)
		projectResourceId = &v2.ResourceId{
			ResourceType: projectResourceType.Id,
			Resource:     projectName,
		}
	}
	description := fmt.Sprintf("action \"%s\" for security namespace \"%s\" at %s level",
		actionDisplayName,
		resource.DisplayName,
		permissionLevel,
	)

	assigmentOptions := []entitlement.EntitlementOption{
		entitlement.WithGrantableTo(userResourceType, groupResourceType, teamResourceType),
		entitlement.WithDescription(description),
		entitlement.WithDisplayName(displayName),
	}

	entitlements = append(entitlements, entitlement.NewPermissionEntitlement(
		resource,
		getEntitlementName("allow", actionName, projectResourceId),
		assigmentOptions...,
	))
	entitlements = append(entitlements, entitlement.NewPermissionEntitlement(
		resource,
		getEntitlementName("deny", actionName, projectResourceId),
		assigmentOptions...,
	))

	return entitlements
}

func parseIntoGrants(syncGrantSources bool, grantResource, resource *v2.Resource, acesDictionary security.AccessControlEntry, action security.ActionDefinition) []*v2.Grant {
	var grants []*v2.Grant
	var basicGrantOptions []grant.GrantOption
	if syncGrantSources {
		basicGrantOptions = append(basicGrantOptions, grant.WithAnnotation(&v2.GrantExpandable{
			EntitlementIds: []string{
				fmt.Sprintf("team:%s:member", grantResource.Id.Resource),
				fmt.Sprintf("group:%s:member", grantResource.Id.Resource),
				fmt.Sprintf("group:%s:admin", grantResource.Id.Resource),
			},
			Shallow: true,
		}))
	}

	if *acesDictionary.Allow&*action.Bit != 0 {
		grants = append(grants, parseIntoGrantItem("allow", *action.Name, resource, grantResource, basicGrantOptions))
	}
	if *acesDictionary.Deny&*action.Bit != 0 {
		grants = append(grants, parseIntoGrantItem("deny", *action.Name, resource, grantResource, basicGrantOptions))
	}
	return grants
}

func parseIntoGrantItem(permission, actionName string, resource, grantResource *v2.Resource, grantOptions []grant.GrantOption) *v2.Grant {
	grantIdentifier := strings.Join([]string{permission, resource.Id.Resource, grantResource.Id.Resource, actionName}, ":")
	if grantResource.ParentResourceId != nil {
		grantIdentifier = strings.Join([]string{grantResource.ParentResourceId.Resource, grantIdentifier}, ":")
	}

	grantOptions = append(grantOptions, grant.WithAnnotation(&v2.V1Identifier{
		Id: grantIdentifier,
	}))
	return grant.NewGrant(
		resource,
		getEntitlementName(permission, actionName, grantResource.ParentResourceId),
		grantResource,
		grantOptions...,
	)
}
