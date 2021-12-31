package main

import (
	"errors"
	"fmt"
	"html"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"testing"

	"github.com/kshuta/zcc2/tickets"
)

// StubDataSource imitiates the api for tests
type StubDataSource struct {
	err        error
	ticketNum  int
	testTicket tickets.Ticket
}

func TestMain(m *testing.M) {
	// set different output path for custom logger
	f, err := os.OpenFile("logfile", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err == nil {
		logger.SetOutput(f)
	}

	os.Exit(m.Run())
}

func (ds *StubDataSource) GetTickets(params url.Values) (tickets.TicketList, error) {
	// ds.err will not be nil when testing for errors
	if ds.err != nil {
		return tickets.TicketList{}, ds.err
	}
	ticks := make([]tickets.Ticket, 0)

	// create tickets
	for i := 0; i < ds.ticketNum; i++ {
		ticks = append(ticks, tickets.Ticket{Subject: fmt.Sprintf("Test %d", i), Id: int64(i + 1), Status: "open"})
	}

	// instantiate TicketList
	tl := tickets.TicketList{}
	tl.Count = ds.ticketNum
	tl.TicketDisplayLimit = tickets.ItemLimit

	// for tests, there are only cases where the maximum number of pages are two
	// meaning if the ticketNum is bigger than the itemLimit, there would only be either Next or Prev
	if ds.ticketNum > tickets.ItemLimit {
		if params.Get("page") == "next" {
			tl.PreviousPage = "/prev"
			tl.Tickets = ticks[tickets.ItemLimit:]
			tl.PageNum = 2
			tl.LastPageNum = 2
		} else {
			// prev or nothing
			tl.NextPage = "/next"
			tl.Tickets = ticks[:tickets.ItemLimit]
			tl.PageNum = 1
			tl.LastPageNum = 2
		}
	} else {
		tl.Tickets = ticks
	}

	return tl, nil

}

func (ds *StubDataSource) GetTicket(path string, param url.Values) (tickets.Ticket, error) {
	if ds.err != nil {
		return tickets.Ticket{}, ds.err
	}

	return ds.testTicket, nil
}

func TestGetIndex(t *testing.T) {
	ds := StubDataSource{}
	// default ticketNum
	ds.ticketNum = tickets.ItemLimit - 1
	server := server{
		Source: &ds,
	}

	t.Run("successful call to single paged index page", func(t *testing.T) {
		req := getNewTestRequest("/tickets/")
		res := httptest.NewRecorder()

		server.ServeHTTP(res, req)

		// checks response body for the word "error"
		assertNoError(t, res.Body.String())

		// assert next and prev button are not displayed
		expression := `<li.*Next.*</li>`
		assertNoExpression(t, res.Body.String(), expression, "next button")
		expression = `<li.*Previous.*</li>`
		assertNoExpression(t, res.Body.String(), expression, "previous button")

		// assert no text regarding page number
		expression = `<p>Showing page [0-9] of [0-9] pages.</p>`
		assertNoExpression(t, res.Body.String(), expression, "page number")

	})

	t.Run("displays error", func(t *testing.T) {
		// set up error
		errString := "error: couldn't authenticate you"
		ds.err = errors.New(errString)

		req := getNewTestRequest("/tickets/")
		res := httptest.NewRecorder()

		server.ServeHTTP(res, req)

		// check status code is StatusBadRequest
		if res.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, res.Code)
		}

		// check response body for errString
		assertExpression(t, res.Body.String(), fmt.Sprintf(".*%s.*", html.EscapeString(errString)), errString)
	})

	t.Run("displays 2 paged index first page", func(t *testing.T) {
		ds.ticketNum = tickets.ItemLimit + 1
		ds.err = nil
		req := getNewTestRequest("/tickets/")
		res := httptest.NewRecorder()

		server.ServeHTTP(res, req)

		// assert no error in response body
		assertNoError(t, res.Body.String())

		// assert next is enabled and prev is disabled
		expression := `<li class=".*disabled".*Next.*</li>`
		assertNoExpression(t, res.Body.String(), expression, "next button")
		expression = `<li class=".*disabled".*Previous.*</li>`
		assertExpression(t, res.Body.String(), expression, "previous button")

		// assert the number of ticket on the page
		expression = `</tr>`
		assertElementCount(t, res.Body.String(), expression, tickets.ItemLimit)

		// assert page number being displayed
		expression = `<p>Showing page 1 of [0-9] pages.</p>`
		assertExpression(t, res.Body.String(), expression, "page number")

	})

	t.Run("display 2 paged index second page", func(t *testing.T) {
		ds.ticketNum = tickets.ItemLimit + 1
		ds.err = nil
		// TODO: change the hardcoded param
		req := getNewTestRequest("/tickets/?page=next")
		res := httptest.NewRecorder()

		server.ServeHTTP(res, req)

		// assert no error in response body
		assertNoError(t, res.Body.String())

		// assert next is disabled and prev is enabled
		expression := `<li class=".*disabled".*Next.*</li>`
		assertExpression(t, res.Body.String(), expression, "next button")
		expression = `<li class=".*disabled".*Previous.*</li>`
		assertNoExpression(t, res.Body.String(), expression, "previous button")

		// assert the number of tickets on the page
		expression = `</tr>`
		assertElementCount(t, res.Body.String(), expression, ds.ticketNum-tickets.ItemLimit)

		// assert page number being displayed
		expression = `<p>Showing page 2 of [0-9] pages.</p>`
		assertExpression(t, res.Body.String(), expression, "page number")
	})

	t.Run("zero results", func(t *testing.T) {
		ds.ticketNum = 0
		ds.err = nil
		req := getNewTestRequest("/tickets/")
		res := httptest.NewRecorder()

		server.ServeHTTP(res, req)

		// assert next and prev button are not displaed
		expression := `<li.*Next.*</li>`
		assertNoExpression(t, res.Body.String(), expression, "next button")
		expression = `<li.*Previous.*</li>`
		assertNoExpression(t, res.Body.String(), expression, "previous button")

		// assert message "No tickets to show" is displayed
		expression = `<tr>.*No tickets to show.*</tr>`
		assertExpression(t, res.Body.String(), expression, "no ticket to show message")
	})
}

func TestGetDetail(t *testing.T) {
	ds := StubDataSource{}
	// default ticketNum
	ds.ticketNum = tickets.ItemLimit - 1
	server := server{
		Source: &ds,
	}

	t.Run("successfully retrieve ticket", func(t *testing.T) {
		// setup test ticket
		ticket := tickets.Ticket{
			Id:            1,
			Subject:       "Test tickets.Ticket",
			Description:   "This is a test ticket",
			Status:        "Closed",
			Priority:      "",
			RequesterName: "Shuta Kumada",
			Tags: []string{
				"tag1",
				"tag2",
				"tag4",
			},
		}
		ds.testTicket = ticket

		req := getNewTestRequest("/tickets/1")
		res := httptest.NewRecorder()

		server.ServeHTTP(res, req)

		assertNoError(t, res.Body.String())

		// check if content is displayed on page
		expression := fmt.Sprintf(`<p>Requester: %s</p>`, ticket.RequesterName)
		assertExpression(t, res.Body.String(), expression, "name")

		// check backlink is working
		expression = `<a href=.*>Go Back To List</a>`
		assertExpression(t, res.Body.String(), expression, "back link")

	})

	t.Run("displays error", func(t *testing.T) {
		errMsg := "api is down"
		ds.err = errors.New(errMsg)
		req := getNewTestRequest("/tickets/1")
		res := httptest.NewRecorder()

		server.ServeHTTP(res, req)

		assertExpression(t, res.Body.String(), errMsg, "error")
	})

}

// getNewTestRequest gets new request with passed in url
func getNewTestRequest(url string) *http.Request {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		logger.Fatal(err)
	}
	return req
}

// assertElementCount will count the number of exprssions within the passed in body
// and envokes error if the count doesn't match
func assertElementCount(t testing.TB, body, expression string, count int) {
	r, err := regexp.Compile(expression)
	if err != nil {
		t.Fatal(err)
	}

	l := len(r.FindAllString(body, -1))
	if l != count+1 {
		t.Errorf("number of items doesn't match: got %d, want %d", l, count)
	}
}

// assertNoError is a convinience function for checking errors
func assertNoError(t testing.TB, body string) {
	t.Helper()
	errExpression := `<div class="alert alert-warning`
	assertNoExpression(t, body, errExpression, "error")
}

// assertNoExpresion asserts that the response body doesn't match the regular expression passed in
func assertNoExpression(t testing.TB, body, expression, unexpectedString string) {
	t.Helper()
	r, err := regexp.Compile(expression)
	if err != nil {
		t.Fatal(err)
	}

	if r.MatchString(body) {
		t.Errorf("unexpected %s found", unexpectedString)
	}
}

// assertExpresion asserts that the response body matches the regular expression passed in
func assertExpression(t testing.TB, body, expression, expectedString string) {
	t.Helper()

	r, err := regexp.Compile(expression)
	if err != nil {
		t.Fatal(err)
	}

	// if there doesn't exist an error
	if !r.MatchString(body) {
		t.Errorf("expected %s, not found", expectedString)
	}
}
