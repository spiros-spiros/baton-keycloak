package connector

import (
	"context"
	// the conductor one SDKs are already built, so this bit should be easy
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/spiros-spiros/baton-keycloak/pkg/keycloak"
)

type groupSyncer struct {
	client *keycloak.Client
}

func newGroupSyncer(client *keycloak.Client) *groupSyncer {
	return &groupSyncer{
		client: client,
	}
}

// might need to add roles as well as groups, but let's see if we can get this working first
func (s *groupSyncer) ResourceType(ctx context.Context) *connectorbuilder.ResourceType {
	return &connectorbuilder.ResourceType{
		ID:          "group",
		DisplayName: "Group",
		TraitOptions: []connectorbuilder.TraitOption{
			connectorbuilder.WithGroupTrait(),
		},
	}
}

func (s *groupSyncer) List(ctx context.Context, parentResourceID *connectorbuilder.ResourceID, pToken *pagination.Token) ([]*connectorbuilder.Resource, string, error) {
	groups, err := s.client.GetGroups(ctx)
	if err != nil {
		return nil, "", err
	}

	var resources []*connectorbuilder.Resource
	for _, group := range groups {
		resource := &connectorbuilder.Resource{
			ID: &connectorbuilder.ResourceID{
				ResourceType: "group",
				Resource:     *group.ID,
			},
			DisplayName: *group.Name,
			Traits: []connectorbuilder.Trait{
				&connectorbuilder.GroupTrait{
					Name: *group.Name,
				},
			},
		}
		resources = append(resources, resource)
	}

	return resources, "", nil
}

// get groups
func (s *groupSyncer) Entitlements(ctx context.Context, resource *connectorbuilder.Resource, pToken *pagination.Token) ([]*connectorbuilder.Entitlement, string, error) {
	return nil, "", nil
}

// now get entitlements
func (s *groupSyncer) Grants(ctx context.Context, resource *connectorbuilder.Resource, pToken *pagination.Token) ([]*connectorbuilder.Grant, string, error) {
	return nil, "", nil
}

