package test

import (
	"fmt"
	"github.com/go-yaaf/yaaf-common-net/model"
	"github.com/go-yaaf/yaaf-common-net/utils"
	"testing"

	"github.com/stretchr/testify/require"
)

var secret = "put your secret string. It must be at least 32 characters long"
var signing = "put your signing key string. It must be at least 32 characters long"

func TestToken(t *testing.T) {

	tu := utils.TokenUtils().WithSecrets(secret, signing)

	td := &model.TokenData{
		AccountId:   "accountId",
		SubjectId:   "subject@email.com",
		SubjectType: 2,
		SubjectRole: 12,
		Status:      1,
		ExpiresIn:   0,
	}

	token, err := tu.CreateToken(td)
	if err != nil {
		t.Fail()
	}
	fmt.Println(token)

	// Parse token
	actual, er := tu.ParseToken(token)

	require.Nil(t, er, "actual is nil")
	require.Equal(t, td.AccountId, actual.AccountId, "incompatible Account ID")
	require.Equal(t, td.SubjectId, actual.SubjectId, "incompatible Subject ID")
	require.Equal(t, td.SubjectType, actual.SubjectType, "incompatible Subject Type")
	require.Equal(t, td.SubjectRole, actual.SubjectRole, "incompatible Subject Role")
	require.Equal(t, td.Status, actual.Status, "incompatible Status")
}

func TestApiKey(t *testing.T) {

	tu := utils.TokenUtils().WithSecrets(secret, signing)
	appName := "rest-server-example"
	apiKey, err := tu.CreateApiKey(appName)
	if err != nil {
		t.Fail()
	}
	fmt.Println(apiKey)

	// Parse token
	actual, er := tu.ParseApiKey(apiKey)

	require.Nil(t, er, "actual is nil")
	require.Equal(t, actual, appName, "incompatible App Name")
}
