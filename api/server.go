package main

import (
	"github.com/gorilla/mux"
	"judging-service/api/Endpoints"
	api "judging-service/api/queue"
	"judging-service/containers"
	"log"
	"net/http"
)

func main() {
	submissionQueue := api.NewSubmissionQueue()
	var manger = containers.NewContainersPoolManger(10)

	go api.ProcessQueueBackground(manger, submissionQueue)

	r := mux.NewRouter()
	r.HandleFunc("/api/submission", func(w http.ResponseWriter, r *http.Request) {
		Endpoints.ReceiveSubmissionHandler(w, r, submissionQueue)
	}).Methods("POST")
	//
	r.HandleFunc("/api/submissions", func(w http.ResponseWriter, r *http.Request) {
		Endpoints.GetAllSubmissionsHandler(w, r, submissionQueue)
	}).Methods("GET")
	//
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
