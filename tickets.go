package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type ApiDataSource struct {
}

type TicketList struct {
	Tickets []Ticket
	Meta    struct {
		HasMore   bool `json:"has_more"`
		HasBefore bool
	}
	Links struct {
		Prev string `json:"prev"`
		Next string `json:"next"`
	}
}

// pretty useless to store all the information when in reality you only need id, subject and url for list view
// but it makes it easier to use json methods, and there will only be 25 tickets at most in memory, so it shouldn't be
// to much of a problem
type Ticket struct {
	Id            int64  `json:"id"`
	Subject       string `json:"subject"`
	Description   string `json:"description"`
	RequesterId   int64  `json:"requester_id"`
	RequesterName string
	Status        string `json:"status"`
	Priority      string `json:"priority"`
}

// retrives tickets with given id from api
func (ads *ApiDataSource) GetTickets(path string) (TicketList, error) {
	var ticketList TicketList
	// fetch api
	req, err := getNewRequest(path)
	if err != nil {
		logger.Println(err)
		return ticketList, err
	}
	res, err := fetchApi(req)
	if err != nil {
		logger.Println(err)
		return ticketList, err
	}

	// parse response body (json)
	decoder := json.NewDecoder(res.Body)

	err = decoder.Decode(&ticketList)
	if err != nil {
		logger.Println(err)
	}

	return ticketList, err
}

// retrives ticket with given id from api
func (ads *ApiDataSource) GetTicket(path string) (Ticket, error) {
	return Ticket{}, nil
}

// retrieves user with given id from api.
func getUser(id int64) {

}

func getNewRequest(path string) (*http.Request, error) {
	godotenv.Load()
	url := parseUrl(path)
	logger.Println(url)
	email := os.Getenv("API_EMAIL")
	token := os.Getenv("API_TOKEN")

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(email+"/token", token)
	req.Header.Add("Accept", "application/json")

	return req, nil
}

func fetchApi(req *http.Request) (*http.Response, error) {
	client := http.Client{
		Timeout: time.Second * 10,
	}
	return client.Do(req)
}

const baseDomain = "https://%s.zendesk.com/api/v2"

func parseUrl(path string) string {
	godotenv.Load()
	subDomain := os.Getenv("API_SUBDOMAIN")
	domain := fmt.Sprintf(baseDomain, subDomain)
	logger.Println(domain + path)
	return domain + path
}
