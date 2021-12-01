package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
)

type server struct {
	source DataSource
}

type DataSource interface {
	GetTickets(params string) (TicketList, error)
	GetTicket(path string) (Ticket, error)
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger.Println("Server HTTP")
	ticketDetail, err := regexp.Compile(`/tickets/[0-9]+$`)
	if err != nil {
		logger.Fatal(err)
	}

	if ticketDetail.MatchString(r.URL.Path) {
		// ticket detail
		s.detailHandlerFunc(w, r, r.URL.RawQuery)
	} else {
		s.indexHandlerFunc(w, r, r.URL.RawQuery)
	}
}

func (s *server) indexHandlerFunc(w http.ResponseWriter, r *http.Request, params string) {
	logger.Println("Serving Index Handler")
	tickets, err := s.source.GetTickets(params)
	if err != nil {
		logger.Println(err)
		s.errorHandlerFunc(w, r, err)
		return
	}
	t := template.New("index.html")
	t = template.Must(t.ParseFiles("templates/index.html"))
	t.Execute(w, tickets)
}

// handler function
func (s *server) detailHandlerFunc(w http.ResponseWriter, r *http.Request, path string) {
	logger.Println("Serving Detail Handler")
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
	logger.Println("errorHandlerFunc envoked")
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
