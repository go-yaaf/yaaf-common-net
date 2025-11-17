package test

import (
	"fmt"
	"github.com/go-yaaf/yaaf-common-net/model"
	"github.com/go-yaaf/yaaf-common-net/utils"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToken(t *testing.T) {

	secret := "thisIsmyS3cr3t4@nyPurp0seButiTmuSTbelongetThan32bytes"
	signing := "thisIsmyS3cr3tSigningKeyAndItIsAls0MUSTbeLongerTh@n32BytES"
	tu := utils.TokenUtils().WithSecrets(secret, signing)

	td := &model.TokenData{
		SubjectId:   "subject@email.com",
		SubjectType: 2,
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
	require.Equal(t, td.SubjectId, actual.SubjectId, "incompatible SubjectId")
	require.Equal(t, td.SubjectType, actual.SubjectType, "incompatible SubjectType")
	require.Equal(t, td.Status, actual.Status, "incompatible Status")
}
