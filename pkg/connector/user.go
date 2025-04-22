package connector

//same thing but for users now since we need both
import (
	"context"

	"github.com/spiros-spiros/baton-keycloak/pkg/keycloak"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
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

func (s *userSyncer) ResourceType(ctx context.Context) *v2.ResourceType {
	return &v2.ResourceType{
		Id:          "user",
		DisplayName: "User",
		Description: "A user from Keycloak",
		TraitTypes: []*v2.ResourceTypeTraitType{
			{
				Id: "user",
			},
		},
	}
}

func (s *userSyncer) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, error) {
	users, err := s.client.GetUsers(ctx)
	if err != nil {
		return nil, "", err
	}

	var resources []*v2.Resource
	for _, user := range users {
		resource := &v2.Resource{
			Id: &v2.ResourceId{
				ResourceType: "user",
				Resource:     *user.ID,
			},
			DisplayName: *user.Username,
			Description: "Keycloak User",
			TraitTypes: []*v2.ResourceTraitType{
				{
					Id: "user",
				},
			},
			UserTrait: &v2.UserTrait{
				Email:     *user.Email,
				Username:  *user.Username,
				FirstName: *user.FirstName,
				LastName:  *user.LastName,
			},
		}
		resources = append(resources, resource)
	}

	return resources, "", nil
}

func (s *userSyncer) Entitlements(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func (s *userSyncer) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}
