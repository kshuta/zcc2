package main

import (
	"errors"
	"fmt"
	"html"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

type StubDataSource struct {
	err       error
	ticketNum int
}

// Fucntion to mock the functionality of the data source with tickets
func (ds *StubDataSource) GetTickets(params string) (TicketList, error) {
	// ds.err will not be nil when testing for errors
	if ds.err != nil {
		return TicketList{}, ds.err
	}
	tickets := make([]Ticket, 0)

	// create items
	for i := 0; i < ds.ticketNum; i++ {
		tickets = append(tickets, Ticket{Subject: fmt.Sprintf("Test %d", i), Id: int64(i + 1), Status: "open"})
	}

	// instantiate TicketList
	tl := TicketList{}

	// for tests, there are only cases where the maximum number of pages are two
	// meaning if the ticketNum is bigger than the itemLimit, there would only be either Next or Prev
	if ds.ticketNum > itemLimit {
		if params == "next" {
			logger.Println("int tickets/next")
			tl.PreviousPage = "/prev"
			tl.Tickets = tickets[itemLimit:]
		} else {
			// prev or nothing
			tl.NextPage = "/next"
			tl.Tickets = tickets[:itemLimit]
		}
	} else {
		tl.Tickets = tickets
	}

	return tl, nil

}

func (ds *StubDataSource) GetTicket(param string) (Ticket, error) {
	if ds.err != nil {
		return Ticket{}, ds.err
	}
	return Ticket{}, nil
}

func TestIndex(t *testing.T) {
	ds := StubDataSource{}
	// default ticketNum
	ds.ticketNum = itemLimit - 1
	server := server{
		source: &ds,
	}

	t.Run("successful call to index without no Next/Prev", func(t *testing.T) {
		req := getNewTestRequest("/tickets/")
		res := httptest.NewRecorder()

		server.ServeHTTP(res, req)

		// checks response body for the word "error"
		assertNoError(t, res.Body.String())

		// assert no next or prev button
		expression := "<li.*Next.*</li>"
		assertNoExpression(t, res.Body.String(), expression, "next button")
		expression = "<li.*Previous.*</li>"
		assertNoExpression(t, res.Body.String(), expression, "previous button")
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
		ds.ticketNum = itemLimit + 1
		ds.err = nil
		req := getNewTestRequest("/tickets/")
		res := httptest.NewRecorder()

		server.ServeHTTP(res, req)

		// assert no error in response body
		assertNoError(t, res.Body.String())

		// check for the element <li> with the word "Next" inside
		expression := "<li.*Next.*</li>"
		assertExpression(t, res.Body.String(), expression, "next button")
		// assert no list element with word "Previous" is found
		expression = "<li.*Previous.*</li>"
		assertNoExpression(t, res.Body.String(), expression, "previous button")
		// assert the number of ticket on the page
		expression = "</tr>"
		assertElementCount(t, res.Body.String(), expression, itemLimit)

	})

	t.Run("display 2 paged index second page", func(t *testing.T) {
		ds.ticketNum = itemLimit + 1
		ds.err = nil
		// TODO: change the hardcoded param
		req := getNewTestRequest("/tickets/?next")
		res := httptest.NewRecorder()

		server.ServeHTTP(res, req)

		// assert no error in response body
		assertNoError(t, res.Body.String())

		// check for the element <li> with the word "Previous" inside
		expression := "<li.*Previous.*</li>"
		assertExpression(t, res.Body.String(), expression, "previous button")
		// assert no list element with word "Next" is found
		expression = "<li.*Next.*</li>"
		assertNoExpression(t, res.Body.String(), expression, "previous button")
		// assert the number of tickets on the page
		expression = "</tr>"
		assertElementCount(t, res.Body.String(), expression, ds.ticketNum-itemLimit)
	})

}

// gets new request with passed in url
func getNewTestRequest(url string) *http.Request {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		logger.Fatal(err)
	}
	return req
}

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

// checks response body for the keyword "error"
func assertNoError(t testing.TB, body string) {
	t.Helper()
	assertNoExpression(t, body, "error:.*", "error")
}

// checks response body doesn't have the expresion passed in
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
