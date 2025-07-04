package userentitlement

import (
	"github.com/google/uuid"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/commerce"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/graph"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/licensing"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/licensingrule"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/operations"
)

type BaseOperationResult struct {
	// List of error codes paired with their corresponding error messages
	Errors *[]azuredevops.KeyValuePair `json:"errors,omitempty"`
	// Success status of the operation
	IsSuccess *bool `json:"isSuccess,omitempty"`
}

type EntitlementBase struct {
	// Member's access level denoted by a license.
	AccessLevel *licensing.AccessLevel `json:"accessLevel,omitempty"`
	// [Readonly] Date the member was added to the collection.
	DateCreated *azuredevops.Time `json:"dateCreated,omitempty"`
	// [Readonly] GroupEntitlements that this member belongs to.
	GroupAssignments *[]GroupEntitlement `json:"groupAssignments,omitempty"`
	// The unique identifier which matches the Id of the Identity associated with the GraphMember.
	Id *uuid.UUID `json:"id,omitempty"`
	// [Readonly] Date the member last accessed the collection.
	LastAccessedDate *azuredevops.Time `json:"lastAccessedDate,omitempty"`
	// Relation between a project and the member's effective permissions in that project.
	ProjectEntitlements *[]ProjectEntitlement `json:"projectEntitlements,omitempty"`
}

type EntitlementOperationResultBase struct {
	// List of error codes paired with their corresponding error messages.
	Errors *[]azuredevops.KeyValuePair `json:"errors,omitempty"`
	// Success status of the operation.
	IsSuccess *bool `json:"isSuccess,omitempty"`
	// Resulting entitlement property. For specific implementations, see also:
	//   T:Microsoft.VisualStudio.Services.MemberEntitlementManagement.WebApi.ServicePrincipalEntitlementOperationResult
	//   T:Microsoft.VisualStudio.Services.MemberEntitlementManagement.WebApi.UserEntitlementOperationResult

	Result interface{} `json:"result,omitempty"`
}

// An extension assigned to a user.
type Extension struct {
	// Assignment source for this extension. I.e. explicitly assigned or from a group rule.
	AssignmentSource *licensing.AssignmentSource `json:"assignmentSource,omitempty"`
	// Gallery Id of the Extension.
	Id *string `json:"id,omitempty"`
	// Friendly name of this extension.
	Name *string `json:"name,omitempty"`
	// Source of this extension assignment. Ex: msdn, account, none, etc.
	Source *licensing.LicensingSource `json:"source,omitempty"`
}

// Summary of Extensions in the organization.
type ExtensionSummaryData struct {
	// Count of Licenses already assigned.
	Assigned *int `json:"assigned,omitempty"`
	// Available Count.
	Available *int `json:"available,omitempty"`
	// Quantity
	IncludedQuantity *int `json:"includedQuantity,omitempty"`
	// Total Count.
	Total *int `json:"total,omitempty"`
	// Count of Extension Licenses assigned to users through msdn.
	AssignedThroughSubscription *int `json:"assignedThroughSubscription,omitempty"`
	// Gallery Id of the Extension
	ExtensionId *string `json:"extensionId,omitempty"`
	// Friendly name of this extension
	ExtensionName *string `json:"extensionName,omitempty"`
	// Whether its a Trial Version.
	IsTrialVersion *bool `json:"isTrialVersion,omitempty"`
	// Minimum License Required for the Extension.
	MinimumLicenseRequired *commerce.MinimumRequiredServiceLevel `json:"minimumLicenseRequired,omitempty"`
	// Days remaining for the Trial to expire.
	RemainingTrialDays *int `json:"remainingTrialDays,omitempty"`
	// Date on which the Trial expires.
	TrialExpiryDate *azuredevops.Time `json:"trialExpiryDate,omitempty"`
}

// Project Group (e.g. Contributor, Reader etc.)
type Group struct {
	// Display Name of the Group
	DisplayName *string `json:"displayName,omitempty"`
	// Group Type
	GroupType *GroupType `json:"groupType,omitempty"`
}

// A group entity with additional properties including its license, extensions, and project membership.
type GroupEntitlement struct {
	// Deprecated:
	ExtensionRules *[]Extension `json:"extensionRules,omitempty"`
	// Member reference.
	Group *graph.GraphGroup `json:"group,omitempty"`
	// The unique identifier which matches the Id of the GraphMember.
	Id *uuid.UUID `json:"id,omitempty"`
	// [Readonly] The last time the group licensing rule was executed (regardless of whether any changes were made).
	LastExecuted *azuredevops.Time `json:"lastExecuted,omitempty"`
	// License Rule.
	LicenseRule *licensing.AccessLevel `json:"licenseRule,omitempty"`
	// Group members. Only used when creating a new group.
	Members *[]UserEntitlement `json:"members,omitempty"`
	// Relation between a project and the member's effective permissions in that project.
	ProjectEntitlements *[]ProjectEntitlement `json:"projectEntitlements,omitempty"`
	// The status of the group rule.
	Status *licensingrule.GroupLicensingRuleStatus `json:"status,omitempty"`
}

type GroupEntitlementOperationReference struct {
	// Unique identifier for the operation.
	Id *uuid.UUID `json:"id,omitempty"`
	// Unique identifier for the plugin.
	PluginId *uuid.UUID `json:"pluginId,omitempty"`
	// The current status of the operation.
	Status *operations.OperationStatus `json:"status,omitempty"`
	// URL to get the full operation object.
	Url *string `json:"url,omitempty"`
	// Operation completed with success or failure.
	Completed *bool `json:"completed,omitempty"`
	// True if all operations were successful.
	HaveResultsSucceeded *bool `json:"haveResultsSucceeded,omitempty"`
	// List of results for each operation.
	Results *[]GroupOperationResult `json:"results,omitempty"`
}

type GroupOperationResult struct {
	// List of error codes paired with their corresponding error messages
	Errors *[]azuredevops.KeyValuePair `json:"errors,omitempty"`
	// Success status of the operation
	IsSuccess *bool `json:"isSuccess,omitempty"`
	// Identifier of the Group being acted upon
	GroupId *uuid.UUID `json:"groupId,omitempty"`
	// Result of the Groupentitlement after the operation
	Result *GroupEntitlement `json:"result,omitempty"`
}

// Group option to add a user to.
type GroupOption struct {
	// Access Level
	AccessLevel *licensing.AccessLevel `json:"accessLevel,omitempty"`
	// Group
	Group *Group `json:"group,omitempty"`
}

// Used when adding users to a project. Each GroupType maps to a well-known group. The lowest GroupType should always be ProjectStakeholder.
type GroupType string

type groupTypeValuesType struct {
	ProjectStakeholder   GroupType
	ProjectReader        GroupType
	ProjectContributor   GroupType
	ProjectAdministrator GroupType
	Custom               GroupType
}

var GroupTypeValues = groupTypeValuesType{
	ProjectStakeholder:   "projectStakeholder",
	ProjectReader:        "projectReader",
	ProjectContributor:   "projectContributor",
	ProjectAdministrator: "projectAdministrator",
	Custom:               "custom",
}

// Summary of Licenses in the organization.
type LicenseSummaryData struct {
	// Count of Licenses already assigned.
	Assigned *int `json:"assigned,omitempty"`
	// Available Count.
	Available *int `json:"available,omitempty"`
	// Quantity
	IncludedQuantity *int `json:"includedQuantity,omitempty"`
	// Total Count.
	Total *int `json:"total,omitempty"`
	// Type of Account License.
	AccountLicenseType *licensing.AccountLicenseType `json:"accountLicenseType,omitempty"`
	// Count of Disabled Licenses.
	Disabled *int `json:"disabled,omitempty"`
	// Designates if this license quantity can be changed through purchase
	IsPurchasable *bool `json:"isPurchasable,omitempty"`
	// Name of the License.
	LicenseName *string `json:"licenseName,omitempty"`
	// Type of MSDN License.
	MsdnLicenseType *licensing.MsdnLicenseType `json:"msdnLicenseType,omitempty"`
	// Specifies the date when billing will charge for paid licenses
	NextBillingDate *azuredevops.Time `json:"nextBillingDate,omitempty"`
	// Source of the License.
	Source *licensing.LicensingSource `json:"source,omitempty"`
	// Total license count after next billing cycle
	TotalAfterNextBillingDate *int `json:"totalAfterNextBillingDate,omitempty"`
}

// Deprecated: Use UserEntitlement instead.
type MemberEntitlement struct {
	// Member's access level denoted by a license.
	AccessLevel *licensing.AccessLevel `json:"accessLevel,omitempty"`
	// [Readonly] Date the member was added to the collection.
	DateCreated *azuredevops.Time `json:"dateCreated,omitempty"`
	// [Readonly] GroupEntitlements that this member belongs to.
	GroupAssignments *[]GroupEntitlement `json:"groupAssignments,omitempty"`
	// The unique identifier which matches the Id of the Identity associated with the GraphMember.
	Id *uuid.UUID `json:"id,omitempty"`
	// [Readonly] Date the member last accessed the collection.
	LastAccessedDate *azuredevops.Time `json:"lastAccessedDate,omitempty"`
	// Relation between a project and the member's effective permissions in that project.
	ProjectEntitlements *[]ProjectEntitlement `json:"projectEntitlements,omitempty"`
	// User's extensions.
	Extensions *[]Extension `json:"extensions,omitempty"`
	// User reference.
	User *graph.GraphUser `json:"user,omitempty"`
	// Member reference
	Member *graph.GraphMember `json:"member,omitempty"`
}

// An AAD member entity.
type MemberEntitlement2 struct {
	// Member's access level denoted by a license.
	AccessLevel *licensing.AccessLevel `json:"accessLevel,omitempty"`
	// [Readonly] Date the member was added to the collection.
	DateCreated *azuredevops.Time `json:"dateCreated,omitempty"`
	// [Readonly] GroupEntitlements that this member belongs to.
	GroupAssignments *[]GroupEntitlement `json:"groupAssignments,omitempty"`
	// The unique identifier which matches the Id of the Identity associated with the GraphMember.
	Id *uuid.UUID `json:"id,omitempty"`
	// [Readonly] Date the member last accessed the collection.
	LastAccessedDate *azuredevops.Time `json:"lastAccessedDate,omitempty"`
	// Relation between a project and the member's effective permissions in that project.
	ProjectEntitlements *[]ProjectEntitlement `json:"projectEntitlements,omitempty"`
	Member              *graph.AadGraphMember `json:"member,omitempty"`
}

type MemberEntitlement2OperationReference struct {
	// Unique identifier for the operation.
	Id *uuid.UUID `json:"id,omitempty"`
	// Unique identifier for the plugin.
	PluginId *uuid.UUID `json:"pluginId,omitempty"`
	// The current status of the operation.
	Status *operations.OperationStatus `json:"status,omitempty"`
	// URL to get the full operation object.
	Url *string `json:"url,omitempty"`
	// Operation completed with success or failure.
	Completed *bool `json:"completed,omitempty"`
	// True if all operations were successful.
	HaveResultsSucceeded *bool `json:"haveResultsSucceeded,omitempty"`
	// List of results for each operation.
	Results *[]MemberEntitlement2OperationResult `json:"results,omitempty"`
}

type MemberEntitlement2OperationResult struct {
	// List of error codes paired with their corresponding error messages.
	Errors *[]azuredevops.KeyValuePair `json:"errors,omitempty"`
	// Success status of the operation.
	IsSuccess *bool `json:"isSuccess,omitempty"`
	// Resulting entitlement property. For specific implementations, see also:
	//   T:Microsoft.VisualStudio.Services.MemberEntitlementManagement.WebApi.ServicePrincipalEntitlementOperationResult
	//   T:Microsoft.VisualStudio.Services.MemberEntitlementManagement.WebApi.UserEntitlementOperationResult

	Result interface{} `json:"result,omitempty"`
	// Identifier of the Member being acted upon.
	MemberId *uuid.UUID `json:"memberId,omitempty"`
}

type MemberEntitlement2PatchResponse struct {
	// True if all operations were successful.
	IsSuccess *bool `json:"isSuccess,omitempty"`
	// Result of the member entitlement after the operations. have been applied
	MemberEntitlement *MemberEntitlement2 `json:"memberEntitlement,omitempty"`
	// List of results for each operation
	OperationResults *[]MemberEntitlement2OperationResult `json:"operationResults,omitempty"`
}

type MemberEntitlement2PostResponse struct {
	// True if all operations were successful.
	IsSuccess *bool `json:"isSuccess,omitempty"`
	// Result of the member entitlement after the operations. have been applied
	MemberEntitlement *MemberEntitlement2 `json:"memberEntitlement,omitempty"`
	// Operation result
	OperationResult *MemberEntitlement2OperationResult `json:"operationResult,omitempty"`
}

type MemberEntitlement2ResponseBase struct {
	// True if all operations were successful.
	IsSuccess *bool `json:"isSuccess,omitempty"`
	// Result of the member entitlement after the operations. have been applied
	MemberEntitlement *MemberEntitlement2 `json:"memberEntitlement,omitempty"`
}

type MemberEntitlementOperationReference struct {
	// Unique identifier for the operation.
	Id *uuid.UUID `json:"id,omitempty"`
	// Unique identifier for the plugin.
	PluginId *uuid.UUID `json:"pluginId,omitempty"`
	// The current status of the operation.
	Status *operations.OperationStatus `json:"status,omitempty"`
	// URL to get the full operation object.
	Url *string `json:"url,omitempty"`
	// Operation completed with success or failure
	Completed *bool `json:"completed,omitempty"`
	// True if all operations were successful
	HaveResultsSucceeded *bool `json:"haveResultsSucceeded,omitempty"`
	// List of results for each operation
	Results *[]OperationResult `json:"results,omitempty"`
}

type MemberEntitlementsPatchResponse struct {
	// True if all operations were successful.
	IsSuccess *bool `json:"isSuccess,omitempty"`
	// Result of the member entitlement after the operations. have been applied
	MemberEntitlement *MemberEntitlement `json:"memberEntitlement,omitempty"`
	// List of results for each operation
	OperationResults *[]OperationResult `json:"operationResults,omitempty"`
}

type MemberEntitlementsPostResponse struct {
	// True if all operations were successful.
	IsSuccess *bool `json:"isSuccess,omitempty"`
	// Result of the member entitlement after the operations. have been applied
	MemberEntitlement *MemberEntitlement `json:"memberEntitlement,omitempty"`
	// Operation result
	OperationResult *OperationResult `json:"operationResult,omitempty"`
}

type MemberEntitlementsResponseBase struct {
	// True if all operations were successful.
	IsSuccess *bool `json:"isSuccess,omitempty"`
	// Result of the member entitlement after the operations. have been applied
	MemberEntitlement *MemberEntitlement `json:"memberEntitlement,omitempty"`
}

type OperationResult struct {
	// List of error codes paired with their corresponding error messages.
	Errors *[]azuredevops.KeyValuePair `json:"errors,omitempty"`
	// Success status of the operation.
	IsSuccess *bool `json:"isSuccess,omitempty"`
	// Identifier of the Member being acted upon.
	MemberId *uuid.UUID `json:"memberId,omitempty"`
	// Result of the MemberEntitlement after the operation.
	Result *MemberEntitlement `json:"result,omitempty"`
}

// A page of users.
type PagedGraphMemberList struct {
	Members *[]UserEntitlement `json:"members,omitempty"`
	// This will be non-null if there is another page of data. There will never be more than one continuation token returned by a request.
	ContinuationToken *string `json:"continuationToken,omitempty"`
}

// Relation between a project and the user's effective permissions in that project.
type ProjectEntitlement struct {
	// Assignment Source (e.g. Group or Unknown).
	AssignmentSource *licensing.AssignmentSource `json:"assignmentSource,omitempty"`
	// Project Group (e.g. Contributor, Reader etc.)
	Group *Group `json:"group,omitempty"`
	// Deprecated: This property is deprecated. Please use ProjectPermissionInherited.
	IsProjectPermissionInherited *bool `json:"isProjectPermissionInherited,omitempty"`
	// Whether the user is inheriting permissions to a project through a Azure DevOps or AAD group membership.
	ProjectPermissionInherited *ProjectPermissionInherited `json:"projectPermissionInherited,omitempty"`
	// Project Ref
	ProjectRef *ProjectRef `json:"projectRef,omitempty"`
	// Team Ref.
	TeamRefs *[]TeamRef `json:"teamRefs,omitempty"`
}

type ProjectPermissionInherited string

type projectPermissionInheritedValuesType struct {
	NotSet       ProjectPermissionInherited
	NotInherited ProjectPermissionInherited
	Inherited    ProjectPermissionInherited
}

var ProjectPermissionInheritedValues = projectPermissionInheritedValuesType{
	NotSet:       "notSet",
	NotInherited: "notInherited",
	Inherited:    "inherited",
}

// A reference to a project.
type ProjectRef struct {
	// Project ID.
	Id *uuid.UUID `json:"id,omitempty"`
	// Project Name.
	Name *string `json:"name,omitempty"`
}

type ServicePrincipalEntitlement struct {
	// Member's access level denoted by a license.
	AccessLevel *licensing.AccessLevel `json:"accessLevel,omitempty"`
	// [Readonly] Date the member was added to the collection.
	DateCreated *azuredevops.Time `json:"dateCreated,omitempty"`
	// [Readonly] GroupEntitlements that this member belongs to.
	GroupAssignments *[]GroupEntitlement `json:"groupAssignments,omitempty"`
	// The unique identifier which matches the Id of the Identity associated with the GraphMember.
	Id *uuid.UUID `json:"id,omitempty"`
	// [Readonly] Date the member last accessed the collection.
	LastAccessedDate *azuredevops.Time `json:"lastAccessedDate,omitempty"`
	// Relation between a project and the member's effective permissions in that project.
	ProjectEntitlements *[]ProjectEntitlement `json:"projectEntitlements,omitempty"`
	// ServicePrincipal reference.
	ServicePrincipal *graph.GraphServicePrincipal `json:"servicePrincipal,omitempty"`
}

type ServicePrincipalEntitlementOperationReference struct {
	// Unique identifier for the operation.
	Id *uuid.UUID `json:"id,omitempty"`
	// Unique identifier for the plugin.
	PluginId *uuid.UUID `json:"pluginId,omitempty"`
	// The current status of the operation.
	Status *operations.OperationStatus `json:"status,omitempty"`
	// URL to get the full operation object.
	Url *string `json:"url,omitempty"`
	// Operation completed with success or failure.
	Completed *bool `json:"completed,omitempty"`
	// True if all operations were successful.
	HaveResultsSucceeded *bool `json:"haveResultsSucceeded,omitempty"`
	// List of results for each operation.
	Results *[]ServicePrincipalEntitlementOperationResult `json:"results,omitempty"`
}

type ServicePrincipalEntitlementOperationResult struct {
	// List of error codes paired with their corresponding error messages.
	Errors *[]azuredevops.KeyValuePair `json:"errors,omitempty"`
	// Success status of the operation.
	IsSuccess *bool `json:"isSuccess,omitempty"`
	// Resulting entitlement property. For specific implementations, see also:
	//   T:Microsoft.VisualStudio.Services.MemberEntitlementManagement.WebApi.ServicePrincipalEntitlementOperationResult
	//   T:Microsoft.VisualStudio.Services.MemberEntitlementManagement.WebApi.UserEntitlementOperationResult
	Result interface{} `json:"result,omitempty"`
	// Identifier of the ServicePrincipal being acted upon.
	ServicePrincipalId *uuid.UUID `json:"servicePrincipalId,omitempty"`
}

type ServicePrincipalEntitlementsPatchResponse struct {
	IsSuccess                   *bool                                         `json:"isSuccess,omitempty"`
	ServicePrincipalEntitlement *ServicePrincipalEntitlement                  `json:"servicePrincipalEntitlement,omitempty"`
	OperationResults            *[]ServicePrincipalEntitlementOperationResult `json:"operationResults,omitempty"`
}

type ServicePrincipalEntitlementsPostResponse struct {
	IsSuccess                   *bool                                       `json:"isSuccess,omitempty"`
	ServicePrincipalEntitlement *ServicePrincipalEntitlement                `json:"servicePrincipalEntitlement,omitempty"`
	OperationResult             *ServicePrincipalEntitlementOperationResult `json:"operationResult,omitempty"`
}

type ServicePrincipalEntitlementsResponseBase struct {
	IsSuccess                   *bool                        `json:"isSuccess,omitempty"`
	ServicePrincipalEntitlement *ServicePrincipalEntitlement `json:"servicePrincipalEntitlement,omitempty"`
}

type SummaryData struct {
	// Count of Licenses already assigned.
	Assigned *int `json:"assigned,omitempty"`
	// Available Count.
	Available *int `json:"available,omitempty"`
	// Quantity
	IncludedQuantity *int `json:"includedQuantity,omitempty"`
	// Total Count.
	Total *int `json:"total,omitempty"`
}

// [Flags].
type SummaryPropertyName string

type summaryPropertyNameValuesType struct {
	AccessLevels       SummaryPropertyName
	Licenses           SummaryPropertyName
	Extensions         SummaryPropertyName
	Projects           SummaryPropertyName
	Groups             SummaryPropertyName
	DefaultAccessLevel SummaryPropertyName
	All                SummaryPropertyName
}

var SummaryPropertyNameValues = summaryPropertyNameValuesType{
	AccessLevels:       "accessLevels",
	Licenses:           "licenses",
	Extensions:         "extensions",
	Projects:           "projects",
	Groups:             "groups",
	DefaultAccessLevel: "defaultAccessLevel",
	All:                "all",
}

// A reference to a team.
type TeamRef struct {
	// Team ID
	Id *uuid.UUID `json:"id,omitempty"`
	// Team Name
	Name *string `json:"name,omitempty"`
}

// A user entity with additional properties including their license, extensions, and project membership.
type UserEntitlement struct {
	// Member's access level denoted by a license.
	AccessLevel *licensing.AccessLevel `json:"accessLevel,omitempty"`
	// [Readonly] Date the member was added to the collection.
	DateCreated *azuredevops.Time `json:"dateCreated,omitempty"`
	// [Readonly] GroupEntitlements that this member belongs to.
	GroupAssignments *[]GroupEntitlement `json:"groupAssignments,omitempty"`
	// The unique identifier which matches the Id of the Identity associated with the GraphMember.
	Id *uuid.UUID `json:"id,omitempty"`
	// [Readonly] Date the member last accessed the collection.
	LastAccessedDate *azuredevops.Time `json:"lastAccessedDate,omitempty"`
	// Relation between a project and the member's effective permissions in that project.
	ProjectEntitlements *[]ProjectEntitlement `json:"projectEntitlements,omitempty"`
	// Deprecated:
	Extensions *[]Extension `json:"extensions,omitempty"`
	// User reference.
	User *graph.GraphUser `json:"user,omitempty"`
}

type UserEntitlementOperationReference struct {
	// Unique identifier for the operation.
	Id *uuid.UUID `json:"id,omitempty"`
	// Unique identifier for the plugin.
	PluginId *uuid.UUID `json:"pluginId,omitempty"`
	// The current status of the operation.
	Status *operations.OperationStatus `json:"status,omitempty"`
	// URL to get the full operation object.
	Url *string `json:"url,omitempty"`
	// Operation completed with success or failure.
	Completed *bool `json:"completed,omitempty"`
	// True if all operations were successful.
	HaveResultsSucceeded *bool `json:"haveResultsSucceeded,omitempty"`
	// List of results for each operation.
	Results *[]UserEntitlementOperationResult `json:"results,omitempty"`
}

type UserEntitlementOperationResult struct {
	// List of error codes paired with their corresponding error messages.
	Errors *[]azuredevops.KeyValuePair `json:"errors,omitempty"`
	// Success status of the operation.
	IsSuccess *bool `json:"isSuccess,omitempty"`
	// Resulting entitlement property. For specific implementations, see also:
	//   T:Microsoft.VisualStudio.Services.MemberEntitlementManagement.WebApi.ServicePrincipalEntitlementOperationResult
	//   T:Microsoft.VisualStudio.Services.MemberEntitlementManagement.WebApi.UserEntitlementOperationResult
	Result interface{} `json:"result,omitempty"`
	// Identifier of the Member being acted upon.
	UserId *uuid.UUID `json:"userId,omitempty"`
}

// [Flags].
type UserEntitlementProperty string

type userEntitlementPropertyValuesType struct {
	License    UserEntitlementProperty
	Extensions UserEntitlementProperty
	Projects   UserEntitlementProperty
	GroupRules UserEntitlementProperty
	All        UserEntitlementProperty
}

var UserEntitlementPropertyValues = userEntitlementPropertyValuesType{
	License:    "license",
	Extensions: "extensions",
	Projects:   "projects",
	GroupRules: "groupRules",
	All:        "all",
}

type UserEntitlementsPatchResponse struct {
	// True if all operations were successful.
	IsSuccess *bool `json:"isSuccess,omitempty"`
	// Result of the user entitlement after the operations have been applied.
	UserEntitlement *UserEntitlement `json:"userEntitlement,omitempty"`
	// List of results for each operation.
	OperationResults *[]UserEntitlementOperationResult `json:"operationResults,omitempty"`
}

type UserEntitlementsPostResponse struct {
	// True if all operations were successful.
	IsSuccess *bool `json:"isSuccess,omitempty"`
	// Result of the user entitlement after the operations have been applied.
	UserEntitlement *UserEntitlement `json:"userEntitlement,omitempty"`
	// Operation result.
	OperationResult *UserEntitlementOperationResult `json:"operationResult,omitempty"`
}

type UserEntitlementsResponseBase struct {
	// True if all operations were successful.
	IsSuccess *bool `json:"isSuccess,omitempty"`
	// Result of the user entitlement after the operations have been applied.
	UserEntitlement *UserEntitlement `json:"userEntitlement,omitempty"`
}

// Summary of licenses and extensions assigned to users in the organization.
type UsersSummary struct {
	// Available Access Levels
	AvailableAccessLevels *[]licensing.AccessLevel `json:"availableAccessLevels,omitempty"`
	// Default Access Level
	DefaultAccessLevel *licensing.AccessLevel `json:"defaultAccessLevel,omitempty"`
	// Deprecated:
	Extensions *[]ExtensionSummaryData `json:"extensions,omitempty"`
	// Group Options
	GroupOptions *[]GroupOption `json:"groupOptions,omitempty"`
	// Summary of Licenses in the organization
	Licenses *[]LicenseSummaryData `json:"licenses,omitempty"`
	// Summary of Projects in the organization
	ProjectRefs *[]ProjectRef `json:"projectRefs,omitempty"`
}
