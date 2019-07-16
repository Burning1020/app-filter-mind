package rest

import (
	"encoding/json"
	"fmt"
	"github.com/edgexfoundry/app-filter-mind/internal/filter"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"net/http"
)

type WebServer struct {
	router        	*mux.Router
	LoggingClient 	logger.LoggingClient
}

func InitAndStart(port string, errChannel chan error, lc logger.LoggingClient) {
	webserver := WebServer{
		router : &mux.Router{},
		LoggingClient: lc,
	}

	// Rules
	webserver.router.HandleFunc("/api/v1/rule", webserver.ruleHandler).Methods(http.MethodPut, http.MethodGet)

	webserver.LoggingClient.Info(fmt.Sprintf("Starting HTTP Server on RulePort :%s", port))
	webserver.StartHTTPServer(port, errChannel)

}

func (webserver *WebServer) ruleHandler(writer http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		writer.Header().Add("Content-Type", "application/json")

		enc := json.NewEncoder(writer)
		r := filter.ReturnRule()
		err := enc.Encode(r)
		// Problems encoding
		if err != nil {
			webserver.LoggingClient.Error("Error encoding the data: " + err.Error())
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodPut:
		// update rule
		update := filter.Rule{}
		jsonbody, _ := ioutil.ReadAll(req.Body)
		// unmarshal json to rule
		err := json.Unmarshal(jsonbody, &update)
		if err != nil {
			webserver.LoggingClient.Error(fmt.Sprintf("Failed read body. Error: %s", err.Error()))
			writer.WriteHeader(http.StatusBadRequest)
			io.WriteString(writer, err.Error())
			return
		}
		if err = filter.RefreshRule(update); err != nil {
			webserver.LoggingClient.Error(fmt.Sprintf("Failed write Rule to file. Error: %s", err.Error()))
			writer.WriteHeader(http.StatusBadRequest)
			io.WriteString(writer, err.Error())
			return
		}
		writer.WriteHeader(http.StatusOK)
		writer.Header().Set("Content-Type", "text/plain")
		writer.Write([]byte("true"))
	}
}
// StartHTTPServer starts the http server
func (webserver *WebServer) StartHTTPServer(port string, errChannel chan error) {
	go func() {
		p := fmt.Sprintf(":%s", port)
		errChannel <- http.ListenAndServe(p, webserver.router)
	}()
}