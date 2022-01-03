package gridon

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
)

// NewWebService - 新しいWebサービスの取得
func NewWebService(port string, strategyStore IStrategyStore, kabusAPI IKabusAPI) IWebService {
	return &webService{
		port:          port,
		strategyStore: strategyStore,
		kabusAPI:      kabusAPI,
		routes:        map[string]map[string]http.Handler{},
	}
}

// IWebService - Webサービスのインターフェース
type IWebService interface {
	StartWebServer() error
}

// webService - Webサービス
type webService struct {
	port          string
	strategyStore IStrategyStore
	kabusAPI      IKabusAPI
	routes        map[string]map[string]http.Handler
}

// StartWebServer - Webサービスの開始
func (s *webService) StartWebServer() error {
	ln, err := net.Listen("tcp", s.port)
	if err != nil {
		return err
	}

	s.routes = map[string]map[string]http.Handler{
		"/api/strategy": {
			"GET":    http.HandlerFunc(s.getStrategy),
			"DELETE": http.HandlerFunc(s.deleteStrategy),
		},
		"/api/strategies": {
			"GET":  http.HandlerFunc(s.getStrategies),
			"POST": http.HandlerFunc(s.postSaveStrategy),
		},
	}

	return http.Serve(ln, s)
}

// ServeHTTP - WebServerのルーティング
func (s *webService) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	if _, ok := s.routes[path]; !ok {
		http.Error(w, "404 Not Found", http.StatusNotFound)
		return
	}

	method := req.Method
	if _, ok := s.routes[path][method]; !ok {
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// 先にヘッダを付けておく
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	s.routes[path][method].ServeHTTP(w, req)
}

// getStrategy - 戦略の取得
func (s *webService) getStrategy(w http.ResponseWriter, req *http.Request) {
	code := req.FormValue("code")
	strategy, err := s.strategyStore.GetByCode(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_ = json.NewEncoder(w).Encode(strategy)
}

// deleteStrategy - 戦略の削除
func (s *webService) deleteStrategy(w http.ResponseWriter, req *http.Request) {
	code := req.FormValue("code")
	strategy, err := s.strategyStore.GetByCode(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.strategyStore.DeleteByCode(code); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(strategy)
}

// getStrategies - 戦略一覧の取得
func (s *webService) getStrategies(w http.ResponseWriter, _ *http.Request) {
	strategies, err := s.strategyStore.GetStrategies()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(strategies)
}

// postSaveStrategy - 戦略の保存
func (s *webService) postSaveStrategy(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	strategy := &Strategy{}
	if err := json.NewDecoder(req.Body).Decode(strategy); err != nil && err != io.EOF {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if strategy.Code == "" {
		http.Error(w, "code is required", http.StatusBadRequest)
		return
	}

	// TICKグループのセット
	symbol, err := s.kabusAPI.GetSymbol(strategy.SymbolCode, strategy.Exchange)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	strategy.TickGroup = symbol.TickGroup

	if err := s.strategyStore.Save(strategy); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(strategy)
}
