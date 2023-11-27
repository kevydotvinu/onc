package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Request struct {
	HostPrefix     int    `json:"hostPrefix"`
	ClusterNetwork string `json:"clusterNetwork"`
	ServiceNetwork string `json:"serviceNetwork"`
	MachineNetwork string `json:"machineNetwork"`
}

type Response struct {
	PodNetwork          string `json:"pod-network"`
	ServiceNetwork      string `json:"service-network"`
	MachineNetwork      string `json:"machine-network"`
	NumPods             int    `json:"number-of-pods"`
	NumServices         int    `json:"number-of-services"`
	NumNodes            int    `json:"number-of-nodes"`
	PodsPerNode         int    `json:"pods-per-node"`
	MachineNetworkNodes int    `json:"machine-network-nodes"`
	Conflict            bool   `json:"network-conflict"`
}

func calculateNetwork(request Request) (*Response, error) {
	networks := []string{request.ClusterNetwork, request.ServiceNetwork}
	for _, network := range networks {
		if !isValidCIDR(network) {
			return nil, fmt.Errorf("invalid network CIDR: %s", network)
		}
	}

	podNetwork := request.ClusterNetwork
	serviceNetwork := request.ServiceNetwork
	machineNetwork := request.MachineNetwork

	_, cNet, err := net.ParseCIDR(podNetwork)
	if err != nil {
		return nil, fmt.Errorf("error parsing CIDR: %v", err)
	}

	clusterNetworkPrefix, err := strconv.Atoi(strings.Split(cNet.String(), "/")[1])
	if err != nil {
		return nil, err
	}
	numPods := int(float64(math.Pow(2, float64(32-clusterNetworkPrefix)) - 2))
	availableHostPrefix := 32 - request.HostPrefix
	podsPerNode := int(math.Pow(2, float64(availableHostPrefix)) - 2)
	numNodes := numPods / podsPerNode

	numServices, err := countIPs(serviceNetwork)
	if err != nil {
		return nil, err
	}

	machineNetworkNodes, err := countIPs(machineNetwork)
	if err != nil {
		return nil, err
	}

	conflict, err := checkCIDRConflict(request.ClusterNetwork, request.ServiceNetwork, request.MachineNetwork)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	return &Response{
		PodNetwork:          podNetwork,
		ServiceNetwork:      serviceNetwork,
		MachineNetwork:      machineNetwork,
		NumPods:             numPods,
		NumServices:         numServices,
		NumNodes:            numNodes,
		PodsPerNode:         podsPerNode,
		MachineNetworkNodes: machineNetworkNodes,
		Conflict:            conflict,
	}, nil
}

func isValidCIDR(cidr string) bool {
	_, _, err := net.ParseCIDR(cidr)
	return err == nil
}

func countIPs(cidr string) (int, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return 0, err
	}
	networkPrefix, err := strconv.Atoi(strings.Split(ipNet.String(), "/")[1])
	if err != nil {
		return 0, err
	}
	IPs := int(float64(math.Pow(2, float64(32-networkPrefix)) - 2))
	return IPs, nil
}

func checkCIDRConflict(cidrs ...string) (bool, error) {
	var networks []*net.IPNet

	for _, cidr := range cidrs {
		_, netIP, err := net.ParseCIDR(cidr)
		if err != nil {
			return false, err
		}
		networks = append(networks, netIP)
	}

	// Check for conflicts
	for i, network1 := range networks {
		for j, network2 := range networks {
			if i != j && (network1.Contains(network2.IP) || network2.Contains(network1.IP)) {
				return true, nil
			}
		}
	}

	return false, nil
}

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

	var req Request
	if err := json.NewDecoder(strings.NewReader(request.Body)).Decode(&req); err != nil {
		return &events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf("Failed to parse payload: %v", err),
		}, nil
	}

	var results *Response
	var err error
	if results, err = calculateNetwork(req); err != nil {
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

func main() {
	lambda.Start(calculatorHandler)
}
