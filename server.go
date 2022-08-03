package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	ROUTINE_POOL_MAX = 2 // number of worker routines
)

type ActiveServer struct {
	Port      uint16
	Scheduler *Scheduler
	Jobs      map[string]string
	DB        *Database
}

func (as *ActiveServer) NewScheduler() {
	as.Scheduler.done = make(chan bool)
}

func NewActiveServer(port uint16) *ActiveServer {
	return &ActiveServer{
		Port: port,
	}
}

func (as *ActiveServer) Schedule(resource_id string) {
	as.Scheduler.Add(resource_id)
}

func (as *ActiveServer) AddToBatch(resource_id string) {
	as.Schedule(resource_id)
}

func (as *ActiveServer) AddFromBatch(resource_id string) (rid time.Time) {
	// Assume it's always available
	rid = as.DB.Select(resource_id)
	return
}

func (as *ActiveServer) GetResource(w http.ResponseWriter, req *http.Request) {
	url := strings.Split(req.RequestURI, " ")
	log.Printf("[%s]%s", req.Method, url[0])
	switch req.Method {
	case http.MethodGet:
		resource_id := req.URL.Query().Get("resource_id")
		// Assume no error when parsing the uint16 value as previously validated
		ttl := as.AddFromBatch(resource_id)
		w.Header().Add("Content-Type", "application/json")
		m, _ := json.Marshal(struct {
			NewTtl time.Time `json:"new_ttl"`
		}{
			NewTtl: ttl,
		})
		io.WriteString(w, string(m[:]))
	default:
		log.Printf("ERROR: Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (as *ActiveServer) UpdateResource(w http.ResponseWriter, req *http.Request) {
	url := strings.Split(req.RequestURI, " ")
	log.Printf("[%s]%s", req.Method, url[0])
	switch req.Method {
	case http.MethodPut:
		resource_id := req.URL.Query().Get("resource_id")
		// Assume no error when parsing the uint16 value as previously validated
		as.AddToBatch(resource_id)
		w.Header().Add("Content-Type", "application/json")
		m, _ := json.Marshal(struct {
			Status     string `json:"status"`
			ResourceId string `json:"resource_id"`
		}{
			Status:     "success",
			ResourceId: resource_id,
		})
		io.WriteString(w, string(m[:]))
	default:
		log.Printf("ERROR: Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (as *ActiveServer) RunScheduler() {
	as.Scheduler.Run(as.DB)
}

func (as *ActiveServer) Terminate() {
	// Termination request is being handles here
	// Proper clean up goes here
	as.Scheduler.Stop()
	as.DB.Close()
	log.Println("ActiveServer has terminated")
	os.Exit(1)
}

func (as *ActiveServer) RegisterHandlers() {
	go func() {
		sigs := make(chan os.Signal, 1)
		defer close(sigs)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-sigs:
			as.Terminate()
		case <-as.Scheduler.done:
			return
		}
	}()
}

func (as *ActiveServer) Run() {
	as.RegisterHandlers()
	// Create and run a new scheduler
	as.NewScheduler()
	as.RunScheduler()
	// Stop scheduler upon terminate
	defer as.Scheduler.Stop()
	// Close DB connection
	defer as.DB.Close()
	http.HandleFunc("/:resource_id", as.UpdateResource)
	http.HandleFunc("/dealers/:resource_id", as.GetResource)
	log.Println("Active server is live")
	log.Fatal(http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(as.Port)), nil))
}
