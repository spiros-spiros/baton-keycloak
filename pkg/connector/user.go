package connector

//same thing but for users now since we need both
import (
	"context"

	"github.com/spiros-spiros/baton-keycloak/pkg/keycloak"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
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
		Traits:      []v2.ResourceType_Trait{v2.ResourceType_TRAIT_USER},
	}
}

func (s *userSyncer) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	users, err := s.client.GetUsers(ctx)
	if err != nil {
		return nil, "", nil, err
	}

	var resources []*v2.Resource
	for _, user := range users {
		resource, err := resource.NewUserResource(
			*user.Username,
			s.ResourceType(ctx),
			*user.ID,
			[]resource.UserTraitOption{
				resource.WithUserEmail(*user.Email),
				resource.WithUserProfile(*user.FirstName, *user.LastName),
				resource.WithUserStatus(v2.UserTrait_Status_STATUS_ENABLED),
			},
		)
		if err != nil {
			return nil, "", nil, err
		}
		resources = append(resources, resource)
	}

	return resources, "", nil, nil
}

func (s *userSyncer) Entitlements(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

func (s *userSyncer) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}
