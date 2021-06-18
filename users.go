package main

import (
	"encoding/base64"
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	uuid "github.com/nu7hatch/gouuid"
)

type AccountType int32

const (
	GUEST AccountType = 0
	USER  AccountType = 1
	ADMIN AccountType = 2
)

type User struct {
	Username string
	token    string
	Role     AccountType
	pubKey   string
}

func genToken(user *User) string {
	u, err := uuid.NewV4()
	if err != nil {
		log.Print(err.Error())
	}
	user.token = u.String()
	return encrypt(user.token, user.pubKey)
}

func encrypt(token string, pubKey string) string {
	return base64.StdEncoding.EncodeToString(([]byte)(token + pubKey))
}

func checkToken(user *User, token string) bool {
	return user.token == token
}

func CheckPermission(r *http.Request, neededPermissions AccountType) error {
	vars := mux.Vars(r)

	username, _ := vars["username"]
	token, _ := vars["token"]

	user, ok := GetUser(username)
	if !(ok && checkToken(user, token)) {
		return errors.New("token is not valid")
	}

	if user.Role < neededPermissions {
		return errors.New("not enough permissions")
	}

	return nil

}
