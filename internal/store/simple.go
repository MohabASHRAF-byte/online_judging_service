package store

import (
	"judging-service/internal/models"
)

type SimpleStore struct {
	submissions []models.Submission
	index       int
}

func NewSimpleStore() *SimpleStore {
	submissions := []models.Submission{
		{
			ID:         "2",
			Language:   "c++",
			Code:       "#include <bits/stdc++.h>\nusing namespace std;\nsigned main() {\n    int x,y;cin>>x>>y;\n    cout<<x+y;\n}\n",
			TestInputs: []string{"1 3", "-9 9", "90 7"},
		},
		{
			ID:         "1",
			Language:   "python",
			Code:       "x, y = map(int, input().split())\nprint(x + y)\n",
			TestInputs: []string{"1 2", "-9 9"},
		},
	}

	return &SimpleStore{
		submissions: submissions,
		index:       0,
	}
}

func (s *SimpleStore) GetNext() *models.Submission {
	if len(s.submissions) == 0 {
		return nil
	}
	submission := &s.submissions[s.index]
	s.index = (s.index + 1) % len(s.submissions)
	return submission
}
