package pocket

import (
	"encoding/json"
	"net/http"

	"github.com/golang/glog"
)

func errorResponse(w http.ResponseWriter, code int, message string) {
	glog.Errorf("%v", message)
	jsonResponse(w, code, map[string]string{"error": message})
}

func jsonResponse(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		glog.Fatalf("Error marshalling response %v: %v", payload, err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
