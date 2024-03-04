package sales

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/plantyplantman/bcapi/pkg/entities"
)

type Service interface {
	GetSales(time.Time, time.Time) ([]*entities.Sale, error)
	GetSaleIdsByDate(time.Time, time.Time) ([]int, error)
	GetSalesByID(int, int) ([]*entities.Sale, error)
}

type service struct {
	URL    *url.URL
	SiteID int
}

func NewService(siteID int, serviceURL string) (Service, error) {
	URL, err := url.Parse(serviceURL)
	if err != nil {
		return nil, err
	}

	URL = URL.JoinPath("api").JoinPath(strconv.Itoa(siteID)).JoinPath("sales")

	return &service{
		URL:    URL,
		SiteID: siteID,
	}, nil
}

func (s *service) GetSales(from, to time.Time) ([]*entities.Sale, error) {
	_url := s.URL
	params := _url.Query()
	params.Add("from", from.Format("2006-01-02T15:04:05"))
	params.Add("to", to.Format("2006-01-02T15:04:05"))
	_url.RawQuery = params.Encode()

	log.Println(_url.String())
	resp, err := http.Get(_url.String())
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data []*entities.Sale
	err = json.Unmarshal(b, &data)
	if err != nil {
		log.Println(string(b))
		return nil, err
	}

	return data, nil
}

func (s *service) GetSaleIdsByDate(from, to time.Time) ([]int, error) {
	_url := s.URL.JoinPath("ids")
	params := _url.Query()
	params.Add("from", from.Format("2006-01-02T15:04:05"))
	params.Add("to", to.Format("2006-01-02T15:04:05"))
	_url.RawQuery = params.Encode()

	resp, err := http.Get(_url.String())
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data []int
	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *service) GetSalesByID(first, last int) ([]*entities.Sale, error) {
	_url := s.URL
	params := _url.Query()
	params.Add("first", strconv.Itoa(first))
	params.Add("last", strconv.Itoa(last))
	_url.RawQuery = params.Encode()

	resp, err := http.Get(_url.String())
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data []*entities.Sale
	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
