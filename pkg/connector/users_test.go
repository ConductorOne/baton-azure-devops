package connector

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/conductorone/baton-azure-devops/pkg/client/userentitlement"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"

	clientMocks "github.com/conductorone/baton-azure-devops/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/google/uuid"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/graph"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/licensing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func makeUserEntitlement(id int) userentitlement.UserEntitlement {
	uid := uuid.New()
	name := fmt.Sprintf("user%d", id)
	return userentitlement.UserEntitlement{
		User: &graph.GraphUser{
			Descriptor:  ptrString(uid.String()),
			DisplayName: &name,
			MailAddress: ptrString(fmt.Sprintf("%s@example.com", name)),
		},
		AccessLevel:      &licensing.AccessLevel{},
		LastAccessedDate: &azuredevops.Time{Time: time.Now()},
	}
}

func ptrString(s string) *string { return &s }

func TestUserBuilder_List_Pagination(t *testing.T) {
	mockClient := &clientMocks.MockAzureUserClient{}
	builder := &userBuilder{client: mockClient}

	// Prepare 3 pages of 5 users each
	page1 := []userentitlement.UserEntitlement{}
	page2 := []userentitlement.UserEntitlement{}
	page3 := []userentitlement.UserEntitlement{}
	for i := 1; i <= 5; i++ {
		page1 = append(page1, makeUserEntitlement(i))
		page2 = append(page2, makeUserEntitlement(i+5))
		page3 = append(page3, makeUserEntitlement(i+10))
	}

	// Set up the mock to return a continuation token for the first two pages
	mockClient.On("ListUsers", mock.Anything, "").Return(page1, "5", nil).Once()
	mockClient.On("ListUsers", mock.Anything, "5").Return(page2, "10", nil).Once()
	mockClient.On("ListUsers", mock.Anything, "10").Return(page3, "", nil).Once()

	var allUsers []*v2.Resource
	var nextToken string
	// var err error

	// Simulate paginated fetching
	for token := ""; ; token = nextToken {
		users, next, _, err := builder.List(context.Background(), nil, &pagination.Token{Token: token})
		assert.NoError(t, err)
		allUsers = append(allUsers, users...)
		if next == "" {
			break
		}
		nextToken = next
	}

	assert.Equal(t, 15, len(allUsers))
	mockClient.AssertExpectations(t)
}
