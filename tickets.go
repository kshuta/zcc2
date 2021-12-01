package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
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
	PageNum          int
	LastPageNum      int
	TicketCountLimit int
}

// pretty useless to store all the information when in reality you only need id, subject and url for list view
// but it makes it easier to use json methods, and there will only be 25 tickets at most in memory, so it shouldn't be
// to much of a problem
type Ticket struct {
	Id          int64  `json:"id"`
	Subject     string `json:"subject"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Priority    string `json:"priority"`
	User        struct {
		Name string
	}
}

// retrives tickets with given id from api
func (ads *ApiDataSource) GetTickets(params string) (TicketList, error) {
	var ticketList TicketList
	req, err := getNewTicketListRequest(params)
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
	} else {
		// parse response body (json)
		err = parseTicketListJson(res, &ticketList)
	}

	return ticketList, err

}

// retrives ticket with given id from api
func (ads *ApiDataSource) GetTicket(path string) (Ticket, error) {
	return Ticket{}, nil
}

const initialPaginationParam = "?page=1&per_page=%d"
const ticketListPath = "/tickets/"

func getNewTicketListRequest(params string) (*http.Request, error) {
	req, err := getNewRequest(ticketListPath)
	if err != nil {
		return nil, err
	}

	if params == "" {
		req.URL.RawQuery = fmt.Sprintf(initialPaginationParam, itemLimit)
	} else {
		req.URL.RawQuery = params
	}
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
		Timeout: time.Second * 10,
	}
	return client.Do(req)
}

const baseDomain = "https://%s.zendesk.com/api/v2"
const ticketPath = "/tickets/%d?include=users"

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
		err = errors.New("unauthorized access: Check your credentials")
	case http.StatusBadGateway:
		err = errors.New("badgateway: The API might be down. Try again in a while")
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
		logger.Println(ticketList.NextPage)
	}
	if ticketList.PreviousPage != "" {
		ticketList.PreviousPage = strings.Replace(re.FindString(ticketList.PreviousPage), ".json", "", -1)
		logger.Println(ticketList.PreviousPage)
	}

	// add page number
	query := r.Request.URL.Query()
	num, ok := query["page"]
	if ticketList.NextPage != "" || ticketList.PreviousPage != "" {
		// if there are multiple pages
		if ok {
			// page num found in query
			i64, err := strconv.ParseInt(num[0], 10, 32)
			if err != nil {
				logger.Println(err)
				return err
			}

			ticketList.PageNum = int(i64)
		} else {
			// page num not found in query i.e. first page
			ticketList.PageNum = 1
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
