package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	ROUTINE_POOL_MAX = 2
)

type ActiveServer struct {
	Port      uint16
	Scheduler *Scheduler
	Jobs      map[string]string
	DB        *Database
}

type Scheduler struct {
	mu    sync.Mutex
	batch map[string]time.Time
	done  chan bool
}

func (as *ActiveServer) NewScheduler() {
	as.Scheduler.done = make(chan bool)
}

func (s *Scheduler) ProcessBatch(db *Database) {
	s.mu.Lock()
	defer s.mu.Unlock()
	db.Update(s.batch)
	s.batch = make(map[string]time.Time)
}

func (s *Scheduler) Run(db *Database) {
	for routine_pool := 0; routine_pool < ROUTINE_POOL_MAX; routine_pool++ {
		go func() {
			limiter := time.Tick(500 * time.Millisecond)
			for {
				select {
				case <-limiter:
					s.ProcessBatch(db)
				case <-s.done:
					break
				}
			}
		}()
	}
}

func (s *Scheduler) Add(resource_id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.batch[resource_id] = time.Time{} //zero time to differentiate between processed and not
}

type Job struct {
	Id         int
	ResourceId string
	Ttl        time.Time
}

type ApiResponse struct {
	NewTTL time.Time
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

func (ws *ActiveServer) UpdateResource(w http.ResponseWriter, req *http.Request) {
	url := strings.Split(req.RequestURI, " ")
	log.Printf("[%s]%s", req.Method, url[0])
	switch req.Method {
	case http.MethodPut:
		resource_id := req.URL.Query().Get("resource_id")
		// Assume no error when parsing the uint16 value as previously validated
		ws.AddToBatch(resource_id)
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

func (as *ActiveServer) Run() {
	as.NewScheduler()
	as.RunScheduler()
	http.HandleFunc("/:resource_id", as.UpdateResource)
	http.HandleFunc("/dealers/:resource_id", as.GetResource)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(as.Port)), nil))
}
