![Baton Logo](./baton-logo.png)

# `baton-azure-devops` [![Go Reference](https://pkg.go.dev/badge/github.com/conductorone/baton-azure-devops.svg)](https://pkg.go.dev/github.com/conductorone/baton-azure-devops) ![main ci](https://github.com/conductorone/baton-azure-devops/actions/workflows/main.yaml/badge.svg)

`baton-azure-devops` is a connector for built using the [Baton SDK](https://github.com/conductorone/baton-sdk).

Check out [Baton](https://github.com/conductorone/baton) to learn more the project in general.

# Prerequisites
Follow [Microsoft Learn Guide](https://learn.microsoft.com/en-us/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate?view=azure-devops&tabs=Windows#create-a-pat) to create a Personal Access Token.
Minimal permission scope needed includes:
- Full access for now. Scope definition is still in progress.

# Getting Started

## brew

```
brew install conductorone/baton/baton conductorone/baton/baton-azure-devops
baton-azure-devops
baton resources
```

## docker

```
docker run --rm -v $(pwd):/out -e BATON_DOMAIN_URL=domain_url -e BATON_API_KEY=apiKey -e BATON_USERNAME=username ghcr.io/conductorone/baton-azure-devops:latest -f "/out/sync.c1z"
docker run --rm -v $(pwd):/out ghcr.io/conductorone/baton:latest -f "/out/sync.c1z" resources
```

## source

```
go install github.com/conductorone/baton/cmd/baton@main
go install github.com/conductorone/baton-azure-devops/cmd/baton-azure-devops@main

baton-azure-devops

baton resources
```

# Data Model

`baton-azure-devops` will pull down information about the following resources:
- Users
- Teams
- Groups
- Projects
- Repositories

# Contributing, Support and Issues

We started Baton because we were tired of taking screenshots and manually
building spreadsheets. We welcome contributions, and ideas, no matter how
small&mdash;our goal is to make identity and permissions sprawl less painful for
everyone. If you have questions, problems, or ideas: Please open a GitHub Issue!

See [CONTRIBUTING.md](https://github.com/ConductorOne/baton/blob/main/CONTRIBUTING.md) for more details.

# `baton-azure-devops` Command Line Usage

```
baton-azure-devops

Usage:
  baton-azure-devops [flags]
  baton-azure-devops [command]

Available Commands:
  capabilities       Get connector capabilities
  completion         Generate the autocompletion script for the specified shell
  help               Help about any command

Flags:
      --client-id string             The client ID used to authenticate with ConductorOne ($BATON_CLIENT_ID)
      --client-secret string         The client secret used to authenticate with ConductorOne ($BATON_CLIENT_SECRET)
  -f, --file string                  The path to the c1z file to sync with ($BATON_FILE) (default "sync.c1z")
  -h, --help                         help for baton-azure-devops
      --log-format string            The output format for logs: json, console ($BATON_LOG_FORMAT) (default "json")
      --log-level string             The log level: debug, info, warn, error ($BATON_LOG_LEVEL) (default "info")
      --organization-url string      required: The organization url to sync data `https://dev.azure.com/{Your_Organization}` ($BATON_ORGANIZATION_URL)
      --personal-access-token string required: The Personal Access Token (PAT) that serves as an alternative password for authenticating into Azure DevOps ($BATON_PAT)
  -p, --provisioning                 If this connector supports provisioning, this must be set in order for provisioning actions to be enabled ($BATON_PROVISIONING)
      --sync-grant-sources boolean   Sync grant sources. If this is not set, grant sources will not be included ($BATON_SYNC_GRANT_SOURCES)
      --ticketing                    This must be set to enable ticketing support ($BATON_TICKETING)
  -v, --version                      version for baton-azure-devops

Use "baton-azure-devops [command] --help" for more information about a command.
```
