package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kevydotvinu/onc"
)

func calculatorHandler(w http.ResponseWriter, r *http.Request) {

	type ErrorResponse struct {
		Error string `json:"error"`
	}

	client := r.RemoteAddr
	var request onc.Request
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&request)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error decoding JSON payload: %s", err), http.StatusBadRequest)
		return
	}

	fmt.Printf("Client: %s\n", client)
	fmt.Printf("Request: %v\n", request)

	if request.ClusterNetwork == "" || request.HostPrefix == 0 || request.ServiceNetwork == "" || request.MachineNetwork == "" {
		http.Error(w, "Missing required field(s)", http.StatusBadRequest)
		return
	}

	response, err := onc.CalculateNetwork(request)
	var jsonResponse []byte
	if err != nil {
		errorResponse := ErrorResponse{
			Error: fmt.Sprintf("failed calculation: %v", err),
		}
		jsonResponse, _ = json.Marshal(errorResponse)
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, string(jsonResponse), http.StatusInternalServerError)
		fmt.Println("Response:", string(jsonResponse))
		return
	}

	jsonResponse, err = json.Marshal(response)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error encoding JSON: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
	fmt.Println("Response:", string(jsonResponse))
}
