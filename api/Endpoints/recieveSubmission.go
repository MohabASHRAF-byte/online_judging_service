package Endpoints

import (
	"encoding/json"
	"judging-service/api/Dtos"
	api "judging-service/api/queue"
	"log"
	"net/http"
)

func ReceiveSubmissionHandler(w http.ResponseWriter, r *http.Request, queue *api.SubmissionQueue) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var submission Dtos.SubmissionQueueDto
	if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	queue.Insert(submission)
}
