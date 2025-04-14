package client

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/graph"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/identity"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/security"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"
	"go.uber.org/zap"
)

type AzureDevOpsClient struct {
	SyncGrantSources bool
	coreClient       core.Client
	graphClient      graph.Client
	securityClient   security.Client
	identityClient   identity.Client
}

func New(ctx context.Context, personalAccessToken, organization string, syncGrantSources bool) (*AzureDevOpsClient, error) {
	l := ctxzap.Extract(ctx)
	connection := azuredevops.NewPatConnection(organization, personalAccessToken)

	// Create a client to interact with the Core area
	coreClient, err := core.NewClient(ctx, connection)
	if err != nil {
		l.Error("error creating core client", zap.Error(err))
	}

	graphClient, err := graph.NewClient(ctx, connection)
	if err != nil {
		l.Info("error creating graph client", zap.Error(err))
	}

	securityClient := security.NewClient(ctx, connection)

	identityClient, err := identity.NewClient(ctx, connection)
	if err != nil {
		l.Info("error creating identity client", zap.Error(err))
	}

	client := AzureDevOpsClient{
		coreClient:       coreClient,
		graphClient:      graphClient,
		securityClient:   securityClient,
		identityClient:   identityClient,
		SyncGrantSources: syncGrantSources,
	}

	return &client, nil
}

func (c *AzureDevOpsClient) ListUsers(ctx context.Context, nextContinuationToken string) ([]graph.GraphUser, string, error) {
	l := ctxzap.Extract(ctx)
	nextPageToken := ""

	userArgs := graph.ListUsersArgs{}
	if nextContinuationToken != "" {
		userArgs.ContinuationToken = &nextContinuationToken
	}

	users, err := c.graphClient.ListUsers(ctx, userArgs)
	if err != nil {
		l.Error(fmt.Sprintf("Error getting resources: %s", err))
		return nil, "", err
	}

	continuationToken := *users.ContinuationToken
	if continuationToken != nil && len(continuationToken) > 0 {
		nextPageToken = continuationToken[0]
	}

	return *users.GraphUsers, nextPageToken, nil
}

func (c *AzureDevOpsClient) ListProjects(ctx context.Context, nextContinuationToken string) ([]core.TeamProjectReference, string, error) {
	l := ctxzap.Extract(ctx)

	projectArgs := core.GetProjectsArgs{}
	if nextContinuationToken != "" {
		nextContinuationTokenInt, err := strconv.Atoi(nextContinuationToken)
		if err == nil {
			projectArgs.ContinuationToken = &nextContinuationTokenInt
		}
	}

	projects, err := c.coreClient.GetProjects(ctx, projectArgs)
	if err != nil {
		l.Error(fmt.Sprintf("Error getting resources: %s", err))
		return nil, "", err
	}

	return projects.Value, projects.ContinuationToken, nil
}

func (c *AzureDevOpsClient) ListTeams(ctx context.Context) ([]core.WebApiTeam, error) {
	l := ctxzap.Extract(ctx)

	//Teams client query is not supporting pagination
	teams, err := c.coreClient.GetAllTeams(ctx, core.GetAllTeamsArgs{})
	if err != nil {
		l.Error(fmt.Sprintf("Error getting resources: %s", err))
		return nil, err
	}

	return *teams, nil
}

func (c *AzureDevOpsClient) ListTeamMembers(ctx context.Context, projectId, teamId string) ([]webapi.TeamMember, error) {
	l := ctxzap.Extract(ctx)

	teamMembersArgs := core.GetTeamMembersWithExtendedPropertiesArgs{
		ProjectId: &projectId,
		TeamId:    &teamId,
	}
	teamMembers, err := c.coreClient.GetTeamMembersWithExtendedProperties(ctx, teamMembersArgs)
	if err != nil {
		l.Error(fmt.Sprintf("Error getting resources: %s", err))
		return nil, err
	}
	return *teamMembers, nil
}

func (c *AzureDevOpsClient) ListGroups(ctx context.Context, nextContinuationToken string) ([]graph.GraphGroup, string, error) {
	l := ctxzap.Extract(ctx)
	nextPageToken := ""

	groupArgs := graph.ListGroupsArgs{}
	if nextContinuationToken != "" {
		groupArgs.ContinuationToken = &nextContinuationToken
	}

	groups, err := c.graphClient.ListGroups(ctx, groupArgs)
	if err != nil {
		l.Error(fmt.Sprintf("Error getting resources: %s", err))
		return nil, "", err
	}

	continuationToken := *groups.ContinuationToken
	if continuationToken != nil && len(continuationToken) > 0 {
		nextPageToken = continuationToken[0]
	}

	return *groups.GraphGroups, nextPageToken, nil
}

func (c *AzureDevOpsClient) ListIdentities(ctx context.Context, identityIds string, descriptors string) ([]identity.Identity, error) {
	l := ctxzap.Extract(ctx)

	readIdentitiesArgs := identity.ReadIdentitiesArgs{
		QueryMembership: &identity.QueryMembershipValues.Expanded,
	}
	if identityIds != "" {
		readIdentitiesArgs.IdentityIds = &identityIds
	}
	if descriptors != "" {
		readIdentitiesArgs.Descriptors = &descriptors
	}
	identities, err := c.identityClient.ReadIdentities(ctx, readIdentitiesArgs)

	if err != nil {
		l.Error(fmt.Sprintf("Error getting resources: %s", err))
		return nil, err
	}

	return *identities, nil
}

func (c *AzureDevOpsClient) ListSecurityNamespaces(ctx context.Context) ([]security.SecurityNamespaceDescription, error) {
	l := ctxzap.Extract(ctx)

	namespaces, err := c.securityClient.QuerySecurityNamespaces(ctx, security.QuerySecurityNamespacesArgs{})
	if err != nil {
		l.Error(fmt.Sprintf("Error getting resources: %s", err))
		return nil, err
	}

	return *namespaces, nil
}

func (c *AzureDevOpsClient) ListActionsBySecurityNamespace(ctx context.Context, securityNamespaceId uuid.UUID) ([]security.ActionDefinition, error) {
	l := ctxzap.Extract(ctx)

	securityNamespace, err := c.securityClient.QuerySecurityNamespaces(ctx, security.QuerySecurityNamespacesArgs{SecurityNamespaceId: &securityNamespaceId})
	if err != nil {
		l.Error(fmt.Sprintf("Error getting resources: %s", err))
		return nil, err
	}

	if *securityNamespace != nil && len(*securityNamespace) > 0 {
		return *(*securityNamespace)[0].Actions, nil
	}

	return nil, nil
}

func (c *AzureDevOpsClient) ListAccessControlsBySecurityNamespace(ctx context.Context, securityNamespaceId uuid.UUID) ([]security.AccessControlList, error) {
	l := ctxzap.Extract(ctx)

	lists, err := c.securityClient.QueryAccessControlLists(ctx, security.QueryAccessControlListsArgs{SecurityNamespaceId: &securityNamespaceId})
	if err != nil {
		l.Error(fmt.Sprintf("Error getting resources: %s", err))
		return nil, err
	}

	return *lists, nil
}

func (c *AzureDevOpsClient) GetUsersMap(ctx context.Context) (map[string]string, error) {
	userMap := make(map[string]string)
	nextPageToken := ""

	for {
		users, nextPageToken, err := c.ListUsers(ctx, nextPageToken)
		if err != nil {
			return nil, err
		}

		for _, user := range users {
			userMap[*user.PrincipalName] = *user.Descriptor
		}

		if nextPageToken == "" {
			break
		}
	}

	return userMap, nil
}

func (c *AzureDevOpsClient) GetIdentity(ctx context.Context, identityID *string) (string, error) {
	l := ctxzap.Extract(ctx)

	identities, err := c.identityClient.ReadIdentities(ctx, identity.ReadIdentitiesArgs{Descriptors: identityID})
	if err != nil {
		l.Error(fmt.Sprintf("Error getting resources: %s;; for identity %s", err, *identityID))
		return "", err
	}

	if identities != nil && len(*identities) > 0 {
		first := (*identities)[0]
		if first.SubjectDescriptor != nil {
			return *first.SubjectDescriptor, nil
		}
	}

	return "", nil
}
