package connector

import (
	"context"
	"fmt"

	"github.com/Nerzal/gocloak/v13"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
	"github.com/spiros-spiros/baton-keycloak/pkg/keycloak"
)

type groupBuilder struct {
	resourceType *v2.ResourceType
	client       *keycloak.Client
}

func (o *groupBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return groupResourceType
}

func (o *groupBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	var resources []*v2.Resource
	annos := annotations.Annotations{}

	groups, err := o.client.GetGroups(ctx)
	if err != nil {
		return nil, "", nil, err
	}

	for _, group := range groups {
		groupResource, err := parseIntoGroupResource(group, nil)
		if err != nil {
			return nil, "", nil, err
		}
		resources = append(resources, groupResource)
	}

	return resources, "", annos, nil
}

func (o *groupBuilder) Entitlements(ctx context.Context, resource *v2.Resource, _ *pagination.Token) ([]*v2.Entitlement, string, annotations.Annotations, error) {
	var entitlements []*v2.Entitlement

	// Create a membership entitlement for the group
	membershipEntitlement := &v2.Entitlement{
		Id:          fmt.Sprintf("group:%s:membership", resource.Id.Resource),
		DisplayName: fmt.Sprintf("Membership in %s", resource.DisplayName),
		Description: fmt.Sprintf("Membership in the %s group", resource.DisplayName),
		GrantableTo: []*v2.ResourceType{userResourceType},
		Slug:        "membership",
	}

	entitlements = append(entitlements, membershipEntitlement)
	return entitlements, "", nil, nil
}

func (o *groupBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	var grants []*v2.Grant
	annos := annotations.Annotations{}

	// Get all users in this group
	users, err := o.client.GetUsers(ctx)
	if err != nil {
		return nil, "", nil, err
	}

	for _, user := range users {
		userGroups, err := o.client.GetUserGroups(ctx, *user.ID)
		if err != nil {
			return nil, "", nil, err
		}

		// Check if user is in this group
		for _, group := range userGroups {
			if *group.ID == resource.Id.Resource {
				userResource, err := parseIntoUserResource(user, nil)
				if err != nil {
					return nil, "", nil, err
				}

				grant := &v2.Grant{
					Id: fmt.Sprintf("grant:%s:%s", resource.Id.Resource, *user.ID),
					Entitlement: &v2.Entitlement{
						Id:          fmt.Sprintf("group:%s:membership", resource.Id.Resource),
						DisplayName: fmt.Sprintf("Membership in %s", resource.DisplayName),
						Description: fmt.Sprintf("Membership in the %s group", resource.DisplayName),
						GrantableTo: []*v2.ResourceType{userResourceType},
						Slug:        "membership",
					},
					Principal: userResource,
				}

				grants = append(grants, grant)
				break
			}
		}
	}

	return grants, "", annos, nil
}

func parseIntoGroupResource(group *gocloak.Group, parentResourceID *v2.ResourceId) (*v2.Resource, error) {
	profile := map[string]interface{}{
		"name": safeString(group.Name),
		"path": safeString(group.Path),
	}

	if group.Attributes != nil {
		if desc, ok := (*group.Attributes)["description"]; ok && len(desc) > 0 {
			profile["description"] = desc[0]
		}
	}

	groupTraits := []resource.GroupTraitOption{
		resource.WithGroupProfile(profile),
	}

	ret, err := resource.NewGroupResource(
		safeString(group.Name),
		groupResourceType,
		*group.ID,
		groupTraits,
		resource.WithParentResourceID(parentResourceID),
	)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func newGroupBuilder(client *keycloak.Client) *groupBuilder {
	return &groupBuilder{
		resourceType: groupResourceType,
		client:       client,
	}
}
