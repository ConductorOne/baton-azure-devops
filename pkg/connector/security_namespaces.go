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
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/identity"
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
	var entitlements []*v2.Entitlement

	namespaceUUID, err := uuid.Parse(resource.Id.Resource)
	if err != nil {
		return nil, "", nil, err
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

			assigmentOptions := []entitlement.EntitlementOption{
				entitlement.WithGrantableTo(userResourceType),
				entitlement.WithGrantableTo(groupResourceType),
				entitlement.WithDescription(fmt.Sprintf("%s for action %s", resource.DisplayName, actionDisplayName)),
				entitlement.WithDisplayName(actionDisplayName),
			}

			entitlements = append(entitlements, entitlement.NewPermissionEntitlement(resource, fmt.Sprintf("allow_%v", actionName), assigmentOptions...))
			entitlements = append(entitlements, entitlement.NewPermissionEntitlement(resource, fmt.Sprintf("deny_%v", actionName), assigmentOptions...))
		}
	}

	return entitlements, "", nil, nil
}

func (o *securityNamespaceBuilder) Grants(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
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
					//user
					grants = append(grants, parseIntoUserGrants(value, action, resource, userMap)...)
				} else {
					//if it is a group or team add grant with expansion, missing support to differentiate between groups and teams
					groupIdentities, err := o.client.ListIdentities(ctx, "", *value.Descriptor)
					if err != nil {
						continue
					}
					if len(groupIdentities) > 0 {
						groupIdentity := groupIdentities[0]
						if groupIdentity.Id != nil {
							grants = append(grants, parseIntoGroupGrants(o.client.SyncGrantSources, value, action, resource, groupIdentity)...)
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

func parseIntoUserGrants(acesDictionary security.AccessControlEntry, action security.ActionDefinition, resource *v2.Resource, userMap map[string]string) []*v2.Grant {
	var grants []*v2.Grant
	parts := strings.Split(*acesDictionary.Descriptor, `\`)
	if len(parts) != 2 {
		return grants
	}
	userPrincipalName := parts[1]

	if userMap[userPrincipalName] != "" {
		userResource := &v2.Resource{
			Id: &v2.ResourceId{
				ResourceType: userResourceType.Id,
				Resource:     userMap[userPrincipalName],
			},
		}
		if *acesDictionary.Allow&*action.Bit != 0 {
			allowGrant := grant.NewGrant(resource, fmt.Sprintf("allow_%v", *action.Name), userResource, grant.WithAnnotation(&v2.V1Identifier{
				Id: fmt.Sprintf("allow:%s:%s:%s", resource.Id.Resource, userMap[userPrincipalName], *action.Name),
			}))
			grants = append(grants, allowGrant)
		}
		if *acesDictionary.Deny&*action.Bit != 0 {
			denyGrant := grant.NewGrant(resource, fmt.Sprintf("deny_%v", *action.Name), userResource, grant.WithAnnotation(&v2.V1Identifier{
				Id: fmt.Sprintf("deny:%s:%s:%s", resource.Id.Resource, userMap[userPrincipalName], *action.Name),
			}))
			grants = append(grants, denyGrant)
		}
	}
	return grants
}

func parseIntoGroupGrants(syncGrantSources bool, acesDictionary security.AccessControlEntry, action security.ActionDefinition, resource *v2.Resource, groupIdentity identity.Identity) []*v2.Grant {
	var grants []*v2.Grant

	//check if it is a team? listing teams?
	groupResource := &v2.Resource{
		Id: &v2.ResourceId{
			ResourceType: groupResourceType.Id,
			Resource:     groupIdentity.Id.String(),
		},
	}
	var basicGrantOptions []grant.GrantOption
	if syncGrantSources {
		basicGrantOptions = append(basicGrantOptions, grant.WithAnnotation(&v2.GrantExpandable{
			EntitlementIds: []string{
				fmt.Sprintf("group:%s:member", groupIdentity.Id.String()),
				fmt.Sprintf("group:%s:admin", groupIdentity.Id.String()),
			},
			Shallow: true,
		}))
	}
	if *acesDictionary.Allow&*action.Bit != 0 {
		grantOptions := make([]grant.GrantOption, len(basicGrantOptions))
		copy(grantOptions, basicGrantOptions)

		grantOptions = append(grantOptions, grant.WithAnnotation(&v2.V1Identifier{
			Id: fmt.Sprintf("allow:%s:%s:%s", resource.Id.Resource, groupIdentity.Id.String(), *action.Name),
		}))
		allowGrant := grant.NewGrant(resource, fmt.Sprintf("allow_%v", *action.Name), groupResource, grantOptions...)
		grants = append(grants, allowGrant)
	}
	if *acesDictionary.Deny&*action.Bit != 0 {
		grantOptions := make([]grant.GrantOption, len(basicGrantOptions))
		copy(grantOptions, basicGrantOptions)

		grantOptions = append(grantOptions, grant.WithAnnotation(&v2.V1Identifier{
			Id: fmt.Sprintf("deny:%s:%s:%s", resource.Id.Resource, groupIdentity.Id.String(), *action.Name),
		}))
		denyGrant := grant.NewGrant(resource, fmt.Sprintf("deny_%v", *action.Name), groupResource, grantOptions...)
		grants = append(grants, denyGrant)
	}
	return grants
}
