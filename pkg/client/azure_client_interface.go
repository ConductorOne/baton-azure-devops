package client

import (
	"context"

	"github.com/google/uuid"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/graph"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"
)

// AzureDevOpsClient defines the interface for interacting with Azure DevOps services.
type AzureDevOpsClientInterface interface {
	GetDescriptor(ctx context.Context, resource uuid.UUID) (string, error)
	CreateMembership(ctx context.Context, teamDescriptor, principalDescriptor string) (*graph.GraphMembership, error)
	RevokeMembership(ctx context.Context, teamDescriptor, principalDescriptor string) error
	ListTeams(ctx context.Context) ([]core.WebApiTeam, error)
	ListTeamMembers(ctx context.Context, projectId, teamId string) ([]webapi.TeamMember, error)
}
