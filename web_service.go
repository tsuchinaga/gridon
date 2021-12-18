package gridon

import (
	"encoding/json"
	"net"
	"net/http"
)

// NewWebService - 新しいWebサービスの取得
func NewWebService(port string, strategyStore IStrategyStore) IWebService {
	return &webService{
		port:          port,
		strategyStore: strategyStore,
	}
}

// IWebService - Webサービスのインターフェース
type IWebService interface {
	StartWebServer() error
}

// webService - Webサービス
type webService struct {
	port          string
	ln            net.Listener
	strategyStore IStrategyStore
}

// StartWebServer - Webサービスの開始
func (s *webService) StartWebServer() error {
	ln, err := net.Listen("tcp", s.port)
	if err != nil {
		return err
	}
	s.ln = ln

	mux := http.NewServeMux()
	mux.HandleFunc("/api/strategies", s.getStrategies)

	return http.Serve(s.ln, mux)
}

func (s *webService) setHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
}

// getStrategies - 戦略一覧の取得
func (s *webService) getStrategies(w http.ResponseWriter, _ *http.Request) {
	strategies, err := s.strategyStore.GetStrategies()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if err := json.NewEncoder(w).Encode(strategies); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	s.setHeader(w)
}
