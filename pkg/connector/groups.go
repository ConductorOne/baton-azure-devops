package connector

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/conductorone/baton-azure-devops/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/entitlement"
	"github.com/conductorone/baton-sdk/pkg/types/grant"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/graph"
)

type groupBuilder struct {
	resourceType *v2.ResourceType
	client       *client.AzureDevOpsClient
}

var memberPermission = "member"

func (o *groupBuilder) ResourceType(_ context.Context) *v2.ResourceType {
	return groupResourceType
}

func (o *groupBuilder) List(ctx context.Context, _ *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	var resources []*v2.Resource

	groups, nextPageToken, err := o.client.ListOnlyGroups(ctx, pToken.Token)
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

func (o *groupBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var entitlements []*v2.Entitlement

	assigmentOptions := []entitlement.EntitlementOption{
		entitlement.WithGrantableTo(userResourceType, groupResourceType),
		entitlement.WithDescription(fmt.Sprintf("%s membership type %s", resource.DisplayName, memberPermission)),
		entitlement.WithDisplayName(memberPermission),
	}

	entitlements = append(entitlements, entitlement.NewPermissionEntitlement(resource, memberPermission, assigmentOptions...))

	return entitlements, "", nil, nil
}

func (o *groupBuilder) Grants(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	var grants []*v2.Grant

	groupId := resource.Id.Resource

	groupIdentities, err := o.client.ListIdentities(ctx, groupId, "")
	if err != nil {
		return nil, "", nil, err
	}

	for _, groupIdentity := range groupIdentities {
		if groupIdentity.MemberIds != nil {
			if len(*groupIdentity.MemberIds) > 0 {
				memberIDs := make([]string, len(*groupIdentity.MemberIds))
				for _, memberId := range *groupIdentity.MemberIds {
					memberIDs = append(memberIDs, memberId.String())
				}
				memberIDsStr := strings.Join(memberIDs, ",")
				identities, err := o.client.ListIdentities(ctx, memberIDsStr, "")
				if err != nil {
					return nil, "", nil, err
				}
				for _, member := range identities {
					properties, err := unmarshalProperties(member.Properties)
					if err != nil {
						continue
					}
					schema, err := unmarshalProperties(properties["SchemaClassName"])
					if err != nil {
						continue
					}
					if schema["$value"] == "User" {
						userResource := &v2.Resource{
							Id: &v2.ResourceId{
								ResourceType: userResourceType.Id,
								Resource:     *member.SubjectDescriptor,
							},
						}
						membershipGrant := grant.NewGrant(resource, memberPermission, userResource.Id)
						grants = append(grants, membershipGrant)
					}
					if schema["$value"] == "Group" {
						groupResource := &v2.Resource{
							Id: &v2.ResourceId{
								ResourceType: groupResourceType.Id,
								Resource:     member.Id.String(),
							},
						}
						membershipGrant := grant.NewGrant(resource, memberPermission, groupResource.Id)
						grants = append(grants, membershipGrant)
					}
				}
			}
		}
	}
	return grants, "", nil, nil
}

func parseIntoGroupResource(group *graph.GraphGroup) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"group_id":     *group.OriginId,
		"display_name": *group.DisplayName,
		"description":  *group.Description,
		"url":          *group.Url,
		"descriptor":   *group.Descriptor,
	}

	groupTraits := []resource.GroupTraitOption{
		resource.WithGroupProfile(profile),
	}

	var parentId *v2.ResourceId = nil
	if strings.Contains(*group.Domain, "TeamProject") {
		parts := strings.Split(*group.Domain, "/")
		parentId = &v2.ResourceId{
			ResourceType: projectResourceType.Id,
			Resource:     parts[len(parts)-1],
		}
	}

	ret, err := resource.NewGroupResource(
		*group.DisplayName,
		groupResourceType,
		*group.OriginId,
		groupTraits,
		resource.WithDescription(*group.Description),
		resource.WithParentResourceID(parentId),
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
