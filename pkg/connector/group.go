package connector

import (
	"context"
	"fmt"
	"strings"

	"github.com/Nerzal/gocloak/v13"
	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/pagination"
	"github.com/conductorone/baton-sdk/pkg/types/resource"
)

type groupBuilder struct {
	resourceType *v2.ResourceType
	client       *Connector
}

func (o *groupBuilder) ResourceType(ctx context.Context) *v2.ResourceType {
	return groupResourceType
}

func (o *groupBuilder) List(ctx context.Context, parentResourceID *v2.ResourceId, pToken *pagination.Token) ([]*v2.Resource, string, annotations.Annotations, error) {
	var resources []*v2.Resource
	annos := annotations.Annotations{}

	if err := o.client.ensureConnected(ctx); err != nil {
		return nil, "", nil, err
	}

	groups, err := o.client.client.GetGroups(ctx)
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

	if err := o.client.ensureConnected(ctx); err != nil {
		return nil, "", nil, err
	}

	// Create a membership entitlement for the group
	membershipEntitlement := &v2.Entitlement{
		Id:          fmt.Sprintf("group:%s:membership", resource.Id.Resource),
		DisplayName: fmt.Sprintf("Membership in %s", resource.DisplayName),
		Description: fmt.Sprintf("Membership in the %s group", resource.DisplayName),
		GrantableTo: []*v2.ResourceType{userResourceType},
		Slug:        "membership",
		Resource:    resource,
	}

	entitlements = append(entitlements, membershipEntitlement)
	return entitlements, "", nil, nil
}

func (o *groupBuilder) Grants(ctx context.Context, resource *v2.Resource, pToken *pagination.Token) ([]*v2.Grant, string, annotations.Annotations, error) {
	var grants []*v2.Grant
	annos := annotations.Annotations{}

	if err := o.client.ensureConnected(ctx); err != nil {
		return nil, "", nil, err
	}

	// Get all users in this group directly
	users, err := o.client.client.GetUsers(ctx)
	if err != nil {
		return nil, "", nil, err
	}

	// Create a map of user IDs to their resources for quick lookup
	userResources := make(map[string]*v2.Resource)
	for _, user := range users {
		userResource, err := parseIntoUserResource(user, nil)
		if err != nil {
			return nil, "", nil, err
		}
		userResources[*user.ID] = userResource
	}

	// Get users in this specific group
	groupUsers, err := o.client.client.GetUsers(ctx)
	if err != nil {
		return nil, "", nil, err
	}

	for _, user := range groupUsers {
		userGroups, err := o.client.client.GetUserGroups(ctx, *user.ID)
		if err != nil {
			return nil, "", nil, err
		}

		// Check if user is in this group
		for _, group := range userGroups {
			if *group.ID == resource.Id.Resource {
				userResource, ok := userResources[*user.ID]
				if !ok {
					continue
				}

				grant := &v2.Grant{
					Id: fmt.Sprintf("grant:%s:%s", resource.Id.Resource, *user.ID),
					Entitlement: &v2.Entitlement{
						Id:          fmt.Sprintf("group:%s:membership", resource.Id.Resource),
						DisplayName: fmt.Sprintf("Membership in %s", resource.DisplayName),
						Description: fmt.Sprintf("Membership in the %s group", resource.DisplayName),
						GrantableTo: []*v2.ResourceType{userResourceType},
						Slug:        "membership",
						Resource:    resource,
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

func (o *groupBuilder) Grant(ctx context.Context, resource *v2.Resource, entitlement *v2.Entitlement) ([]*v2.Grant, annotations.Annotations, error) {
	fmt.Printf("DEBUG: Starting Grant operation\n")
	fmt.Printf("DEBUG: Resource ID: %s, DisplayName: %s\n", resource.Id.Resource, resource.DisplayName)
	fmt.Printf("DEBUG: Entitlement ID: %s\n", entitlement.Id)

	if err := o.client.ensureConnected(ctx); err != nil {
		fmt.Printf("DEBUG: Failed to ensure connection: %v\n", err)
		return nil, nil, err
	}

	// The entitlement ID should be in the format: group:<groupID>:membership
	parts := strings.Split(entitlement.Id, ":")
	fmt.Printf("DEBUG: Split entitlement ID parts: %v\n", parts)
	if len(parts) != 3 || parts[0] != "group" || parts[2] != "membership" {
		fmt.Printf("DEBUG: Invalid entitlement ID format\n")
		return nil, nil, fmt.Errorf("invalid entitlement ID format: %s", entitlement.Id)
	}

	// Get the group ID from the entitlement ID
	groupID := parts[1]
	if groupID == "" {
		fmt.Printf("DEBUG: Group ID not found in entitlement ID\n")
		return nil, nil, fmt.Errorf("group ID not found in entitlement ID")
	}
	fmt.Printf("DEBUG: Extracted group ID: %s\n", groupID)

	// Get the username from the resource
	username := resource.Id.Resource
	if username == "" {
		fmt.Printf("DEBUG: Username not found in resource\n")
		return nil, nil, fmt.Errorf("username not found in resource")
	}
	fmt.Printf("DEBUG: Extracted username: %s\n", username)

	// Verify the user exists
	fmt.Printf("DEBUG: Fetching all users to verify user exists\n")
	users, err := o.client.client.GetUsers(ctx)
	if err != nil {
		fmt.Printf("DEBUG: Failed to get users: %v\n", err)
		return nil, nil, fmt.Errorf("failed to get users: %w", err)
	}
	fmt.Printf("DEBUG: Found %d total users\n", len(users))

	var userID string
	for _, user := range users {
		fmt.Printf("DEBUG: Checking user username: %s\n", *user.Username)
		if *user.Username == username {
			userID = *user.ID
			fmt.Printf("DEBUG: Found matching user with ID: %s\n", userID)
			break
		}
	}

	if userID == "" {
		fmt.Printf("DEBUG: User not found in Keycloak\n")
		return nil, nil, fmt.Errorf("user not found: %s", username)
	}

	// Add user to group
	fmt.Printf("DEBUG: Attempting to add user %s (ID: %s) to group %s\n", username, userID, groupID)
	err = o.client.client.AddUserToGroup(ctx, userID, groupID)
	if err != nil {
		fmt.Printf("DEBUG: Failed to add user to group: %v\n", err)
		return nil, nil, fmt.Errorf("failed to add user to group: %w", err)
	}
	fmt.Printf("DEBUG: Successfully added user to group\n")

	// Create and return the grant
	grant := &v2.Grant{
		Id: fmt.Sprintf("grant:%s:%s", groupID, userID),
		Entitlement: &v2.Entitlement{
			Id:          fmt.Sprintf("group:%s:membership", groupID),
			DisplayName: fmt.Sprintf("Membership in %s", resource.DisplayName),
			Description: fmt.Sprintf("Membership in the %s group", resource.DisplayName),
			GrantableTo: []*v2.ResourceType{userResourceType},
			Slug:        "membership",
			Resource:    resource,
		},
		Principal: &v2.Resource{
			Id: &v2.ResourceId{
				ResourceType: userResourceType.Id,
				Resource:     userID,
			},
		},
	}
	fmt.Printf("DEBUG: Created grant with ID: %s\n", grant.Id)

	return []*v2.Grant{grant}, nil, nil
}

func (o *groupBuilder) Revoke(ctx context.Context, grant *v2.Grant) (annotations.Annotations, error) {
	if err := o.client.ensureConnected(ctx); err != nil {
		return nil, err
	}

	// Extract user ID and group ID from the grant
	userID := grant.Principal.Id.Resource
	groupID := grant.Entitlement.Resource.Id.Resource

	// Remove user from group
	err := o.client.client.RemoveUserFromGroup(ctx, userID, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to remove user from group: %w", err)
	}

	return nil, nil
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

func newGroupBuilder(client *Connector) *groupBuilder {
	return &groupBuilder{
		resourceType: groupResourceType,
		client:       client,
	}
}
