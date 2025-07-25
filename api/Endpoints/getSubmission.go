package Endpoints

import (
	"encoding/json"
	api "judging-service/api/queue"
	"log"
	"net/http"
)

func GetAllSubmissionsHandler(w http.ResponseWriter, r *http.Request, queue *api.SubmissionQueue) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("Retrieving all submissions from queue")

	submissions := queue.GetAll()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"submissions": submissions,
		"totalCount":  len(submissions),
	})
}
