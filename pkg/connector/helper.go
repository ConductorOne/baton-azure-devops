package connector

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/conductorone/baton-azure-devops/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/security"
)

func unmarshalProperties(properties interface{}) (map[string]interface{}, error) {
	rawBytes, err := json.Marshal(properties)
	if err != nil {
		return nil, err
	}

	var propsMap map[string]interface{}
	err = json.Unmarshal(rawBytes, &propsMap)
	if err != nil {
		return nil, err
	}

	return propsMap, nil
}

func getIdentityResourceByDescriptor(client *client.AzureDevOpsClient, ctx context.Context, descriptor string, userMap map[string]string) (*v2.Resource, error) {
	// get user email to map with user
	parts := strings.Split(descriptor, `\`)

	if len(parts) == 2 {
		userPrincipalName := parts[1]
		if userMap[userPrincipalName] != "" {
			userResource := &v2.Resource{
				Id: &v2.ResourceId{
					ResourceType: userResourceType.Id,
					Resource:     userMap[userPrincipalName],
				},
			}
			return userResource, nil
		}
	} else {
		teamsMap, err := client.ListTeamIDs(ctx)
		if err != nil {
			return nil, err
		}
		groupIdentities, err := client.ListIdentities(ctx, "", descriptor)
		if err != nil {
			return nil, err
		}
		if len(groupIdentities) > 0 {
			groupIdentity := groupIdentities[0]
			if groupIdentity.Id != nil {
				if _, isTeam := teamsMap[groupIdentity.Id.String()]; isTeam {
					teamResource := &v2.Resource{
						Id: &v2.ResourceId{
							ResourceType: teamResourceType.Id,
							Resource:     groupIdentity.Id.String(),
						},
					}
					return teamResource, nil
				} else {
					groupResource := &v2.Resource{
						Id: &v2.ResourceId{
							ResourceType: groupResourceType.Id,
							Resource:     groupIdentity.Id.String(),
						},
					}
					return groupResource, nil
				}
			}
		}
	}
	return nil, errors.New("no identity resource found")
}

func getEntitlementsFromSecurityNamespaces(namespaces []security.SecurityNamespaceDescription, resource *v2.Resource) []*v2.Entitlement {
	var entitlements []*v2.Entitlement

	for _, namespace := range namespaces {
		readPermissionName := getPermissionName(resource.DisplayName, *namespace.Name, "read")
		readOptions := []entitlement.EntitlementOption{
			entitlement.WithGrantableTo(userResourceType, groupResourceType, teamResourceType),
			entitlement.WithDescription(fmt.Sprintf("Read permission in %s security namespace at %s %s level", *namespace.Name, resource.DisplayName, resource.Id.ResourceType)),
			entitlement.WithDisplayName(readPermissionName),
		}

		writePermissionName := getPermissionName(resource.DisplayName, *namespace.Name, "write")
		writeOptions := []entitlement.EntitlementOption{
			entitlement.WithGrantableTo(userResourceType, groupResourceType, teamResourceType),
			entitlement.WithDescription(fmt.Sprintf("Write permission in %s security namespace at %s %s level", *namespace.Name, resource.DisplayName, resource.Id.ResourceType)),
			entitlement.WithDisplayName(writePermissionName),
		}

		entitlements = append(entitlements, entitlement.NewPermissionEntitlement(resource, readPermissionName, readOptions...))
		entitlements = append(entitlements, entitlement.NewPermissionEntitlement(resource, writePermissionName, writeOptions...))
	}

	return entitlements
}

func getGrantsFromSecurityNamespaces(
	ctx context.Context,
	client *client.AzureDevOpsClient,
	users map[string]string,
	namespaces []security.SecurityNamespaceDescription,
	resource *v2.Resource,
) ([]*v2.Grant, error) {
	var grants []*v2.Grant

	for _, namespace := range namespaces {
		readPermissionBit := *namespace.ReadPermission
		writePermissionBit := *namespace.WritePermission
		bitMask := *namespace.SystemBitMask

		ACLs, err := client.ListAccessControlsBySecurityNamespace(ctx, *namespace.NamespaceId, parseTokenBySecurityNamespace(namespace.NamespaceId.String(), resource))
		if err != nil {
			return nil, err
		}
		for _, acl := range ACLs {
			for _, ace := range *acl.AcesDictionary {
				grantResource, err := getIdentityResourceByDescriptor(client, ctx, *ace.Descriptor, users)
				if err != nil {
					continue
				}
				var basicGrantOptions []grant.GrantOption

				if client.SyncGrantSources {
					basicGrantOptions = append(basicGrantOptions, grant.WithAnnotation(&v2.GrantExpandable{
						EntitlementIds: []string{
							fmt.Sprintf("team:%s:member", grantResource.Id.Resource),
							fmt.Sprintf("group:%s:member", grantResource.Id.Resource),
							fmt.Sprintf("group:%s:admin", grantResource.Id.Resource),
						},
						Shallow: true,
					}))
				}
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

	return grants, nil
}

func parseTokenBySecurityNamespace(securityNamespace string, resource *v2.Resource) string {
	switch securityNamespace {
	case projectSecurityNamespace:
		return fmt.Sprintf("$PROJECT:vstfs:///Classification/TeamProject/%s", resource.Id.Resource)
	case taggingSecurityNamespace:
		return fmt.Sprintf("/%s", resource.Id.Resource)
	case versionControlItemsSecurityNamespace:
		return fmt.Sprintf("$/%s", resource.DisplayName)
	case analyticsViewsSecurityNamespace:
		return fmt.Sprintf("$/Shared/%s", resource.Id.Resource)
	case buildSecurityNamespace:
		return resource.Id.Resource
	case gitRepositoriesSecurityNamespace:
		if resource.ParentResourceId != nil {
			return fmt.Sprintf("repoV2/%s/%s", resource.ParentResourceId.Resource, resource.Id.Resource)
		}
		return fmt.Sprintf("repoV2/%s", resource.Id.Resource)
	case metaTaskSecurityNamespace:
		return resource.Id.Resource
	case releaseManagementSecurityNamespace:
		return resource.Id.Resource
	}
	return ""
}
