package connector

//same thing but for users now since we need both
import (
	"context"

	"github.com/spiros-spiros/baton-keycloak/pkg/keycloak"
	"github.com/conductorone/baton-sdk/pkg/connectorbuilder"
	"github.com/conductorone/baton-sdk/pkg/pagination"
)

type userSyncer struct {
	client *keycloak.Client
}

func newUserSyncer(client *keycloak.Client) *userSyncer {
	return &userSyncer{
		client: client,
	}
}

func (s *userSyncer) ResourceType(ctx context.Context) *connectorbuilder.ResourceType {
	return &connectorbuilder.ResourceType{
		ID:          "user",
		DisplayName: "User",
		TraitOptions: []connectorbuilder.TraitOption{
			connectorbuilder.WithUserTrait(),
		},
	}
}

func (s *userSyncer) List(ctx context.Context, parentResourceID *connectorbuilder.ResourceID, pToken *pagination.Token) ([]*connectorbuilder.Resource, string, error) {
	users, err := s.client.GetUsers(ctx)
	if err != nil {
		return nil, "", err
	}

	var resources []*connectorbuilder.Resource
	for _, user := range users {
		resource := &connectorbuilder.Resource{
			ID: &connectorbuilder.ResourceID{
				ResourceType: "user",
				Resource:     *user.ID,
			},
			DisplayName: *user.Username,
			Traits: []connectorbuilder.Trait{
				&connectorbuilder.UserTrait{
					// we need all of this to look up in C1, otherwise we might miss users - since they're not federated with our IdP!!!!!
					Email:     *user.Email,
					Username:  *user.Username,
					FirstName: *user.FirstName,
					LastName:  *user.LastName,
				},
			},
		}
		resources = append(resources, resource)
	}

	return resources, "", nil
}

func (s *userSyncer) Entitlements(ctx context.Context, resource *connectorbuilder.Resource, pToken *pagination.Token) ([]*connectorbuilder.Entitlement, string, error) {
	return nil, "", nil
}

func (s *userSyncer) Grants(ctx context.Context, resource *connectorbuilder.Resource, pToken *pagination.Token) ([]*connectorbuilder.Grant, string, error) {
	return nil, "", nil
}

