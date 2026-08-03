// Package teamservice teamhandler.go implements
// endpoints of team-service
package teamservice

import "net/http"

func (ts *teamservice) newTeam(w http.ResponseWriter, r *http.Request) {
	const op = "taskservice.newTeam"
}

func (ts *teamservice) getTeams(w http.ResponseWriter, r *http.Request) {
	const op = "taskservice.getTeams"
}

func (ts *teamservice) inviteByID(w http.ResponseWriter, r *http.Request) {
	const op = "taskservice.inviteByID"
}
