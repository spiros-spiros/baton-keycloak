package connector

import (
	"context"

	"github.com/Nerzal/gocloak/v13"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/spiros-spiros/baton-keycloak/pkg/keycloak"
)

// userBuilder implements the resource builder interface for Keycloak user resources.
// It handles the creation and synchronization of user resources between Keycloak and Baton.
type userBuilder struct {
	resourceType *v2.ResourceType
	client       *keycloak.Client
}

// ResourceType returns the v2.ResourceType for users.
// This identifies the type of resources this builder manages.
func (o *userBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return userResourceType
}

// List retrieves all user resources from Keycloak and converts them to the Baton format.
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - parentResourceID: The parent resource ID (unused in this implementation)
//   - pToken: Pagination token for handling large result sets
//
// Returns:
//   - []*v2.Resource: List of user resources
//   - string: Next page token for pagination
//   - annotations.Annotations: Additional metadata
//   - error: Any error that occurred during the operation
func (o *userBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	var resource []*v2.Resource
	annos := annotations.Annotations{}

	users, err := o.client.GetUsers(ctx)
	if err != nil {
		return nil, "", nil, err
	}

	for _, user := range users {
		userResource, err := parseIntoUserResource(user, nil)
		if err != nil {
			return nil, "", nil, err
		}
		resource = append(resource, userResource)
	}

	return resource, "", annos, nil
}

// Entitlements returns empty as users don't have direct entitlements in this implementation.
// Entitlements are managed separately by the permissionBuilder.
func (o *userBuilder) Entitlements(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// Grants returns empty as this builder doesn't handle grants directly.
// Grants are managed separately by the permissionBuilder.
func (o *userBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	return nil, "", nil, nil
}

// newUserBuilder creates a new instance of userBuilder.
// This is the constructor function for the userBuilder struct.
func newUserBuilder(client *keycloak.Client) *userBuilder {
	return &userBuilder{
		resourceType: userResourceType,
		client:       client,
	}
}

// parseIntoUserResource converts a Linode user object into a Baton SDK user resource.
// Parameters:
//   - ctx: Context (currently unused)
//   - user: Pointer to the Linode user object to convert
//   - parentResourceID: Optional parent resource ID for hierarchy
//
// Returns:
//   - *v2.Resource: The converted Baton resource
//   - error: Any conversion error that occurred
func parseIntoUserResource(user *gocloak.User, parentResourceID *v2.ResourceId) (*v2.Resource, error) {
	var userStatus = v2.UserTrait_Status_STATUS_ENABLED

	username := ""
	if user.Username != nil {
		username = *user.Username
	}

	profile := map[string]interface{}{
		"username":  username,
		"email":     safeString(user.Email),
		"firstName": safeString(user.FirstName),
		"lastName":  safeString(user.LastName),
	}

	userTraits := []resource.UserTraitOption{
		resource.WithUserProfile(profile),
		resource.WithUserLogin(username),
		resource.WithStatus(userStatus),
	}

	ret, err := resource.NewUserResource(
		username,
		userResourceType,
		username,
		userTraits,
		resource.WithParentResourceID(parentResourceID),
	)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
