package connector

import (
	"context"
	"testing"

	mockService "github.com/conductorone/baton-azure-devops/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/google/uuid"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamBuilderWithMockClient(t *testing.T) {
	// Create test team ID and matching uuid, and team descriptor
	const testTeamId = "11c0f886-25c4-11f0-b643-325096b39f47"
	var parsedTeamUUUID = uuid.UUID{0x11, 0xc0, 0xf8, 0x86, 0x25, 0xc4, 0x11, 0xf0, 0xb6, 0x43, 0x32, 0x50, 0x96, 0xb3, 0x9f, 0x47}
	const testTeamDescriptor = "testTeamDescriptor"
	const testPrincipalDescriptor = "aad.OTk5ZDIwNjQtOWQyMy03YzBmLWFmYDUtNWQ3ZmU1MzNhMTc4"
	containerDescriptor := "ContainerDescriptor"
	memberDescriptor := "MemberDescriptor"
	testMembership := &graph.GraphMembership{
		ContainerDescriptor: &containerDescriptor,
		MemberDescriptor:    &memberDescriptor,
	}
	ctx := context.Background()
	mockClient := &mockService.MockAzureClient{}

	// Mock the behavior of the methods
	mockClient.On("GetDescriptor", ctx, parsedTeamUUUID).Return(testTeamDescriptor, nil)
	mockClient.On("CreateMembership", ctx, testTeamDescriptor, testPrincipalDescriptor).Return(testMembership, nil)
	mockClient.On("RevokeMembership", ctx, testTeamDescriptor, testPrincipalDescriptor).Return(nil)
	builder := &teamBuilder{client: mockClient}

	t.Run("Grant team member entitlement", func(t *testing.T) {
		principal := &v2.Resource{Id: &v2.ResourceId{Resource: testPrincipalDescriptor}}
		entitlementResource := &v2.Entitlement{DisplayName: "member", Resource: &v2.Resource{Id: &v2.ResourceId{Resource: testTeamId}}}

		_, err := builder.Grant(ctx, principal, entitlementResource)
		require.NoError(t, err)
	})

	t.Run("Grant invalid team admin entitlement should error", func(t *testing.T) {
		principal := &v2.Resource{Id: &v2.ResourceId{Resource: testPrincipalDescriptor}}
		entitlementResource := &v2.Entitlement{DisplayName: "admin", Resource: &v2.Resource{Id: &v2.ResourceId{Resource: testTeamId}}}

		_, err := builder.Grant(ctx, principal, entitlementResource)
		require.Error(t, err)
	})

	t.Run("Revoke team member entitlement", func(t *testing.T) {
		entitlementResource := &v2.Entitlement{DisplayName: "member", Resource: &v2.Resource{Id: &v2.ResourceId{Resource: testTeamId}}}
		grant := &v2.Grant{
			Principal:   &v2.Resource{Id: &v2.ResourceId{Resource: testPrincipalDescriptor}},
			Entitlement: entitlementResource,
		}
		_, err := builder.Revoke(ctx, grant)
		assert.NoError(t, err)
	})

	t.Run("Revoke invalid team admin entitlement should error", func(t *testing.T) {
		entitlementResource := &v2.Entitlement{DisplayName: "admin", Resource: &v2.Resource{Id: &v2.ResourceId{Resource: testTeamId}}}
		grant := &v2.Grant{
			Principal:   &v2.Resource{Id: &v2.ResourceId{Resource: testPrincipalDescriptor}},
			Entitlement: entitlementResource,
		}
		_, err := builder.Revoke(ctx, grant)
		require.Error(t, err)
	})

	mockClient.AssertExpectations(t)
}
