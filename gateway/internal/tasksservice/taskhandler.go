// Package taskservice taskhandler.go implements
// endpoints of tasks-service
package taskservice

import "net/http"

func (ts *taskservice) newTask(w http.ResponseWriter, r *http.Request) {
	const op = "taskservice.newTask"
}

func (ts *taskservice) getTasks(w http.ResponseWriter, r *http.Request) {
	const op = "taskservice.getTasks"
}

func (ts *taskservice) updTask(w http.ResponseWriter, r *http.Request) {
	const op = "taskservice.updTask"
}

func (ts *taskservice) getTaskHistory(w http.ResponseWriter, r *http.Request) {
	const op = "taskservice.getTaskHistory"
}
