package server

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/NotSoFancyName/SimpleWebServer/persistance"
	"github.com/gorilla/mux"
)

const (
	readTimeout    = 10 * time.Second
	writeTimeout   = 10 * time.Second
	maxHeaderBytes = 1 << 20

	apiURL = "/api/block/{block}/total"
)

const (
	cacheMaxSize   = 10000
	expirationTime = 24 * time.Hour
)

var (
	wieToEthRatio = big.NewFloat(math.Pow10(-18))
)

type Server struct {
	server  *http.Server
	cache   *cache
	querier *persistance.DBQuerier
	stop    chan struct{}
}

func NewServer(stop chan struct{}, port int) *Server {
	r := mux.NewRouter()
	server := &Server{
		server: &http.Server{
			Addr:           ":" + strconv.FormatInt(int64(port), 10),
			Handler:        r,
			ReadTimeout:    readTimeout,
			WriteTimeout:   writeTimeout,
			MaxHeaderBytes: maxHeaderBytes,
		}, 
		cache: NewCache(cacheMaxSize, expirationTime), 
		querier: persistance.NewDBQuerier(),
		stop: stop,
	}
	r.HandleFunc(apiURL, server.getBlockInfo).Methods(http.MethodGet)
	return server
}

func (s *Server) Run(errs chan<- error) {
	log.Printf("Running server on port %v", s.server.Addr)
	go func() {
		<-s.stop
		if err := s.server.Shutdown(context.Background()); err != nil {
			log.Printf("Failed to shutdown server properly: %v", err)
		}
		if err := s.querier.Shutdown(); err != nil {
			log.Printf("Failed to shutdown DB properly: %v", err)
		} 
		log.Println("Server is shut")
		s.stop <- struct{}{}
	}()
	errs <- s.server.ListenAndServe()
}

type blockInfo struct {
	Transactions int     `json:"transactions"`
	Amount       float64 `json:"amount"`
}

func (s *Server) getBlockInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	blockNum, err := strconv.Atoi(mux.Vars(r)["block"])
	if err != nil {
		log.Printf("Failed to parse block number: %v \n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("Trying to calculate total transactions value for block number: %v\n", blockNum)

	count, err := queryTransactionsCount(blockNum)
	if err != nil {
		log.Printf("Failed to get transaction count %v: %v", blockNum, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	bi, presentInCache := s.cache.get(blockNum)
	if !presentInCache || (presentInCache && bi.Transactions != count) {
		dbInfo, presentInDB := s.querier.Get(blockNum)
		if !presentInDB || (presentInDB && dbInfo.Transactions != count) {
			bi, err = queryBlockInfo(blockNum)
			if err != nil {
				log.Printf("Failed to get info for block number %v: %v \n", blockNum, err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			s.cache.put(blockNum, bi)
			s.querier.Put(blockNum, bi.Transactions, bi.Amount)
		} else {
			bi = &blockInfo {
				Transactions: dbInfo.Transactions,
				Amount: dbInfo.Amount,
			}
			s.cache.put(blockNum, bi)
		}
	}

	rawResp, err := json.Marshal(bi)
	if err != nil {
		log.Printf("Failed to marshal block info. Error: %v \n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(rawResp)
}
