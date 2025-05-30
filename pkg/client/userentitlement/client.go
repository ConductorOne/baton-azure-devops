package userentitlement

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/google/uuid"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/licensingrule"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/webapi"
)

var ResourceAreaId, _ = uuid.Parse("68ddce18-2501-45f1-a17b-7931a9922690")

type Client interface {
	// [Preview API] Create a group entitlement with license rule, extension rule.
	AddGroupEntitlement(context.Context, AddGroupEntitlementArgs) (*GroupEntitlementOperationReference, error)
	// [Preview API] Add a member to a Group.
	AddMemberToGroup(context.Context, AddMemberToGroupArgs) error
	// [Preview API] Add a service principal, assign license and extensions and make them a member of a project group in an account.
	AddServicePrincipalEntitlement(context.Context, AddServicePrincipalEntitlementArgs) (*ServicePrincipalEntitlementsPostResponse, error)
	// [Preview API] Add a user, assign license and extensions and make them a member of a project group in an account.
	AddUserEntitlement(context.Context, AddUserEntitlementArgs) (*UserEntitlementsPostResponse, error)
	// [Preview API] Delete a group entitlement.
	DeleteGroupEntitlement(context.Context, DeleteGroupEntitlementArgs) (*GroupEntitlementOperationReference, error)
	// [Preview API] Delete a service principal from the account.
	DeleteServicePrincipalEntitlement(context.Context, DeleteServicePrincipalEntitlementArgs) error
	// [Preview API] Delete a user from the account.
	DeleteUserEntitlement(context.Context, DeleteUserEntitlementArgs) error
	// [Preview API] Get a group entitlement.
	GetGroupEntitlement(context.Context, GetGroupEntitlementArgs) (*GroupEntitlement, error)
	// [Preview API] Get the group entitlements for an account.
	GetGroupEntitlements(context.Context, GetGroupEntitlementsArgs) (*[]GroupEntitlement, error)
	// [Preview API] Get direct members of a Group.
	GetGroupMembers(context.Context, GetGroupMembersArgs) (*PagedGraphMemberList, error)
	// [Preview API] Get Service principal Entitlement for a service principal.
	GetServicePrincipalEntitlement(context.Context, GetServicePrincipalEntitlementArgs) (*ServicePrincipalEntitlement, error)
	// [Preview API] Get User Entitlement for a user.
	GetUserEntitlement(context.Context, GetUserEntitlementArgs) (*UserEntitlement, error)
	// [Preview API] Get summary of Licenses, Extension, Projects, Groups and their assignments in the collection.
	GetUsersSummary(context.Context, GetUsersSummaryArgs) (*UsersSummary, error)
	// [Preview API] Remove a member from a Group.
	RemoveMemberFromGroup(context.Context, RemoveMemberFromGroupArgs) error
	// [Preview API]
	SearchMemberEntitlements(context.Context, SearchMemberEntitlementsArgs) (*[]MemberEntitlement2, error)
	// [Preview API] Get a paged set of user entitlements matching the filter and sort criteria built with properties that match the select input.
	SearchUserEntitlements(context.Context, SearchUserEntitlementsArgs) (*PagedGraphMemberList, error)
	// [Preview API] Update entitlements (License Rule, Extensions Rule, Project memberships etc.) for a group.
	UpdateGroupEntitlement(context.Context, UpdateGroupEntitlementArgs) (*GroupEntitlementOperationReference, error)
	// [Preview API] Edit the entitlements (License, Extensions, Projects, Teams etc) for a service principal.
	UpdateServicePrincipalEntitlement(context.Context, UpdateServicePrincipalEntitlementArgs) (*ServicePrincipalEntitlementsPatchResponse, error)
	// [Preview API] Edit the entitlements (License, Extensions, Projects, Teams etc) for one or more service principals.
	UpdateServicePrincipalEntitlements(context.Context, UpdateServicePrincipalEntitlementsArgs) (*ServicePrincipalEntitlementOperationReference, error)
	// [Preview API] Edit the entitlements (License, Extensions, Projects, Teams etc) for a user.
	UpdateUserEntitlement(context.Context, UpdateUserEntitlementArgs) (*UserEntitlementsPatchResponse, error)
	// [Preview API] Edit the entitlements (License, Extensions, Projects, Teams etc) for one or more users.
	UpdateUserEntitlements(context.Context, UpdateUserEntitlementsArgs) (*UserEntitlementOperationReference, error)
}

type ClientImpl struct {
	Client azuredevops.Client
}

func NewClient(ctx context.Context, connection *azuredevops.Connection) (Client, error) {
	client, err := connection.GetClientByResourceAreaId(ctx, ResourceAreaId)
	if err != nil {
		return nil, err
	}
	return &ClientImpl{
		Client: *client,
	}, nil
}

// [Preview API] Create a group entitlement with license rule, extension rule.
func (client *ClientImpl) AddGroupEntitlement(ctx context.Context, args AddGroupEntitlementArgs) (*GroupEntitlementOperationReference, error) {
	if args.GroupEntitlement == nil {
		return nil, &azuredevops.ArgumentNilError{ArgumentName: "args.GroupEntitlement"}
	}
	queryParams := url.Values{}
	if args.RuleOption != nil {
		queryParams.Add("ruleOption", string(*args.RuleOption))
	}
	body, marshalErr := json.Marshal(*args.GroupEntitlement)
	if marshalErr != nil {
		return nil, marshalErr
	}
	locationId, _ := uuid.Parse("2280bffa-58a2-49da-822e-0764a1bb44f7")
	resp, err := client.Client.Send(ctx, http.MethodPost, locationId, "7.1-preview.1", nil, queryParams, bytes.NewReader(body), "application/json", "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var responseValue GroupEntitlementOperationReference
	err = client.Client.UnmarshalBody(resp, &responseValue)
	return &responseValue, err
}

// Arguments for the AddGroupEntitlement function.
type AddGroupEntitlementArgs struct {
	// (required) GroupEntitlement object specifying the License and Extensions
	// rules for the group. Based on these rules, the members of the group will
	// be assigned licenses and extensions. The Group Entitlement can also be
	// used to add the group to other project-level groups.
	GroupEntitlement *GroupEntitlement
	// (optional) RuleOption [ApplyGroupRule/TestApplyGroupRule] - specifies if the rules defined in group entitlement should be created and applied to it’s members (default option) or just be tested
	RuleOption *licensingrule.RuleOption
}

// [Preview API] Add a member to a Group.
func (client *ClientImpl) AddMemberToGroup(ctx context.Context, args AddMemberToGroupArgs) error {
	routeValues := make(map[string]string)
	if args.GroupId == nil {
		return &azuredevops.ArgumentNilError{ArgumentName: "args.GroupId"}
	}
	routeValues["groupId"] = args.GroupId.String()
	if args.MemberId == nil {
		return &azuredevops.ArgumentNilError{ArgumentName: "args.MemberId"}
	}
	routeValues["memberId"] = args.GroupId.String()

	locationId, _ := uuid.Parse("45a36e53-5286-4518-aa72-2d29f7acc5d8")
	resp, err := client.Client.Send(ctx, http.MethodPut, locationId, "7.1-preview.1", routeValues, nil, nil, "", "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// Arguments for the AddMemberToGroup function.
type AddMemberToGroupArgs struct {
	// (required) Id of the Group.
	GroupId *uuid.UUID
	// (required) Id of the member to add.
	MemberId *uuid.UUID
}

// [Preview API] Add a service principal, assign license and extensions and make them a member of a project group in an account.
func (client *ClientImpl) AddServicePrincipalEntitlement(ctx context.Context, args AddServicePrincipalEntitlementArgs) (*ServicePrincipalEntitlementsPostResponse, error) {
	if args.ServicePrincipalEntitlement == nil {
		return nil, &azuredevops.ArgumentNilError{ArgumentName: "args.ServicePrincipalEntitlement"}
	}
	body, marshalErr := json.Marshal(*args.ServicePrincipalEntitlement)
	if marshalErr != nil {
		return nil, marshalErr
	}
	locationId, _ := uuid.Parse("f03dbf50-80f8-41b7-8ca2-65b6a178caba")
	resp, err := client.Client.Send(ctx, http.MethodPost, locationId, "7.1-preview.1", nil, nil, bytes.NewReader(body), "application/json", "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var responseValue ServicePrincipalEntitlementsPostResponse
	err = client.Client.UnmarshalBody(resp, &responseValue)
	return &responseValue, err
}

// Arguments for the AddServicePrincipalEntitlement function.
type AddServicePrincipalEntitlementArgs struct {
	// (required) ServicePrincipalEntitlement object specifying License, Extensions and Project/Team groups the service principal should be added to.
	ServicePrincipalEntitlement *ServicePrincipalEntitlement
}

// [Preview API] Add a user, assign license and extensions and make them a member of a project group in an account.
func (client *ClientImpl) AddUserEntitlement(ctx context.Context, args AddUserEntitlementArgs) (*UserEntitlementsPostResponse, error) {
	if args.UserEntitlement == nil {
		return nil, &azuredevops.ArgumentNilError{ArgumentName: "args.UserEntitlement"}
	}
	body, marshalErr := json.Marshal(*args.UserEntitlement)
	if marshalErr != nil {
		return nil, marshalErr
	}
	locationId, _ := uuid.Parse("387f832c-dbf2-4643-88e9-c1aa94dbb737")
	resp, err := client.Client.Send(ctx, http.MethodPost, locationId, "7.1-preview.3", nil, nil, bytes.NewReader(body), "application/json", "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var responseValue UserEntitlementsPostResponse
	err = client.Client.UnmarshalBody(resp, &responseValue)
	return &responseValue, err
}

// Arguments for the AddUserEntitlement function.
type AddUserEntitlementArgs struct {
	// (required) UserEntitlement object specifying License, Extensions and Project/Team groups the user should be added to.
	UserEntitlement *UserEntitlement
}

// [Preview API] Delete a group entitlement.
func (client *ClientImpl) DeleteGroupEntitlement(ctx context.Context, args DeleteGroupEntitlementArgs) (*GroupEntitlementOperationReference, error) {
	routeValues := make(map[string]string)
	if args.GroupId == nil {
		return nil, &azuredevops.ArgumentNilError{ArgumentName: "args.GroupId"}
	}
	routeValues["groupId"] = args.GroupId.String()

	queryParams := url.Values{}
	if args.RuleOption != nil {
		queryParams.Add("ruleOption", string(*args.RuleOption))
	}
	if args.RemoveGroupMembership != nil {
		queryParams.Add("removeGroupMembership", strconv.FormatBool(*args.RemoveGroupMembership))
	}
	locationId, _ := uuid.Parse("2280bffa-58a2-49da-822e-0764a1bb44f7")
	resp, err := client.Client.Send(ctx, http.MethodDelete, locationId, "7.1-preview.1", routeValues, queryParams, nil, "", "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var responseValue GroupEntitlementOperationReference
	err = client.Client.UnmarshalBody(resp, &responseValue)
	return &responseValue, err
}

// Arguments for the DeleteGroupEntitlement function.
type DeleteGroupEntitlementArgs struct {
	// (required) ID of the group to delete.
	GroupId *uuid.UUID
	// (optional) RuleOption [ApplyGroupRule/TestApplyGroupRule] -
	// Specifies whether the rules defined in the group entitlement
	// should be applied to its members (default option) or just tested.
	//
	// - ApplyGroupRule: The changes are applied to the group's members.
	// - TestApplyGroupRule: The rules are tested without applying changes.
	RuleOption *licensingrule.RuleOption
	// (optional) Optional parameter that specifies whether the group with the given ID should be removed from all other groups
	RemoveGroupMembership *bool
}

// [Preview API] Delete a service principal from the account.
func (client *ClientImpl) DeleteServicePrincipalEntitlement(ctx context.Context, args DeleteServicePrincipalEntitlementArgs) error {
	routeValues := make(map[string]string)
	if args.ServicePrincipalId == nil {
		return &azuredevops.ArgumentNilError{ArgumentName: "args.ServicePrincipalId"}
	}
	routeValues["servicePrincipalId"] = args.ServicePrincipalId.String()

	locationId, _ := uuid.Parse("1d491a66-190b-43ae-86b8-9c2688c55186")
	resp, err := client.Client.Send(ctx, http.MethodDelete, locationId, "7.1-preview.1", routeValues, nil, nil, "", "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// Arguments for the DeleteServicePrincipalEntitlement function.
type DeleteServicePrincipalEntitlementArgs struct {
	// (required) ID of the service principal.
	ServicePrincipalId *uuid.UUID
}

// [Preview API] Delete a user from the account.
func (client *ClientImpl) DeleteUserEntitlement(ctx context.Context, args DeleteUserEntitlementArgs) error {
	routeValues := make(map[string]string)
	if args.UserId == nil {
		return &azuredevops.ArgumentNilError{ArgumentName: "args.UserId"}
	}
	routeValues["userId"] = args.UserId.String()

	locationId, _ := uuid.Parse("8480c6eb-ce60-47e9-88df-eca3c801638b")
	resp, err := client.Client.Send(ctx, http.MethodDelete, locationId, "7.1-preview.3", routeValues, nil, nil, "", "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// Arguments for the DeleteUserEntitlement function.
type DeleteUserEntitlementArgs struct {
	// (required) ID of the user.
	UserId *uuid.UUID
}

// [Preview API] Get a group entitlement.
func (client *ClientImpl) GetGroupEntitlement(ctx context.Context, args GetGroupEntitlementArgs) (*GroupEntitlement, error) {
	routeValues := make(map[string]string)
	if args.GroupId == nil {
		return nil, &azuredevops.ArgumentNilError{ArgumentName: "args.GroupId"}
	}
	routeValues["groupId"] = args.GroupId.String()

	locationId, _ := uuid.Parse("2280bffa-58a2-49da-822e-0764a1bb44f7")
	resp, err := client.Client.Send(ctx, http.MethodGet, locationId, "7.1-preview.1", routeValues, nil, nil, "", "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var responseValue GroupEntitlement
	err = client.Client.UnmarshalBody(resp, &responseValue)
	return &responseValue, err
}

// Arguments for the GetGroupEntitlement function.
type GetGroupEntitlementArgs struct {
	// (required) ID of the group.
	GroupId *uuid.UUID
}

// [Preview API] Get the group entitlements for an account.
func (client *ClientImpl) GetGroupEntitlements(ctx context.Context, args GetGroupEntitlementsArgs) (*[]GroupEntitlement, error) {
	locationId, _ := uuid.Parse("9bce1f43-2629-419f-8f6c-7503be58a4f3")
	resp, err := client.Client.Send(ctx, http.MethodGet, locationId, "7.1-preview.1", nil, nil, nil, "", "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var responseValue []GroupEntitlement
	err = client.Client.UnmarshalCollectionBody(resp, &responseValue)
	return &responseValue, err
}

// Arguments for the GetGroupEntitlements function.
type GetGroupEntitlementsArgs struct {
}

// [Preview API] Get direct members of a Group.
func (client *ClientImpl) GetGroupMembers(ctx context.Context, args GetGroupMembersArgs) (*PagedGraphMemberList, error) {
	routeValues := make(map[string]string)
	if args.GroupId == nil {
		return nil, &azuredevops.ArgumentNilError{ArgumentName: "args.GroupId"}
	}
	routeValues["groupId"] = args.GroupId.String()

	queryParams := url.Values{}
	if args.MaxResults != nil {
		queryParams.Add("maxResults", strconv.Itoa(*args.MaxResults))
	}
	if args.PagingToken != nil {
		queryParams.Add("pagingToken", *args.PagingToken)
	}
	locationId, _ := uuid.Parse("45a36e53-5286-4518-aa72-2d29f7acc5d8")
	resp, err := client.Client.Send(ctx, http.MethodGet, locationId, "7.1-preview.1", routeValues, queryParams, nil, "", "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var responseValue PagedGraphMemberList
	err = client.Client.UnmarshalBody(resp, &responseValue)
	return &responseValue, err
}

// Arguments for the GetGroupMembers function.
type GetGroupMembersArgs struct {
	// (required) Id of the Group.
	GroupId *uuid.UUID
	// (optional) Maximum number of results to retrieve.
	MaxResults *int
	// (optional) Paging Token from the previous page fetched. If the 'pagingToken' is null, the results would be fetched from the beginning of the Members List.
	PagingToken *string
}

// [Preview API] Get Service principal Entitlement for a service principal.
func (client *ClientImpl) GetServicePrincipalEntitlement(ctx context.Context, args GetServicePrincipalEntitlementArgs) (*ServicePrincipalEntitlement, error) {
	routeValues := make(map[string]string)
	if args.ServicePrincipalId == nil {
		return nil, &azuredevops.ArgumentNilError{ArgumentName: "args.ServicePrincipalId"}
	}
	routeValues["servicePrincipalId"] = args.ServicePrincipalId.String()

	locationId, _ := uuid.Parse("1d491a66-190b-43ae-86b8-9c2688c55186")
	resp, err := client.Client.Send(ctx, http.MethodGet, locationId, "7.1-preview.1", routeValues, nil, nil, "", "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var responseValue ServicePrincipalEntitlement
	err = client.Client.UnmarshalBody(resp, &responseValue)
	return &responseValue, err
}

// Arguments for the GetServicePrincipalEntitlement function.
type GetServicePrincipalEntitlementArgs struct {
	// (required) ID of the service principal.
	ServicePrincipalId *uuid.UUID
}

// [Preview API] Get User Entitlement for a user.
func (client *ClientImpl) GetUserEntitlement(ctx context.Context, args GetUserEntitlementArgs) (*UserEntitlement, error) {
	routeValues := make(map[string]string)
	if args.UserId == nil {
		return nil, &azuredevops.ArgumentNilError{ArgumentName: "args.UserId"}
	}
	routeValues["userId"] = args.UserId.String()

	locationId, _ := uuid.Parse("8480c6eb-ce60-47e9-88df-eca3c801638b")
	resp, err := client.Client.Send(ctx, http.MethodGet, locationId, "7.1-preview.3", routeValues, nil, nil, "", "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var responseValue UserEntitlement
	err = client.Client.UnmarshalBody(resp, &responseValue)
	return &responseValue, err
}

// Arguments for the GetUserEntitlement function.
type GetUserEntitlementArgs struct {
	// (required) ID of the user.
	UserId *uuid.UUID
}

// [Preview API] Get summary of Licenses, Extension, Projects, Groups and their assignments in the collection.
func (client *ClientImpl) GetUsersSummary(ctx context.Context, args GetUsersSummaryArgs) (*UsersSummary, error) {
	queryParams := url.Values{}
	if args.Select != nil {
		queryParams.Add("select", *args.Select)
	}
	locationId, _ := uuid.Parse("5ae55b13-c9dd-49d1-957e-6e76c152e3d9")
	resp, err := client.Client.Send(ctx, http.MethodGet, locationId, "7.1-preview.1", nil, queryParams, nil, "", "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var responseValue UsersSummary
	err = client.Client.UnmarshalBody(resp, &responseValue)
	return &responseValue, err
}

// Arguments for the GetUsersSummary function.
type GetUsersSummaryArgs struct {
	// (optional) Comma (",") separated list of properties to select. Supported property names are {AccessLevels, Licenses, Projects, Groups}.
	Select *string
}

// [Preview API] Remove a member from a Group.
func (client *ClientImpl) RemoveMemberFromGroup(ctx context.Context, args RemoveMemberFromGroupArgs) error {
	routeValues := make(map[string]string)
	if args.GroupId == nil {
		return &azuredevops.ArgumentNilError{ArgumentName: "args.GroupId"}
	}
	routeValues["groupId"] = args.GroupId.String()
	if args.MemberId == nil {
		return &azuredevops.ArgumentNilError{ArgumentName: "args.MemberId"}
	}
	routeValues["memberId"] = args.MemberId.String()

	locationId, _ := uuid.Parse("45a36e53-5286-4518-aa72-2d29f7acc5d8")
	resp, err := client.Client.Send(ctx, http.MethodDelete, locationId, "7.1-preview.1", routeValues, nil, nil, "", "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// Arguments for the RemoveMemberFromGroup function.
type RemoveMemberFromGroupArgs struct {
	// (required) Id of the group.
	GroupId *uuid.UUID
	// (required) Id of the member to remove.
	MemberId *uuid.UUID
}

// [Preview API].
func (client *ClientImpl) SearchMemberEntitlements(ctx context.Context, args SearchMemberEntitlementsArgs) (*[]MemberEntitlement2, error) {
	queryParams := url.Values{}
	if args.ContinuationToken != nil {
		queryParams.Add("continuationToken", *args.ContinuationToken)
	}
	if args.Select != nil {
		queryParams.Add("select", string(*args.Select))
	}
	if args.Filter != nil {
		queryParams.Add("$filter", *args.Filter)
	}
	if args.OrderBy != nil {
		queryParams.Add("$orderBy", *args.OrderBy)
	}
	locationId, _ := uuid.Parse("1e8cabfb-1fda-461e-860f-eeeae54d06bb")
	resp, err := client.Client.Send(ctx, http.MethodGet, locationId, "7.1-preview.2", nil, queryParams, nil, "", "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var responseValue []MemberEntitlement2
	err = client.Client.UnmarshalBody(resp, &responseValue)
	return &responseValue, err
}

// Arguments for the SearchMemberEntitlements function.
type SearchMemberEntitlementsArgs struct {
	// (optional)
	ContinuationToken *string
	// (optional)
	Select *UserEntitlementProperty
	// (optional)
	Filter *string
	// (optional)
	OrderBy *string
}

// [Preview API] Get a paged set of user entitlements matching the filter and sort criteria built with properties that match the select input.
func (client *ClientImpl) SearchUserEntitlements(ctx context.Context, args SearchUserEntitlementsArgs) (*PagedGraphMemberList, error) {
	queryParams := url.Values{}
	if args.ContinuationToken != nil {
		queryParams.Add("continuationToken", *args.ContinuationToken)
	}
	if args.Select != nil {
		queryParams.Add("select", string(*args.Select))
	}
	if args.Filter != nil {
		queryParams.Add("$filter", *args.Filter)
	}
	if args.OrderBy != nil {
		queryParams.Add("$orderBy", *args.OrderBy)
	}
	locationId, _ := uuid.Parse("387f832c-dbf2-4643-88e9-c1aa94dbb737")
	resp, err := client.Client.Send(ctx, http.MethodGet, locationId, "7.1-preview.3", nil, queryParams, nil, "", "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var responseValue PagedGraphMemberList
	err = client.Client.UnmarshalBody(resp, &responseValue)
	return &responseValue, err
}

// Arguments for the SearchUserEntitlements function.
type SearchUserEntitlementsArgs struct {
	// (optional) Continuation token for getting the next page of data set. If null is passed, gets the first page.
	ContinuationToken *string
	// (optional) Comma (",") separated list of properties to select in the result entitlements. names of the properties are - 'Projects, 'Extensions' and 'Grouprules'.
	Select *UserEntitlementProperty
	// (optional) Equality operators relating to searching user entitlements,
	// separated by 'and' clauses. Valid filters include: licenseId,
	// licenseStatus, userType, and name.
	//
	// licenseId: filters based on license assignment using license names.
	// Example: licenseId eq 'Account-Stakeholder' or licenseId eq 'Account-Express'.
	//
	// licenseStatus: filters based on license status. Currently, only supports
	// 'disabled'. Example: licenseStatus eq 'Disabled'. To get disabled basic
	// licenses, you would pass (licenseId eq 'Account-Express' and
	// licenseStatus eq 'Disabled').
	//
	// userType: filters based on identity type. Supported types are 'member'
	// or 'guest'. Example: userType eq 'member'.
	//
	// name: filters on the user's display name or email, containing the given input.
	// Example: get all users with "test" in email or display name.
	// Example query: name eq 'test'.
	//
	// A valid query could be:
	// (licenseId eq 'Account-Stakeholder' or
	//  (licenseId eq 'Account-Express' and licenseStatus eq 'Disabled')) and
	//  name eq 'test' and userType eq 'guest'.
	Filter *string
	// (optional) PropertyName and Order (separated by a space) to sort on
	// (e.g. lastAccessed desc). Order defaults to ascending. Valid properties
	// to order by are dateCreated, lastAccessed, and name.
	OrderBy *string
}

// [Preview API] Update entitlements (License Rule, Extensions Rule, Project memberships etc.) for a group.
func (client *ClientImpl) UpdateGroupEntitlement(ctx context.Context, args UpdateGroupEntitlementArgs) (*GroupEntitlementOperationReference, error) {
	if args.Document == nil {
		return nil, &azuredevops.ArgumentNilError{ArgumentName: "args.Document"}
	}
	routeValues := make(map[string]string)
	if args.GroupId == nil {
		return nil, &azuredevops.ArgumentNilError{ArgumentName: "args.GroupId"}
	}
	routeValues["groupId"] = args.GroupId.String()

	queryParams := url.Values{}
	if args.RuleOption != nil {
		queryParams.Add("ruleOption", string(*args.RuleOption))
	}
	body, marshalErr := json.Marshal(*args.Document)
	if marshalErr != nil {
		return nil, marshalErr
	}
	locationId, _ := uuid.Parse("2280bffa-58a2-49da-822e-0764a1bb44f7")
	resp, err := client.Client.Send(ctx, http.MethodPatch, locationId, "7.1-preview.1", routeValues, queryParams, bytes.NewReader(body), "application/json-patch+json", "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var responseValue GroupEntitlementOperationReference
	err = client.Client.UnmarshalBody(resp, &responseValue)
	return &responseValue, err
}

// Arguments for the UpdateGroupEntitlement function.
type UpdateGroupEntitlementArgs struct {
	// (required) JsonPatchDocument containing the operations to perform on the group.
	Document *[]webapi.JsonPatchOperation
	// (required) ID of the group.
	GroupId *uuid.UUID
	// (optional) RuleOption [ApplyGroupRule/TestApplyGroupRule] - specifies if the rules
	// defined in group entitlement should be updated and the changes are applied to its
	// members (default option) or just be tested.
	RuleOption *licensingrule.RuleOption
}

// [Preview API] Edit the entitlements (License, Extensions, Projects, Teams etc) for a service principal.
func (client *ClientImpl) UpdateServicePrincipalEntitlement(ctx context.Context, args UpdateServicePrincipalEntitlementArgs) (*ServicePrincipalEntitlementsPatchResponse, error) {
	if args.Document == nil {
		return nil, &azuredevops.ArgumentNilError{ArgumentName: "args.Document"}
	}
	routeValues := make(map[string]string)
	if args.ServicePrincipalId == nil {
		return nil, &azuredevops.ArgumentNilError{ArgumentName: "args.ServicePrincipalId"}
	}
	routeValues["servicePrincipalId"] = args.ServicePrincipalId.String()

	body, marshalErr := json.Marshal(*args.Document)
	if marshalErr != nil {
		return nil, marshalErr
	}
	locationId, _ := uuid.Parse("1d491a66-190b-43ae-86b8-9c2688c55186")
	resp, err := client.Client.Send(ctx, http.MethodPatch, locationId, "7.1-preview.1", routeValues, nil, bytes.NewReader(body), "application/json-patch+json", "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var responseValue ServicePrincipalEntitlementsPatchResponse
	err = client.Client.UnmarshalBody(resp, &responseValue)
	return &responseValue, err
}

// Arguments for the UpdateServicePrincipalEntitlement function.
type UpdateServicePrincipalEntitlementArgs struct {
	// (required) JsonPatchDocument containing the operations to perform on the service principal.
	Document *[]webapi.JsonPatchOperation
	// (required) ID of the service principal.
	ServicePrincipalId *uuid.UUID
}

// [Preview API] Edit the entitlements (License, Extensions, Projects, Teams etc) for one or more service principals.
func (client *ClientImpl) UpdateServicePrincipalEntitlements(ctx context.Context, args UpdateServicePrincipalEntitlementsArgs) (*ServicePrincipalEntitlementOperationReference, error) {
	if args.Document == nil {
		return nil, &azuredevops.ArgumentNilError{ArgumentName: "args.Document"}
	}
	body, marshalErr := json.Marshal(*args.Document)
	if marshalErr != nil {
		return nil, marshalErr
	}
	locationId, _ := uuid.Parse("f03dbf50-80f8-41b7-8ca2-65b6a178caba")
	resp, err := client.Client.Send(ctx, http.MethodPatch, locationId, "7.1-preview.1", nil, nil, bytes.NewReader(body), "application/json-patch+json", "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var responseValue ServicePrincipalEntitlementOperationReference
	err = client.Client.UnmarshalBody(resp, &responseValue)
	return &responseValue, err
}

// Arguments for the UpdateServicePrincipalEntitlements function.
type UpdateServicePrincipalEntitlementsArgs struct {
	// (required) JsonPatchDocument containing the operations to perform.
	Document *[]webapi.JsonPatchOperation
}

// [Preview API] Edit the entitlements (License, Extensions, Projects, Teams etc) for a user.
func (client *ClientImpl) UpdateUserEntitlement(ctx context.Context, args UpdateUserEntitlementArgs) (*UserEntitlementsPatchResponse, error) {
	if args.Document == nil {
		return nil, &azuredevops.ArgumentNilError{ArgumentName: "args.Document"}
	}
	routeValues := make(map[string]string)
	if args.UserId == nil {
		return nil, &azuredevops.ArgumentNilError{ArgumentName: "args.UserId"}
	}
	routeValues["userId"] = args.UserId.String()

	body, marshalErr := json.Marshal(*args.Document)
	if marshalErr != nil {
		return nil, marshalErr
	}
	locationId, _ := uuid.Parse("8480c6eb-ce60-47e9-88df-eca3c801638b")
	resp, err := client.Client.Send(ctx, http.MethodPatch, locationId, "7.1-preview.3", routeValues, nil, bytes.NewReader(body), "application/json-patch+json", "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var responseValue UserEntitlementsPatchResponse
	err = client.Client.UnmarshalBody(resp, &responseValue)
	return &responseValue, err
}

// Arguments for the UpdateUserEntitlement function.
type UpdateUserEntitlementArgs struct {
	// (required) JsonPatchDocument containing the operations to perform on the user.
	Document *[]webapi.JsonPatchOperation
	// (required) ID of the user.
	UserId *uuid.UUID
}

// [Preview API] Edit the entitlements (License, Extensions, Projects, Teams etc) for one or more users.
func (client *ClientImpl) UpdateUserEntitlements(ctx context.Context, args UpdateUserEntitlementsArgs) (*UserEntitlementOperationReference, error) {
	if args.Document == nil {
		return nil, &azuredevops.ArgumentNilError{ArgumentName: "args.Document"}
	}
	queryParams := url.Values{}
	if args.DoNotSendInviteForNewUsers != nil {
		queryParams.Add("doNotSendInviteForNewUsers", strconv.FormatBool(*args.DoNotSendInviteForNewUsers))
	}
	body, marshalErr := json.Marshal(*args.Document)
	if marshalErr != nil {
		return nil, marshalErr
	}
	locationId, _ := uuid.Parse("387f832c-dbf2-4643-88e9-c1aa94dbb737")
	resp, err := client.Client.Send(ctx, http.MethodPatch, locationId, "7.1-preview.3", nil, queryParams, bytes.NewReader(body), "application/json-patch+json", "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var responseValue UserEntitlementOperationReference
	err = client.Client.UnmarshalBody(resp, &responseValue)
	return &responseValue, err
}

// Arguments for the UpdateUserEntitlements function.
type UpdateUserEntitlementsArgs struct {
	// (required) JsonPatchDocument containing the operations to perform.
	Document *[]webapi.JsonPatchOperation
	// (optional) Whether to send email invites to new users or not
	DoNotSendInviteForNewUsers *bool
}
