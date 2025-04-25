package connector

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/conductorone/baton-azure-devops/pkg/client"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

func unmarshalProperties(properties interface{}) (map[string]interface{}, error) {
	rawBytes, err := json.Marshal(properties)
	if err != nil {
		return nil, err
	}

	var propsMap map[string]interface{}
	err = json.Unmarshal(rawBytes, &propsMap)
	if err != nil {
		return nil, err
	}

	return propsMap, nil
}

func getIdentityResourceByDescriptor(client *client.AzureDevOpsClient, ctx context.Context, descriptor string, userMap map[string]string) (*v2.Resource, error) {
	// get user email to map with user
	parts := strings.Split(descriptor, `\`)

	if len(parts) == 2 {
		userPrincipalName := parts[1]
		if userMap[userPrincipalName] != "" {
			userResource := &v2.Resource{
				Id: &v2.ResourceId{
					ResourceType: userResourceType.Id,
					Resource:     userMap[userPrincipalName],
				},
			}
			return userResource, nil
		}
	} else {
		teamsMap, err := client.ListTeamIDs(ctx)
		if err != nil {
			return nil, err
		}
		groupIdentities, err := client.ListIdentities(ctx, "", descriptor)
		if err != nil {
			return nil, err
		}
		if len(groupIdentities) > 0 {
			groupIdentity := groupIdentities[0]
			if groupIdentity.Id != nil {
				if _, isTeam := teamsMap[groupIdentity.Id.String()]; isTeam {
					teamResource := &v2.Resource{
						Id: &v2.ResourceId{
							ResourceType: teamResourceType.Id,
							Resource:     groupIdentity.Id.String(),
						},
					}
					return teamResource, nil
				} else {
					groupResource := &v2.Resource{
						Id: &v2.ResourceId{
							ResourceType: groupResourceType.Id,
							Resource:     groupIdentity.Id.String(),
						},
					}
					return groupResource, nil
				}
			}
		}
	}
	return nil, errors.New("no identity resource found")
}
