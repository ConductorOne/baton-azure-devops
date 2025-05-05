While developing the connector, please fill out this form. This information is needed to write docs and to help other users set up the connector.

## Connector capabilities

1. What resources does the connector sync?
- Users
- Teams
- Groups
- Projects
- Repositories

2. Can the connector provision any resources? If so, which ones?

This connector supports:
- Account provisioning
- Teams (grant/revoke membership to a team)
- Groups (grant/revoke membership to a group)
- Projects (grant/revoke permissions to read and write at project level) (in-progress)
- Repositories (grant/revoke permissions to read and write at repository level) (TODO)

## Connector credentials

1. What credentials or information are needed to set up the connector? (For example, API key, client ID and secret, domain, etc.)
- Personal access token
- Organization url

2. For each item in the list above:

    * How does a user create or look up that credential or info? Please include links to (non-gated) documentation, screenshots (of the UI or of gated docs), or a video of the process.
      * Documentation to create a Personal Access Token [here](https://learn.microsoft.com/en-us/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate?view=azure-devops&tabs=Windows#create-a-pat)

    * Does the credential need any specific scopes or permissions? If so, list them here.
      * For now, we are using Full Access scope. Minimal scope definitions are in progress (BB-476)

    * If applicable: Is the list of scopes or permissions different to sync (read) versus provision (read-write)? If so, list the difference here.
      * TODO BB-476: complete this section

    * What level of access or permissions does the user need in order to create the credentials? (For example, must be a super administrator, must have access to the admin console, etc.)  
      * The user needs to be `Project Collection Administrator` at organization level and `Project Administrator` at project level
      * TODO BB-476: complete this section with tested info