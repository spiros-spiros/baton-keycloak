package connector

import (
	"context"
	// the conductor one SDKs are already built, so this bit should be easy
	"github.com/conductorone/baton-keycloak/pkg/keycloak"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
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
func (s *groupSyncer) ResourceType(ctx context.Context) *v2.ResourceType {
	return &v2.ResourceType{
		Id:          "group",
		DisplayName: "Group",
		Description: "A group from Keycloak",
		TraitOptions: []*v2.ResourceTypeTraitOption{
			{
				Trait: &v2.ResourceTypeTrait{
					Id: "group",
				},
			},
		},
	}
}

func (s *groupSyncer) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	groups, err := s.client.GetGroups(ctx)
	if err != nil {
		return nil, "", nil, err
	}

	var resources []*v2.Resource
	for _, group := range groups {
		resource := &v2.Resource{
			Id: &v2.ResourceId{
				ResourceType: "group",
				Resource:     *group.ID,
			},
			DisplayName: *group.Name,
			Description: "Keycloak Group",
			Traits: []*v2.ResourceTrait{
				{
					Id: "group",
					Trait: &v2.ResourceTrait_GroupTrait{
						GroupTrait: &v2.GroupTrait{
							Name: *group.Name,
						},
					},
				},
			},
		}
		resources = append(resources, resource)
	}

	return resources, "", nil, nil
}

// get groups
func (s *groupSyncer) Entitlements(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// now get entitlements
func (s *groupSyncer) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}
