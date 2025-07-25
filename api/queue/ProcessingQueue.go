package api

import (
	"fmt"
	"judging-service/containers"
	"judging-service/internal/processor"
	"time"
)

func ProcessQueueBackground(poolManger *containers.ContainersPoolManger, queue *SubmissionQueue) {
	for {
		if submission, ok := queue.Pull(); ok {
			var output, err = processor.RunCodeWithTestcases(poolManger, *submission)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(output.Verdict)
			for _, testCase := range output.Outputs {
				fmt.Println(testCase.TestCaseId, testCase.Output)
			}
			err = SubmitJudgingResult(output, "http://localhost:5129/")
			if err != nil {
				fmt.Println(err)
			}
		} else {
			time.Sleep(1 * time.Second)
		}
	}
}
