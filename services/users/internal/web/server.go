package web

import (
	. "eda-users/internal/types"
	"errors"
	"fmt"
	"log"
	"net/http"
)

type Server struct {
	Db    IDatabase
	Kakfa IBroker
}

func NewServer(db IDatabase, kf IBroker) *Server {
	return &Server{Db: db, Kakfa: kf}
}

func (s Server) formValidator(r *http.Request) (string, string, error) {
	err := r.ParseForm()

	if err != nil {
		return "", "", errors.New("Invalid form data")
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		return "", "", errors.New("Username and password are required")
	}

	return username, password, nil
}

func (s Server) Register(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username, password, err := s.formValidator(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res := s.Db.Register(username, password)

	if res != nil {
		http.Error(w, res.Error(), http.StatusBadRequest)
		return
	}

	output := fmt.Sprintf("User %s registered successfully", username)

	log.Println(output)

	s.Kakfa.Send(output)

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(output))
}

func (s Server) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username, password, err := s.formValidator(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.Db.Login(username, password)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := CreateClaims()

	if err != nil {
		http.Error(w, "Error creating token", http.StatusInternalServerError)
		return
	}

	output := fmt.Sprintf("User %s successfully logged\n", username)

	err = s.Kakfa.Send(output)

	if err != nil {
		http.Error(w, "Error sending log to Kafka", http.StatusInternalServerError)
		return
	}

	log.Println(output)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(token))
}

func (s Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	mux := http.NewServeMux()

	mux.HandleFunc("/register", s.Register)
	mux.HandleFunc("/login", s.Login)

	mux.ServeHTTP(w, r)

}
