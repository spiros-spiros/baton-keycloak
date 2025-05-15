package utils

import (
	"strconv"

	"github.com/conductorone/baton-sdk/pkg/pagination"
)

func ParseToken(pToken *pagination.Token) int {
	if pToken == nil {
		return 0
	}

	token, err := strconv.Atoi(pToken.Token)
	if err != nil {
		return 0
	}

	return token
}
