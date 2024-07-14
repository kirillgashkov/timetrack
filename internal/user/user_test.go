package user

import "context"

type PeopleInfoServiceMock struct{}

func (s *PeopleInfoServiceMock) Get(_ context.Context, _ int, passportNumber int) (*PeopleInfo, error) {
	var patronymic *string
	if passportNumber%7 == 0 {
		patronymic = new(string)
		*patronymic = "Ivanovich"
	}

	return &PeopleInfo{
		Name:       "Ivan",
		Patronymic: nil,
		Surname:    "Ivanov",
		Address:    "Ivanovskaya st., 1",
	}, nil
}
