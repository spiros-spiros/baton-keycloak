package keycloak

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Nerzal/gocloak/v13"
)

type Client struct {
	client       *gocloak.GoCloak
	realm        string
	clientID     string
	clientSecret string
	token        *gocloak.JWT
}

func NewClient(serverURL, realm, clientID, clientSecret string) *Client {
	return &Client{
		client:       gocloak.NewClient(serverURL),
		realm:        realm,
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

func (c *Client) Connect(ctx context.Context) error {
	token, err := c.client.LoginClient(ctx, c.clientID, c.clientSecret, c.realm)
	if err != nil {
		return err
	}

	c.token = token
	return nil
}

func (c *Client) AddUserToGroup(ctx context.Context, userID, groupID string) error {
	return c.client.AddUserToGroup(ctx, c.token.AccessToken, c.realm, userID, groupID)
}

func (c *Client) RemoveUserFromGroup(ctx context.Context, userID, groupID string) error {
	return c.client.DeleteUserFromGroup(ctx, c.token.AccessToken, c.realm, userID, groupID)
}

func (c *Client) GetUsers(ctx context.Context, first int) ([]*gocloak.User, string, error) {
	max := 300

	users, err := c.client.GetUsers(ctx, c.token.AccessToken, c.realm, gocloak.GetUsersParams{
		First: pointer(first),
		Max:   pointer(max),
	})
	if err != nil {
		return nil, strconv.Itoa(first), fmt.Errorf("failed to get users: %w", err)
	}

	if len(users) == 0 {
		return nil, "", nil
	}

	return users, strconv.Itoa(first + max), nil
}

func (c *Client) GetGroupMembers(ctx context.Context, groupID string) ([]*gocloak.User, error) {
	return c.client.GetGroupMembers(ctx, c.token.AccessToken, c.realm, groupID, gocloak.GetGroupsParams{})
}

func (c *Client) GetGroups(ctx context.Context, first int) ([]*gocloak.Group, string, error) {
	max := 300

	groups, err := c.client.GetGroups(ctx, c.token.AccessToken, c.realm, gocloak.GetGroupsParams{
		First: pointer(first),
		Max:   pointer(max),
	})
	if err != nil {
		return nil, strconv.Itoa(first), fmt.Errorf("failed to get groups: %w", err)
	}

	if len(groups) == 0 {
		return nil, "", nil
	}

	return groups, strconv.Itoa(first + max), nil
}

func (c *Client) GetUserGroups(ctx context.Context, userID string) ([]*gocloak.Group, error) {
	return c.client.GetUserGroups(ctx, c.token.AccessToken, c.realm, userID, gocloak.GetGroupsParams{})
}

func (c *Client) Close() error {
	return nil
}

func (c *Client) GetUsersByUsername(ctx context.Context, username string) ([]*gocloak.User, error) {
	users, err := c.client.GetUsers(ctx, c.token.AccessToken, c.realm, gocloak.GetUsersParams{
		Username: pointer(username),
	})
	if err != nil {
		return nil, err
	}

	return users, nil
}

func pointer[T any](v T) *T {
	return &v
}
