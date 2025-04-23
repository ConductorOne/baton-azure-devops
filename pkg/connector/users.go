package connector

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-azure-devops/pkg/client"
	"github.com/conductorone/baton-azure-devops/pkg/client/userentitlement"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/accounts"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/graph"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/licensing"
)

type userBuilder struct {
	resourceType *v2.ResourceType
	client       *client.AzureDevOpsClient
}

func (o *userBuilder) ResourceType(_ context.Context) *v2.ResourceType {
	return userResourceType
}

// List returns all the users from the database as resource objects.
// Users include a UserTrait because they are the 'shape' of a standard user.
func (o *userBuilder) List(ctx context.Context, _ *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	var resources []*v2.Resource

	users, nextPageToken, err := o.client.ListUsers(ctx, pToken.Token)
	if err != nil {
		return nil, "", nil, err
	}

	for _, user := range users {
		userCopy := &user
		organizationResource, err := parseIntoUserResource(userCopy)
		if err != nil {
			return nil, "", nil, err
		}
		resources = append(resources, organizationResource)
	}

	return resources, nextPageToken, nil, nil
}

func (o *userBuilder) CreateAccountCapabilityDetails(ctx context.Context) (*v2.CredentialDetailsAccountProvisioning, annotations.Annotations, error) {
	return &v2.CredentialDetailsAccountProvisioning{
		SupportedCredentialOptions: []v2.CapabilityDetailCredentialOption{
			v2.CapabilityDetailCredentialOption_CAPABILITY_DETAIL_CREDENTIAL_OPTION_NO_PASSWORD,
		},
		PreferredCredentialOption: v2.CapabilityDetailCredentialOption_CAPABILITY_DETAIL_CREDENTIAL_OPTION_NO_PASSWORD,
	}, nil, nil
}

func (o *userBuilder) CreateAccount(
	ctx context.Context,
	accountInfo *v2.AccountInfo,
	_ *v2.CredentialOptions) (
	connectorbuilder.CreateAccountResponse,
	[]*v2.PlaintextData,
	annotations.Annotations,
	error,
) {
	profile := accountInfo.GetProfile().AsMap()

	principalNameRaw, ok := profile["principal_name"]
	if !ok {
		return nil, nil, nil, fmt.Errorf("missing 'principal_name' in account profile")
	}
	principalName := principalNameRaw.(string)

	licenseTypeRaw, ok := profile["license_type"]
	if !ok {
		return nil, nil, nil, fmt.Errorf("missing 'license_type' in account profile")
	}
	licenseType := licenseTypeRaw.(string)

	accountLicenseType, licensingSource, err := mapLicenseType(licenseType)
	if err != nil {
		return nil, nil, nil, err
	}

	// In azure devops the subjectKind is always user.
	subjectKind := "user"
	userEntitlement := &userentitlement.UserEntitlement{
		AccessLevel: &licensing.AccessLevel{
			LicensingSource:    licensingSource,
			AccountLicenseType: accountLicenseType,
		},
		User: &graph.GraphUser{
			PrincipalName: &principalName,
			SubjectKind:   &subjectKind,
		},
	}

	created, err := o.client.CreateUserEntitlement(ctx, userEntitlement)
	if err != nil {
		return nil, nil, nil, err
	}

	resourceC, err := parseIntoUserResource(created)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to build user resource: %w", err)
	}

	return &v2.CreateAccountResponse_SuccessResult{
		Resource: resourceC,
	}, nil, nil, nil
}

// Function to validate and parse the license type.
// Valid license types are:
// - express
// - stakeholder
// - Visual Studio Subscriber
// If the license type is not valid, it returns an error.
func mapLicenseType(input string) (*licensing.AccountLicenseType, *licensing.LicensingSource, error) {
	switch input {
	case "express":
		return &licensing.AccountLicenseTypeValues.Express, &licensing.LicensingSourceValues.Account, nil
	case "stakeholder":
		return &licensing.AccountLicenseTypeValues.Stakeholder, &licensing.LicensingSourceValues.Account, nil
	case "Visual Studio Subscriber":
		return &licensing.AccountLicenseTypeValues.Advanced, &licensing.LicensingSourceValues.Msdn, nil
	default:
		return nil, nil, fmt.Errorf("invalid license_type '%s'; must be one of: express, stakeholder, Visual Studio Subscriber", input)
	}
}

// Entitlements always returns an empty slice for users.
func (o *userBuilder) Entitlements(_ context.Context, _ *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants always returns an empty slice for users since they don't have any entitlements.
func (o *userBuilder) Grants(_ context.Context, _ *v2.Resource, _ *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func parseIntoUserResource(userEntitlement *userentitlement.UserEntitlement) (*v2.Resource, error) {
	userStatus := v2.UserTrait_Status_STATUS_ENABLED

	if userEntitlement.AccessLevel.Status != nil {
		// status valid options: none, active, disabled, deleted, pending, expired, pendingDisabled
		status := *userEntitlement.AccessLevel.Status
		switch status {
		case accounts.AccountUserStatusValues.Disabled:
			userStatus = v2.UserTrait_Status_STATUS_DISABLED
		case accounts.AccountUserStatusValues.Deleted:
			userStatus = v2.UserTrait_Status_STATUS_DELETED
		}
	}

	var accountType v2.UserTrait_AccountType
	if userEntitlement.User.MetaType != nil && *userEntitlement.User.MetaType == "application" {
		accountType = v2.UserTrait_ACCOUNT_TYPE_SERVICE
	}

	profile := map[string]interface{}{
		"user_descriptor": *userEntitlement.User.Descriptor,
		"username":        *userEntitlement.User.DisplayName,
		"email":           *userEntitlement.User.MailAddress,
	}

	userTraits := []resource.UserTraitOption{
		resource.WithUserProfile(profile),
		resource.WithStatus(userStatus),
		resource.WithEmail(*userEntitlement.User.MailAddress, true),
		resource.WithUserLogin(*userEntitlement.User.DisplayName),
		resource.WithLastLogin(userEntitlement.LastAccessedDate.Time),
		resource.WithAccountType(accountType),
	}

	userResource, err := resource.NewUserResource(
		*userEntitlement.User.DisplayName,
		userResourceType,
		*userEntitlement.User.Descriptor,
		userTraits,
	)
	if err != nil {
		return nil, err
	}

	return userResource, nil
}

func newUserBuilder(c *client.AzureDevOpsClient) *userBuilder {
	return &userBuilder{
		resourceType: userResourceType,
		client:       c,
	}
}
