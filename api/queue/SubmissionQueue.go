package api

import (
	"judging-service/api/Dtos"
	"sync"
)

type SubmissionQueue struct {
	queue []Dtos.SubmissionQueueDto
	mutex sync.Mutex
}

func NewSubmissionQueue() *SubmissionQueue {
	return &SubmissionQueue{
		queue: make([]Dtos.SubmissionQueueDto, 0),
	}
}

func (sq *SubmissionQueue) Insert(submission Dtos.SubmissionQueueDto) {
	sq.mutex.Lock()
	defer sq.mutex.Unlock()
	sq.queue = append(sq.queue, submission)
}

func (sq *SubmissionQueue) Pull() (*Dtos.SubmissionQueueDto, bool) {
	sq.mutex.Lock()
	defer sq.mutex.Unlock()

	if len(sq.queue) == 0 {
		return nil, false
	}

	submission := sq.queue[0]
	sq.queue = sq.queue[1:]
	return &submission, true
}

func (sq *SubmissionQueue) Size() int {
	sq.mutex.Lock()
	defer sq.mutex.Unlock()
	return len(sq.queue)
}

func (sq *SubmissionQueue) IsEmpty() bool {
	sq.mutex.Lock()
	defer sq.mutex.Unlock()
	return len(sq.queue) == 0
}
func (sq *SubmissionQueue) GetAll() []Dtos.SubmissionQueueDto {
	sq.mutex.Lock()
	defer sq.mutex.Unlock()

	result := make([]Dtos.SubmissionQueueDto, len(sq.queue))
	copy(result, sq.queue)
	return result
}
