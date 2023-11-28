package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kevydotvinu/onc"

	"github.com/aws/aws-lambda-go/events"
)

func calculatorHandler(request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {

	// Log request
	fmt.Printf("Incoming Request: %+v\n", request)

	if request.HTTPMethod != "POST" {
		origin := request.Headers["origin"]
		if strings.Contains(origin, "localhost") {
			return &events.APIGatewayProxyResponse{
				StatusCode: 200,
				Headers: map[string]string{
					"Access-Control-Allow-Origin":  origin,
					"Access-Control-Allow-Headers": "*",
				},
			}, nil
		} else {
			return &events.APIGatewayProxyResponse{
				StatusCode: 404,
			}, nil
		}
	}

	var req onc.Request
	if err := json.NewDecoder(strings.NewReader(request.Body)).Decode(&req); err != nil {
		return &events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf("Failed to parse payload: %v", err),
		}, nil
	}

	var results *onc.Response
	var err error
	if results, err = onc.CalculateNetwork(req); err != nil {
		return &events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf("Failed to make DNS request: %v", err),
		}, nil
	}

	output, err := json.Marshal(results)
	if err != nil {
		// Log error
		fmt.Printf("Error marshaling JSON response: %v\n", err)
		return nil, err
	}

	// Log response
	fmt.Printf("Outgoing Response: %+v\n", string(output))

	return &events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":                 "application/json",
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Headers": "*",
		},
		Body:            string(output),
		IsBase64Encoded: false,
	}, nil
}
