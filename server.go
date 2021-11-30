package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

type server struct {
	source DataSource
}

type DataSource interface {
	GetTickets(path string) (TicketList, error)
	GetTicket(path string) (Ticket, error)
}

// should not be less than 2 for test to pass
const itemLimit = 5
const paginationPath = "/tickets/page[size]=%d"

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	ticketDetail, err :=  

	if strings.HasPrefix(r.URL.Path, "/tickets/[0-9]*") {
		// ticket detail
		s.detailHandlerFunc(w, r, r.URL.Path)
	} else {
		if strings.HasPrefix(r.URL.Path, "/tickets/$") {
			path := fmt.Sprintf(paginationPath, itemLimit)
			s.indexHandlerFunc(w, r, path)
		} else {
			s.indexHandlerFunc(w, r, r.URL.Path)
		}
	}
}

func (s *server) indexHandlerFunc(w http.ResponseWriter, r *http.Request, path string) {
	tickets, err := s.source.GetTickets(path)
	if err != nil {
		s.errorHandlerFunc(w, r, err)
		return
	}
	t := template.New("index.html")
	t = template.Must(t.ParseFiles("templates/index.html"))
	t.Execute(w, tickets)
}

// handler function
func (s *server) detailHandlerFunc(w http.ResponseWriter, r *http.Request, path string) {
	ticket, err := s.source.GetTicket(path)
	if err != nil {
		s.errorHandlerFunc(w, r, err)
		return
	}
	t := template.New("detail.html")
	t = template.Must(t.ParseFiles("templates/detail.html"))
	w.WriteHeader(http.StatusAccepted)
	t.Execute(w, ticket)
}

// handler for displaying error messages
func (s *server) errorHandlerFunc(w http.ResponseWriter, r *http.Request, err error) {
	w.WriteHeader(http.StatusBadRequest)
	t := template.New("errors.html")
	t = template.Must(t.ParseFiles("templates/errors.html"))
	t.Execute(w, err.Error())
}

var logger = log.New(os.Stderr, "logger: ", log.LstdFlags|log.Lshortfile)

func main() {
	s := server{&ApiDataSource{}}
	logger.Fatalln(http.ListenAndServe(":5000", &s))
}
