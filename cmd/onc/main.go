package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/cors"
)

func main() {
	// http.HandleFunc("/", calculatorHandler)
	corsHandler := cors.Default().Handler(http.HandlerFunc(calculatorHandler))

	// Setup graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nShutting down server...")
		os.Exit(0)
	}()

	serverAddr := "0.0.0.0"
	serverPort := "8080"
	server := fmt.Sprintf("%s:%s", serverAddr, serverPort)
	fmt.Printf("Server is running at http://%s\n", server)

	// err := http.ListenAndServe(server, nil)
	err := http.ListenAndServe(server, corsHandler)
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}
}
