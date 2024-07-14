package user

import (
	"context"
	"errors"
	"net/http"

	"github.com/kirillgashkov/timetrack/api/peopleinfoapi/v1"
)

var (
	ErrPeopleInfoNotFound    = errors.New("people info not found")
	ErrPeopleInfoUnavailable = errors.New("people info service is unavailable")
)

type PeopleInfo struct {
	Name       string
	Patronymic *string
	Surname    string
	Address    string
}

type PeopleInfoService interface {
	Get(ctx context.Context, passportSeries int, passportNumber int) (*PeopleInfo, error)
}

type PeopleInfoServiceImpl struct {
	client *peopleinfoapi.Client
}

func NewPeopleInfoServiceImpl(serverURL string, httpClient *http.Client) (*PeopleInfoServiceImpl, error) {
	peopleInfoClient, err := peopleinfoapi.NewClient(
		serverURL, peopleinfoapi.WithHTTPClient(httpClient),
	)
	if err != nil {
		return nil, errors.Join(errors.New("failed to create people info client"), err)
	}
	return &PeopleInfoServiceImpl{client: peopleInfoClient}, nil
}

func (s *PeopleInfoServiceImpl) Get(ctx context.Context, passportSeries int, passportNumber int) (*PeopleInfo, error) {
	peopleInfoParams := &peopleinfoapi.GetInfoParams{
		PassportSerie: passportSeries, PassportNumber: passportNumber,
	}

	resp, err := s.client.GetInfo(ctx, peopleInfoParams)
	if err != nil {
		return nil, errors.Join(errors.New("failed to send request for people info"), ErrPeopleInfoUnavailable, err)
	}
	if resp.StatusCode != http.StatusOK {
		// Per assignment, the PeopleInfo service only responds with 400 Bad
		// Request when something is wrong.
		if resp.StatusCode == http.StatusBadRequest {
			return nil, ErrPeopleInfoNotFound
		}
		return nil, errors.Join(errors.New("people info request failed with status "+resp.Status), ErrPeopleInfoUnavailable)
	}

	getInfoResponse, err := peopleinfoapi.ParseGetInfoResponse(resp)
	if err != nil {
		return nil, errors.Join(errors.New("failed to parse people info response"), ErrPeopleInfoUnavailable, err)
	}
	if getInfoResponse.JSON200 == nil {
		return nil, errors.Join(errors.New("people info response JSON is empty"), ErrPeopleInfoUnavailable)
	}

	peopleInfo := getInfoResponse.JSON200
	return &PeopleInfo{
		Name:       peopleInfo.Name,
		Patronymic: peopleInfo.Patronymic,
		Surname:    peopleInfo.Surname,
		Address:    peopleInfo.Address,
	}, nil
}
