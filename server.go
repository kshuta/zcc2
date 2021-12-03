package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
)

// struct that implements Handler
// different data source can be used for tests, making them independent
// from the api
type server struct {
	source DataSource
}

// DataSource is an interface that acts as a data source for the server.
type DataSource interface {
	GetTickets(query url.Values) (TicketList, error)
	GetTicket(path string, query url.Values) (Ticket, error)
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger.Println("Serving HTTP")
	ticketDetail, err := regexp.Compile(`/tickets/[0-9]+[a-z]*$`)
	if err != nil {
		logger.Fatal(err)
	}

	if ticketDetail.MatchString(r.URL.Path) {
		s.detailHandlerFunc(w, r, r.URL.Path, r.URL.Query())
	} else {
		s.indexHandlerFunc(w, r, r.URL.Query())
	}
}

// indexHandlerFunc displays the ticket list view
// will envoke error handler if GetTickets returns err
func (s *server) indexHandlerFunc(w http.ResponseWriter, r *http.Request, query url.Values) {
	logger.Println("invoking Index Handler")
	tickets, err := s.source.GetTickets(query)
	if err != nil {
		logger.Println(err)
		s.errorHandlerFunc(w, r, err)
		return
	}
	t := template.New("layout")
	t = template.Must(t.ParseFiles("templates/layout.html", "templates/index.html"))
	t.Execute(w, tickets)
}

// detailHandlerFunc displays the ticket detail view
// will encokde error handler if GetTicket returns err
func (s *server) detailHandlerFunc(w http.ResponseWriter, r *http.Request, path string, query url.Values) {
	logger.Println("invoking Detail Handler")
	ticket, err := s.source.GetTicket(path, query)
	if err != nil {
		s.errorHandlerFunc(w, r, err)
		return
	}
	t := template.New("layout")
	t = template.Must(t.ParseFiles("templates/layout.html", "templates/detail.html"))
	w.WriteHeader(http.StatusAccepted)
	t.Execute(w, ticket)
}

// errorHandlerFunc displays the error page
func (s *server) errorHandlerFunc(w http.ResponseWriter, r *http.Request, err error) {
	logger.Println("invoking error handler")
	w.WriteHeader(http.StatusBadRequest)
	t := template.New("layout")
	t = template.Must(t.ParseFiles("templates/layout.html", "templates/errors.html"))
	t.Execute(w, err.Error())
}

// custom logger
var logger = log.New(os.Stderr, "logger: ", log.LstdFlags|log.Lshortfile)

func main() {
	port := flag.String("port", ":5000", "Port number")
	flag.Parse()

	s := server{&ApiDataSource{}}

	fmt.Printf("serving at port %v", *port)
	logger.Fatalln(http.ListenAndServe(*port, &s))
}
