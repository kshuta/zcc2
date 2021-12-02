package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
)

type server struct {
	source DataSource
}

type DataSource interface {
	GetTickets(path string, query url.Values) (TicketList, error)
	GetTicket(path string, query url.Values) (Ticket, error)
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger.Println("Server HTTP")
	ticketDetail, err := regexp.Compile(`/tickets/[0-9]+[a-z]*$`)
	if err != nil {
		logger.Fatal(err)
	}

	if ticketDetail.MatchString(r.URL.Path) {
		s.detailHandlerFunc(w, r, r.URL.Path, r.URL.Query())
	} else {
		s.indexHandlerFunc(w, r, r.URL.Path, r.URL.Query())
	}
}

func (s *server) indexHandlerFunc(w http.ResponseWriter, r *http.Request, path string, query url.Values) {
	logger.Println("Serving Index Handler")
	tickets, err := s.source.GetTickets(path, query)
	if err != nil {
		logger.Println(err)
		s.errorHandlerFunc(w, r, err)
		return
	}
	t := template.New("layout")
	t = template.Must(t.ParseFiles("templates/layout.html", "templates/index.html"))
	t.Execute(w, tickets)
}

// handler function
func (s *server) detailHandlerFunc(w http.ResponseWriter, r *http.Request, path string, query url.Values) {
	logger.Println("Serving Detail Handler")
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

// handler for displaying error messages
func (s *server) errorHandlerFunc(w http.ResponseWriter, r *http.Request, err error) {
	logger.Println("errorHandlerFunc envoked")
	w.WriteHeader(http.StatusBadRequest)
	t := template.New("layout")
	t = template.Must(t.ParseFiles("templates/layout.html", "templates/errors.html"))
	t.Execute(w, err.Error())
}

var logger = log.New(os.Stderr, "logger: ", log.LstdFlags|log.Lshortfile)

func main() {
	port := flag.String("port", ":5000", "Port number")

	flag.Parse()
	s := server{&ApiDataSource{}}
	logger.Print(*port)
	logger.Fatalln(http.ListenAndServe(*port, &s))
}
