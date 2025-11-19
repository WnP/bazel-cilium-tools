package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
)

type NodeAddress struct {
	Type    string `json:"type"`
	Address string `json:"address"`
}

type NodeStatus struct {
	Addresses []NodeAddress `json:"addresses"`
}

type Node struct {
	Status NodeStatus `json:"status"`
}

type NodeList struct {
	Items []Node `json:"items"`
}

type SingleNode struct {
	Status NodeStatus `json:"status"`
}

type PoolConfig struct {
	PoolName string
	PoolCIDR string
}

const poolTemplate = `apiVersion: cilium.io/v2alpha1
kind: CiliumLoadBalancerIPPool
metadata:
  name: {{.PoolName}}
spec:
  blocks:
  - cidr: {{.PoolCIDR}}
`

func extractInternalIP(data []byte) (string, error) {
	// Try parsing as NodeList first
	var nodeList NodeList
	if err := json.Unmarshal(data, &nodeList); err == nil && len(nodeList.Items) > 0 {
		// Get first node's internal IP
		for _, node := range nodeList.Items {
			for _, addr := range node.Status.Addresses {
				if addr.Type == "InternalIP" {
					return addr.Address, nil
				}
			}
		}
	}

	// Try parsing as single Node
	var singleNode SingleNode
	if err := json.Unmarshal(data, &singleNode); err == nil {
		for _, addr := range singleNode.Status.Addresses {
			if addr.Type == "InternalIP" {
				return addr.Address, nil
			}
		}
	}

	return "", fmt.Errorf("could not find InternalIP in nodes data")
}

func computePoolCIDR(ip string, poolOffset int, poolMask int) (string, error) {
	// Transform x.y.z.w -> x.y.255.<poolOffset>/<poolMask>
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return "", fmt.Errorf("invalid IP address format: %s", ip)
	}

	return fmt.Sprintf("%s.%s.255.%d/%d", parts[0], parts[1], poolOffset, poolMask), nil
}

func main() {
	var nodesFile string
	var poolName string
	var outputFile string
	var poolOffset int
	var poolMask int

	flag.StringVar(&nodesFile, "nodes", "", "Path to JSON file containing nodes data")
	flag.StringVar(&poolName, "pool-name", "", "Name for the CiliumLoadBalancerIPPool resource")
	flag.StringVar(&outputFile, "output", "-", "Output file path (- for stdout)")
	flag.IntVar(&poolOffset, "pool-offset", 200, "Third octet of the pool IP (default: 200)")
	flag.IntVar(&poolMask, "pool-mask", 29, "CIDR mask for the pool (default: 29)")
	flag.Parse()

	if nodesFile == "" || poolName == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s -nodes <file> -pool-name <name> [-output <file>] [-pool-offset <offset>] [-pool-mask <mask>]\n", os.Args[0])
		os.Exit(1)
	}

	// Read nodes file
	data, err := os.ReadFile(nodesFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading nodes file: %v\n", err)
		os.Exit(1)
	}

	// Extract internal IP
	internalIP, err := extractInternalIP(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error extracting internal IP: %v\n", err)
		os.Exit(1)
	}

	// Compute CIDR
	poolCIDR, err := computePoolCIDR(internalIP, poolOffset, poolMask)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error computing pool CIDR: %v\n", err)
		os.Exit(1)
	}

	// Parse and execute template
	tmpl, err := template.New("pool").Parse(poolTemplate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing template: %v\n", err)
		os.Exit(1)
	}

	config := PoolConfig{
		PoolName: poolName,
		PoolCIDR: poolCIDR,
	}

	// Determine output writer
	var writer io.Writer
	if outputFile == "-" {
		writer = os.Stdout
	} else {
		file, err := os.Create(outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		writer = file
	}

	// Execute template
	if err := tmpl.Execute(writer, config); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing template: %v\n", err)
		os.Exit(1)
	}
}