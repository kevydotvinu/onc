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

var version string

type Request struct {
	HostPrefix     int    `json:"hostPrefix"`
	ClusterNetwork string `json:"clusterNetwork"`
	ServiceNetwork string `json:"serviceNetwork"`
	MachineNetwork string `json:"machineNetwork"`
}

type Response struct {
	PodNetwork          string      `json:"pod-network"`
	ServiceNetwork      string      `json:"service-network"`
	MachineNetwork      string      `json:"machine-network"`
	NumPods             int         `json:"number-of-pods"`
	NumServices         int         `json:"number-of-services"`
	NumNodes            int         `json:"number-of-nodes"`
	PodsPerNode         PodsPerNode `json:"pods-per-node"`
	MachineNetworkNodes int         `json:"machine-network-nodes"`
	Conflict            bool        `json:"network-conflict"`
}

type PodsPerNode struct {
	Sdn int `json:"sdn"`
	Ovn int `json:"ovn"`
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
	hostPrefix := request.HostPrefix

	numPods, err := countIPs(podNetwork)
	if err != nil {
		return nil, err
	}
	numNodes := len(splitSubnet(podNetwork, hostPrefix))
	podsPerNode := numPods / numNodes
	cniPodsPerNode := PodsPerNode{
		Sdn: podsPerNode - 2,
		Ovn: podsPerNode - 3,
	}

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
		fmt.Println("error:", err)
		return nil, err
	}

	return &Response{
		PodNetwork:          podNetwork,
		ServiceNetwork:      serviceNetwork,
		MachineNetwork:      machineNetwork,
		NumPods:             numPods,
		NumServices:         numServices,
		NumNodes:            numNodes,
		PodsPerNode:         cniPodsPerNode,
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

func splitSubnet(subnet string, prefixLength int) []*net.IPNet {
	var subnets []*net.IPNet

	_, snet, err := net.ParseCIDR(subnet)
	if err != nil {
		fmt.Println("error parsing subnet:", err)
		return subnets
	}
	ip := snet.IP
	ones, _ := snet.Mask.Size()
	subnetCount := 2 << uint(prefixLength-ones-1)
	stepSize := 1 << uint(32-ones)

	for i := 0; i < subnetCount; i++ {
		newIP := make(net.IP, len(ip))
		copy(newIP, ip)
		for j := 3; j >= 0; j-- {
			newIP[j] += byte((i * stepSize) >> uint((3-j)*8))
		}

		newSubnet := &net.IPNet{
			IP:   newIP,
			Mask: net.CIDRMask(prefixLength, 32),
		}

		subnets = append(subnets, newSubnet)
	}

	return subnets
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
	fmt.Printf("Starting onc, version %s\n", version)
	lambda.Start(calculatorHandler)
}
