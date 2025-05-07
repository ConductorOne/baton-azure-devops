package client

import (
	"context"

	"github.com/google/uuid"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/graph"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"
	"github.com/stretchr/testify/mock"
)

type MockAzureClient struct {
	mock.Mock
}

func (m *MockAzureClient) GetDescriptor(ctx context.Context, resource uuid.UUID) (string, error) {
	args := m.Called(ctx, resource)
	return args.String(0), args.Error(1)
}

func (m *MockAzureClient) RevokeMembership(ctx context.Context, teamDescriptor, principalDescriptor string) error {
	args := m.Called(ctx, teamDescriptor, principalDescriptor)
	return args.Error(0)
}

func (m *MockAzureClient) CreateMembership(ctx context.Context, teamDescriptor, principalDescriptor string) (*graph.GraphMembership, error) {
	args := m.Called(ctx, teamDescriptor, principalDescriptor)
	return args.Get(0).(*graph.GraphMembership), args.Error(1)
}

func (m *MockAzureClient) ListTeams(ctx context.Context) ([]core.WebApiTeam, error) {
	args := m.Called(ctx)
	return args.Get(0).([]core.WebApiTeam), args.Error(1)
}

func (m *MockAzureClient) ListTeamMembers(ctx context.Context, projectId, teamId string) ([]webapi.TeamMember, error) {
	args := m.Called(ctx, projectId, teamId)
	return args.Get(0).([]webapi.TeamMember), args.Error(1)
}
