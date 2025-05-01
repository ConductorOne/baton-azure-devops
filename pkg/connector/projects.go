package connector

import (
	"context"
	"strings"

	"github.com/conductorone/baton-azure-devops/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
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
	connector    *Connector
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
	namespaces, err := o.client.ListSecurityNamespaces(ctx, securityNamespaces)
	if err != nil {
		return nil, "", nil, err
	}

	return getEntitlementsFromSecurityNamespaces(namespaces, resource), "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *projectBuilder) Grants(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	err := o.connector.loadUsers(ctx)
	if err != nil {
		return nil, "", nil, err
	}

	namespaces, err := o.client.ListSecurityNamespaces(ctx, securityNamespaces)
	if err != nil {
		return nil, "", nil, err
	}

	grants, err := getGrantsFromSecurityNamespaces(ctx, o.client, o.connector.users, namespaces, resource)
	if err != nil {
		return nil, "", nil, err
	}
	return grants, "", nil, nil
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

func newProjectBuilder(c *client.AzureDevOpsClient, d *Connector) *projectBuilder {
	return &projectBuilder{
		resourceType: projectResourceType,
		client:       c,
		connector:    d,
	}
}
