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
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
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

	for _, user := range users {
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
	l := ctxzap.Extract(ctx)
	l.Info("Starting Grant operation",
		zap.String("resource_id", resource.Id.Resource),
		zap.String("resource_display_name", resource.DisplayName),
		zap.String("entitlement_id", entitlement.Id),
	)

	if err := o.client.ensureConnected(ctx); err != nil {
		l.Error("Failed to ensure connection", zap.Error(err))
		return nil, nil, err
	}

	// The entitlement ID should be in the format: group:<groupID>:membership
	parts := strings.Split(entitlement.Id, ":")
	l.Info("Split entitlement ID parts", zap.Strings("parts", parts))
	if len(parts) != 3 || parts[0] != "group" || parts[2] != "membership" {
		l.Error("Invalid entitlement ID format")
		return nil, nil, fmt.Errorf("invalid entitlement ID format: %s", entitlement.Id)
	}

	// Get the group ID from the entitlement ID
	groupID := parts[1]
	if groupID == "" {
		l.Error("Group ID not found in entitlement ID")
		return nil, nil, fmt.Errorf("group ID not found in entitlement ID")
	}
	l.Info("Extracted group ID", zap.String("group_id", groupID))

	// Get the username from the resource
	username := resource.Id.Resource
	if username == "" {
		l.Error("Username not found in resource")
		return nil, nil, fmt.Errorf("username not found in resource")
	}
	l.Info("Extracted username", zap.String("username", username))

	// Verify the user exists
	l.Info("Searching users to verify user exists")
	users, err := o.client.client.GetUsersByUsername(ctx, username)
	if err != nil {
		l.Error("Failed to get users", zap.Error(err))
		return nil, nil, fmt.Errorf("failed to search users: %w", err)
	}
	if len(users) == 0 {
		l.Error("User not found in Keycloak", zap.String("username", username))
		return nil, nil, fmt.Errorf("user not found: %s", username)
	}
	l.Info("Found total users", zap.Int("count", len(users)))

	userID := *users[0].ID

	// Add user to group
	l.Info("Attempting to add user to group",
		zap.String("username", username),
		zap.String("user_id", userID),
		zap.String("group_id", groupID),
	)
	err = o.client.client.AddUserToGroup(ctx, userID, groupID)
	if err != nil {
		l.Error("Failed to add user to group", zap.Error(err))
		return nil, nil, fmt.Errorf("failed to add user to group: %w", err)
	}
	l.Info("Successfully added user to group")

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
	l.Info("Created grant", zap.String("grant_id", grant.Id))

	return []*v2.Grant{grant}, nil, nil
}

func (o *groupBuilder) Revoke(ctx context.Context, grant *v2.Grant) (annotations.Annotations, error) {
	l := ctxzap.Extract(ctx)
	l.Info("Starting Revoke operation",
		zap.String("grant_id", grant.Id),
		zap.String("entitlement_id", grant.Entitlement.Id),
	)

	if err := o.client.ensureConnected(ctx); err != nil {
		l.Error("Failed to ensure connection", zap.Error(err))
		return nil, err
	}

	// Extract group ID from the entitlement ID
	parts := strings.Split(grant.Entitlement.Id, ":")
	if len(parts) != 3 || parts[0] != "group" || parts[2] != "membership" {
		l.Error("Invalid entitlement ID format")
		return nil, fmt.Errorf("invalid entitlement ID format: %s", grant.Entitlement.Id)
	}

	groupID := parts[1]
	if groupID == "" {
		l.Error("Group ID not found in entitlement ID")
		return nil, fmt.Errorf("group ID not found in entitlement ID")
	}
	l.Info("Extracted group ID", zap.String("group_id", groupID))

	// Get the username from the principal
	username := grant.Principal.Id.Resource
	if username == "" {
		l.Error("Username not found in principal")
		return nil, fmt.Errorf("username not found in principal")
	}
	l.Info("Extracted username", zap.String("username", username))

	// Verify the user exists
	l.Info("Searching users to verify user exists")
	users, err := o.client.client.GetUsersByUsername(ctx, username)
	if err != nil {
		l.Error("Failed to get users", zap.Error(err))
		return nil, fmt.Errorf("failed to search users: %w", err)
	}
	if len(users) == 0 {
		l.Error("User not found in Keycloak", zap.String("username", username))
		return nil, fmt.Errorf("user not found: %s", username)
	}
	l.Info("Found total users", zap.Int("count", len(users)))

	userID := *users[0].ID

	// Remove user from group
	l.Info("Attempting to remove user from group",
		zap.String("username", username),
		zap.String("user_id", userID),
		zap.String("group_id", groupID),
	)
	err = o.client.client.RemoveUserFromGroup(ctx, userID, groupID)
	if err != nil {
		l.Error("Failed to remove user from group", zap.Error(err))
		return nil, fmt.Errorf("failed to remove user from group: %w", err)
	}
	l.Info("Successfully removed user from group")

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
