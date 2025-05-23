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
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"go.uber.org/zap"
)

type teamBuilder struct {
	resourceType *v2.ResourceType
	client       client.AzureDevOpsClientInterface
}

var (
	adminPermission = "admin"

	permissions = []string{memberPermission, adminPermission}
)

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
		teamResource, err := parseIntoTeamResource(ctx, teamCopy)
		if err != nil {
			return nil, "", nil, err
		}
		resources = append(resources, teamResource)
	}

	return resources, "", nil, nil
}

func (o *teamBuilder) Entitlements(_ context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var entitlements []*v2.Entitlement

	for _, permission := range permissions {
		assigmentOptions := []entitlement.EntitlementOption{
			entitlement.WithGrantableTo(userResourceType),
			entitlement.WithDescription(fmt.Sprintf("%s membership type %s", resource.DisplayName, permission)),
			entitlement.WithDisplayName(permission),
		}

		entitlements = append(entitlements, entitlement.NewPermissionEntitlement(resource, permission, assigmentOptions...))
	}

	return entitlements, "", nil, nil
}

func (o *teamBuilder) Grants(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	var grants []*v2.Grant

	teamId := resource.Id.Resource
	if resource.ParentResourceId == nil {
		return grants, "", nil, nil
	}
	projectId := resource.ParentResourceId.Resource

	members, err := o.client.ListTeamMembers(ctx, projectId, teamId)
	if err != nil {
		return nil, "", nil, err
	}

	for _, member := range members {
		finalResource := &v2.Resource{}
		if member.Identity.IsContainer != nil && *member.Identity.IsContainer {
			finalResource.Id = &v2.ResourceId{
				ResourceType: groupResourceType.Id,
				Resource:     *member.Identity.Id,
			}
		} else {
			finalResource.Id = &v2.ResourceId{
				ResourceType: userResourceType.Id,
				Resource:     *member.Identity.Descriptor,
			}
		}
		permissionName := memberPermission
		if member.IsTeamAdmin != nil && *member.IsTeamAdmin {
			permissionName = adminPermission
		}
		var grantOptions []grant.GrantOption
		if permissionName == adminPermission {
			grantOptions = append(grantOptions, grant.WithAnnotation(&v2.GrantImmutable{}))
		}
		membershipGrant := grant.NewGrant(resource, permissionName, finalResource.Id, grantOptions...)
		grants = append(grants, membershipGrant)
	}
	return grants, "", nil, nil
}

func (o *teamBuilder) Grant(ctx context.Context, principal *v2.Resource, entitlementResource *v2.Entitlement) (annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)
	grantType := entitlementResource.DisplayName
	if grantType != memberPermission {
		l.Debug("Grant type is not supported", zap.String("grantType", grantType))
		return nil, fmt.Errorf("grant type %s not supported", grantType)
	}
	resourceId := entitlementResource.Resource.Id.Resource
	parsedUUID, err := uuid.Parse(resourceId)
	if err != nil {
		return nil, err
	}
	teamDescriptor, err := o.client.GetDescriptor(ctx, parsedUUID)
	if err != nil {
		l.Debug("Error getting group descriptor", zap.Error(err))
		return nil, err
	}
	memberDescriptor := principal.Id.Resource

	if parsedUUID, err := uuid.Parse(principal.Id.Resource); err == nil {
		memberDescriptor, err = o.client.GetDescriptor(ctx, parsedUUID)
		if err != nil {
			l.Debug("Error fetching principal descriptor", zap.Error(err))
			return nil, err
		}
	}

	_, err = o.client.CreateMembership(ctx, teamDescriptor, memberDescriptor)
	if err != nil {
		l.Debug("Error creating membership", zap.Error(err))
		return nil, err
	}

	return nil, nil
}

func (o *teamBuilder) Revoke(ctx context.Context, grantResource *v2.Grant) (annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)
	grantType := grantResource.Entitlement.DisplayName
	if grantType != memberPermission {
		l.Debug("Grant type is not supported", zap.String("grantType", grantType))
		return nil, fmt.Errorf("grant type %s not supported", grantType)
	}
	principal := grantResource.Principal
	principalDescriptor := principal.Id.Resource
	if parsedUUID, err := uuid.Parse(principal.Id.Resource); err == nil {
		principalDescriptor, err = o.client.GetDescriptor(ctx, parsedUUID)
		if err != nil {
			l.Debug("Error fetching principal descriptor", zap.Error(err))
			return nil, err
		}
	}
	resourceId := grantResource.Entitlement.Resource.Id.Resource
	parsedUUID, err := uuid.Parse(resourceId)
	if err != nil {
		l.Debug("Error parsing team uuid", zap.Error(err))
		return nil, err
	}
	teamDescriptor, err := o.client.GetDescriptor(ctx, parsedUUID)
	if err != nil {
		return nil, err
	}
	err = o.client.RevokeMembership(ctx, teamDescriptor, principalDescriptor)
	if err != nil {
		l.Debug("Error revoking team membership", zap.Error(err))
		return nil, err
	}

	return nil, nil
}

func parseIntoTeamResource(ctx context.Context, team *core.WebApiTeam) (*v2.Resource, error) {
	l := ctxzap.Extract(ctx)

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
	parentResourceId, err := resource.NewResourceID(projectResourceType, team.ProjectId.String())
	if err != nil {
		l.Error(fmt.Sprintf("Failed to create parent resource: %s", err))
	}

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
