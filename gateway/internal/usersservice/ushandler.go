// Package usersservice ushandler.go implements
// endpoints of users-service
package usersservice

import "net/http"

func (us *UsersService) register(w http.ResponseWriter, r *http.Request) {
	const op = "usersservice.register"
}

func (us *UsersService) login(w http.ResponseWriter, r *http.Request) {
	const op = "usersservice.login"
}
