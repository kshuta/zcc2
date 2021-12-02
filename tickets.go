package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const itemLimit = 25 // test will not pass if itemLimit less than 2

type ApiDataSource struct {
}

type TicketList struct {
	Tickets          []Ticket `json:"tickets"`
	Count            int      `json:"count"`
	NextPage         string   `json:"next_page"`
	PreviousPage     string   `json:"previous_page"`
	PageNum          string
	LastPageNum      int
	TicketCountLimit int
}

// pretty useless to store all the information when in reality you only need id, subject and url for list view
// but it makes it easier to use json methods, and there will only be 25 tickets at most in memory, so it shouldn't be
// to much of a problem
type Ticket struct {
	Id            int64  `json:"id"`
	Subject       string `json:"subject"`
	Description   string `json:"description"`
	Status        string `json:"status"`
	Priority      string `json:"priority"`
	RequesterName string
	BackPage      string
	Tags          []string `json:"tags"`
}

// retrives tickets with given id from api
func (ads *ApiDataSource) GetTickets(path string, query url.Values) (TicketList, error) {
	var ticketList TicketList
	req, err := getNewTicketListRequest(path, query)
	if err != nil {
		logger.Println(err)
		return ticketList, err
	}

	res, err := fetchApi(req)
	if err != nil {
		logger.Println(err)
		return ticketList, err
	}

	// parse fetched content
	if res.StatusCode >= 400 {
		err = checkErrorStatus(res.StatusCode)
		logger.Println(req.URL)
	} else {
		err = parseTicketListJson(res, &ticketList)
	}

	return ticketList, err

}

// retrives ticket with given id from api
func (ads *ApiDataSource) GetTicket(path string, query url.Values) (Ticket, error) {
	var ticket Ticket
	req, err := getNewTicketRequest(path)
	if err != nil {
		logger.Println(err)
		return ticket, err
	}

	res, err := fetchApi(req)
	if err != nil {
		logger.Println(err)
		return ticket, err
	}

	// parse fetched ticket
	if res.StatusCode >= 400 {
		err = checkErrorStatus(res.StatusCode)
	} else {
		err = parseTicketJson(res, &ticket, query)
	}

	return ticket, err
}

const initialPaginationParam = "page=%v&per_page=%v"

func getNewTicketListRequest(path string, query url.Values) (*http.Request, error) {
	req, err := getNewRequest(path)
	if err != nil {
		return nil, err
	}

	page := query.Get("page")
	perPage := query.Get("per_page")
	if page != "" && perPage != "" {
		req.URL.RawQuery = fmt.Sprintf(initialPaginationParam, page, perPage)
	} else if page != "" {
		req.URL.RawQuery = fmt.Sprintf(initialPaginationParam, page, itemLimit)
	} else if perPage != "" {
		req.URL.RawQuery = fmt.Sprintf(initialPaginationParam, 1, perPage)
	} else {
		req.URL.RawQuery = fmt.Sprintf(initialPaginationParam, 1, itemLimit)
	}

	return req, nil
}

func getNewTicketRequest(path string) (*http.Request, error) {
	req, err := getNewRequest(path)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = "include=users"

	return req, nil
}

func getNewRequest(path string) (*http.Request, error) {
	godotenv.Load()
	url := parseUrl(path)
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
		Timeout: time.Second * 5,
	}
	return client.Do(req)
}

const baseDomain = "https://%s.zendesk.com/api/v2"

func parseUrl(path string) string {
	godotenv.Load()
	subDomain := os.Getenv("API_SUBDOMAIN")
	domain := fmt.Sprintf(baseDomain, subDomain) // up until api/v2

	return domain + path

}

// switches through error status and returns appropriate error
func checkErrorStatus(code int) error {
	// api errors
	// hard coded because errors have different json format
	var err error
	switch code {
	case http.StatusUnauthorized:
		err = errors.New("unauthorized access: check your credentials")
	default:
		errMsg := fmt.Sprintf("there was an error with the API, Status Code %d", code)
		err = errors.New(errMsg)
	}
	return err
}

// parses json in passed in io.Reader, and stores data in passed in ticketList
func parseTicketListJson(r *http.Response, ticketList *TicketList) error {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(ticketList); err != nil {
		logger.Println(err)
		return err
	}

	// regex to acquire path to next and prev links
	re, err := regexp.Compile(`/tickets.json\?[\w=&]+`)
	if err != nil {
		logger.Fatalln(err)
	}

	// modify links to be processable by server.go
	if ticketList.NextPage != "" {
		ticketList.NextPage = strings.Replace(re.FindString(ticketList.NextPage), ".json", "", -1)
	}
	if ticketList.PreviousPage != "" {
		ticketList.PreviousPage = strings.Replace(re.FindString(ticketList.PreviousPage), ".json", "", -1)
	}

	// add page number
	query := r.Request.URL.Query()
	num := query.Get("page")
	if ticketList.NextPage != "" || ticketList.PreviousPage != "" {
		// if there are multiple pages
		if num != "" {
			// page num found in query
			ticketList.PageNum = num
		} else {
			// page num not found in query i.e. first page
			ticketList.PageNum = "1"
		}
	}

	ticketList.TicketCountLimit = itemLimit
	lastPageNum := ticketList.Count / ticketList.TicketCountLimit
	if ticketList.Count%ticketList.TicketCountLimit > 0 {
		lastPageNum += 1
	}
	ticketList.LastPageNum = lastPageNum

	return nil
}

func parseTicketJson(r *http.Response, ticket *Ticket, param url.Values) error {
	decoder := json.NewDecoder(r.Body)

	type User struct {
		Name string `json:"name"`
	}

	wrapper := struct {
		Ticket *Ticket `json:"ticket"`
		Users  []User
	}{
		ticket,
		[]User{},
	}
	if err := decoder.Decode(&wrapper); err != nil {
		logger.Println(err)
		return err
	}

	ticket.RequesterName = wrapper.Users[0].Name

	// set BackLink
	ticket.BackPage = param.Get("backPage")

	return nil
}
