package app

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/sekaiichi/gosql/pkg/customers"
)

// Server представляет собой логический сервер нашего приложения.
type Server struct {
	mux          *http.ServeMux
	customersSvc *customers.Service
}

// NewServer ...
func NewServer(mux *http.ServeMux, customersSvc *customers.Service) *Server {
	return &Server{mux: mux, customersSvc: customersSvc}
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mux.ServeHTTP(writer, request)
}

// Init инициализирует сервер (регистрирует все Handler'ы)
func (s *Server) Init()  {
	s.mux.HandleFunc("/customers.getById", s.handleGetCustomerByID)
	s.mux.HandleFunc("/customers.getAll", s.handleGetAll)
	s.mux.HandleFunc("/customers.getAllActive", s.handleGetAllActive)
	s.mux.HandleFunc("/customers.save", s.handleSave)
	s.mux.HandleFunc("/customers.removeById", s.handleRemoveByID)
	s.mux.HandleFunc("/customers.blockById", s.handleBlockByID)
	s.mux.HandleFunc("/customers.unblockById", s.handleUnblockByID)
}

func (s *Server) handleGetCustomerByID(writer http.ResponseWriter, request *http.Request) {
	idParam := request.URL.Query().Get("id")

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	item, err := s.customersSvc.ByID(request.Context(), id)
	if errors.Is(err, customers.ErrNotFound) {
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(item)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}
}

func (s *Server) handleGetAll(writer http.ResponseWriter, request *http.Request)  {
	item, err := s.customersSvc.All(request.Context())
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	data, err := json.Marshal(item)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}
}

func (s *Server) handleGetAllActive(writer http.ResponseWriter, request *http.Request)  {
	item, err := s.customersSvc.AllActive(request.Context())
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	data, err := json.Marshal(item)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}
}

func (s *Server) handleSave(writer http.ResponseWriter, request *http.Request)  {
	customerID := request.FormValue("id")
	customerName := request.FormValue("name")
	customerPhone := request.FormValue("phone")

	convID, err := strconv.ParseInt(customerID, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if customerName == "" && customerPhone == "" {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	item := &customers.Customer{
		ID:      convID,
		Name:    customerName,
		Phone:   customerPhone,
		Active:  true,
		Created: time.Now(),
	}

	newCustomer, err := s.customersSvc.Save(request.Context(), item)
	if err != nil {
		if err == customers.ErrNotFound {
			log.Print(err)
			http.Error(writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(newCustomer)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}
}

func (s *Server) handleRemoveByID(writer http.ResponseWriter, request *http.Request)  {
	customerID := request.URL.Query().Get("id")
	convID, err := strconv.ParseInt(customerID, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	removedCustomer, err := s.customersSvc.RemoveByID(request.Context(), convID)
	if err != nil {
		if err == customers.ErrNotFound {
			log.Print(err)
			http.Error(writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(removedCustomer)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}	
}

func (s *Server) handleBlockByID(writer http.ResponseWriter, request *http.Request)  {
	customerID := request.URL.Query().Get("id")
	convID, err := strconv.ParseInt(customerID, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}	
	
	blockedUser, err := s.customersSvc.BlockUser(request.Context(), convID)
	if err != nil {
		if err == customers.ErrNotFound {
			log.Print(err)
			http.Error(writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(blockedUser)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}	
}

func (s *Server) handleUnblockByID(writer http.ResponseWriter, request *http.Request)  {
	customerID := request.URL.Query().Get("id")
	convID, err := strconv.ParseInt(customerID, 10, 64)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}	
	
	unblockedUser, err := s.customersSvc.UnblockUser(request.Context(), convID)
	if err != nil {
		if err == customers.ErrNotFound {
			log.Print(err)
			http.Error(writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(unblockedUser)
	if err != nil {
		log.Print(err)
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(data)
	if err != nil {
		log.Print(err)
	}	
}