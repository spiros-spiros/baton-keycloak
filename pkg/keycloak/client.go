package keycloak

import (
	"context"

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

func (c *Client) GetUsers(ctx context.Context) ([]*gocloak.User, error) {
	return c.client.GetUsers(ctx, c.token.AccessToken, c.realm, gocloak.GetUsersParams{})
}

func (c *Client) GetGroups(ctx context.Context) ([]*gocloak.Group, error) {
	return c.client.GetGroups(ctx, c.token.AccessToken, c.realm, gocloak.GetGroupsParams{})
}

func (c *Client) GetUserGroups(ctx context.Context, userID string) ([]*gocloak.Group, error) {
	return c.client.GetUserGroups(ctx, c.token.AccessToken, c.realm, userID, gocloak.GetGroupsParams{})
}

func (c *Client) Close() error {
	return nil
}

