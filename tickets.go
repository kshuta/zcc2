package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// itemLimit sets the number of tickets to be displayed on a single page.
// test will not pass if itemLimit is less than 2, which is reasonable as it doesn't make sense to have a one ticket list page
const itemLimit = 25

// TicketList acts as a wrapper to Ticket to make unmarshalling the json easier.
// It also stores information such as links to the next and previous page for pagination
type TicketList struct {
	Tickets            []Ticket `json:"tickets"`
	Count              int      `json:"count"`
	NextPage           string   `json:"next_page"`
	PreviousPage       string   `json:"previous_page"`
	PageNum            int      // used to assign BackPage for Ticket
	LastPageNum        int      // used for ">>" button to skip to last page in index.html
	TicketDisplayLimit int      // used to generate query in index.html
}

// Ticket keeps track of information regarding a single ticket
type Ticket struct {
	Id            int64    `json:"id"`
	Subject       string   `json:"subject"`
	Description   string   `json:"description"`
	Status        string   `json:"status"`
	Priority      string   `json:"priority"`
	RequesterName string   // set by sideloaded user info
	BackPage      string   // used to create link to go back to the original page in ticket list view
	Tags          []string `json:"tags"`
}

// ApiDataSource is a type that implements server.
// It is used as a datasource that retrieves data from the zendesk api.
type ApiDataSource struct {
}

// GetTickets retrives tickets with given id from api
// path
func (ads *ApiDataSource) GetTickets(query url.Values) (TicketList, error) {
	var ticketList TicketList

	req, err := getNewTicketListRequest(query)
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

// GetTicket retrives ticket from the zendesk api, with the id in path
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

// initialPaginationParam is a string used to generate query parameter
// to fetch paginated results from the API.
const initialPaginationParam = "page=%v&per_page=%v"

// getNewTicketListReques creates a request to fetch tickets from the api
// The argument query is expected to have a "page" parameter to specify the pagination.
// if query doesn't have a "page", it will default to 1
func getNewTicketListRequest(query url.Values) (*http.Request, error) {
	req, err := getNewRequest("/tickets/")
	if err != nil {
		return nil, err
	}

	page := query.Get("page")

	if page != "" {
		req.URL.RawQuery = fmt.Sprintf(initialPaginationParam, page, itemLimit)
	} else {
		req.URL.RawQuery = fmt.Sprintf(initialPaginationParam, 1, itemLimit)
	}

	return req, nil
}

// getNewTicketRequest creates a request to fetch a ticket ffrom the api
// the argument path is expected to have the id of the ticket it's trying to retrieve.
// the request will have a query to sideload users when fetching a single ticket.
func getNewTicketRequest(path string) (*http.Request, error) {
	req, err := getNewRequest(path)
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = "include=users"

	return req, nil
}

// getNewRequest generates a request to fetch ticket/s from the api.
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

// fetchApi will invoke the get request with the request passed in.
func fetchApi(req *http.Request) (*http.Response, error) {
	client := http.Client{
		Timeout: time.Second * 5,
	}
	return client.Do(req)
}

// baseDomain is a string used to generate the url
// to fetch the api.
const baseDomain = "https://%s.zendesk.com/api/v2"

func parseUrl(path string) string {
	godotenv.Load()
	subDomain := os.Getenv("API_SUBDOMAIN")
	domain := fmt.Sprintf(baseDomain, subDomain)

	return domain + path

}

// checkErrorStatus switches through error status and returns appropriate error
func checkErrorStatus(code int) error {
	// errors are hard coded because errors have different json format
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

// parseTicketListJson parses json from passed in response, and stores data in passed in ticketList
func parseTicketListJson(r *http.Response, ticketList *TicketList) error {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(ticketList); err != nil {
		logger.Println(err)
		return err
	}

	// regex to acquire query in NextPage and PreviousPage
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

	setCustomTicketListFields(r, ticketList)

	return nil
}

// setCustomTicketListFields sets the custom TicketListFields
func setCustomTicketListFields(r *http.Response, ticketList *TicketList) {
	// set PageNum
	query := r.Request.URL.Query()
	num := query.Get("page")
	if ticketList.NextPage != "" || ticketList.PreviousPage != "" {
		// if there are multiple pages
		if num != "" {
			// page num found in query
			i64, _ := strconv.ParseInt(num, 10, 32)
			ticketList.PageNum = int(i64)
		}
	}

	ticketList.TicketDisplayLimit = itemLimit

	lastPageNum := ticketList.Count / ticketList.TicketDisplayLimit
	if ticketList.Count%ticketList.TicketDisplayLimit > 0 {
		lastPageNum += 1
	}
	ticketList.LastPageNum = lastPageNum
}

// parseTicketJson parses json from passed in response, and stores data in passed in ticket
func parseTicketJson(r *http.Response, ticket *Ticket, param url.Values) error {
	decoder := json.NewDecoder(r.Body)

	type User struct {
		Name string `json:"name"`
	}

	// wrapper to make unmarshalling json easier
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

	ticket.BackPage = param.Get("backPage")

	return nil
}
