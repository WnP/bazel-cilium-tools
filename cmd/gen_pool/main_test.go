package main

import (
	"bytes"
	"strings"
	"testing"
	"text/template"
)

func TestExtractInternalIP(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name: "single_node_json",
			input: `{
				"status": {
					"addresses": [
						{"address": "192.168.1.100", "type": "InternalIP"},
						{"address": "test-node", "type": "Hostname"}
					]
				}
			}`,
			want:    "192.168.1.100",
			wantErr: false,
		},
		{
			name: "node_list_json",
			input: `{
				"items": [
					{
						"status": {
							"addresses": [
								{"address": "172.18.0.3", "type": "InternalIP"}
							]
						}
					},
					{
						"status": {
							"addresses": [
								{"address": "172.18.0.2", "type": "InternalIP"}
							]
						}
					}
				]
			}`,
			want:    "172.18.0.3",
			wantErr: false,
		},
		{
			name: "no_internal_ip",
			input: `{
				"status": {
					"addresses": [
						{"address": "test-node", "type": "Hostname"},
						{"address": "203.0.113.1", "type": "ExternalIP"}
					]
				}
			}`,
			want:    "",
			wantErr: true,
		},
		{
			name: "empty_json",
			input: `{}`,
			want:    "",
			wantErr: true,
		},
		{
			name: "invalid_json",
			input: `{invalid json content}`,
			want:    "",
			wantErr: true,
		},
		{
			name: "empty_node_list",
			input: `{"items": []}`,
			want:    "",
			wantErr: true,
		},
		{
			name: "node_with_multiple_ips_returns_first",
			input: `{
				"items": [
					{
						"status": {
							"addresses": [
								{"address": "10.0.0.1", "type": "InternalIP"},
								{"address": "10.0.0.2", "type": "InternalIP"}
							]
						}
					}
				]
			}`,
			want:    "10.0.0.1",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractInternalIP([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("extractInternalIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("extractInternalIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComputePoolCIDR(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		want    string
		wantErr bool
	}{
		{
			name:    "standard_ip",
			ip:      "192.168.1.100",
			want:    "192.168.255.200/29",
			wantErr: false,
		},
		{
			name:    "10_network",
			ip:      "10.244.0.1",
			want:    "10.244.255.200/29",
			wantErr: false,
		},
		{
			name:    "172_network",
			ip:      "172.18.0.3",
			want:    "172.18.255.200/29",
			wantErr: false,
		},
		{
			name:    "single_digit_octets",
			ip:      "1.2.3.4",
			want:    "1.2.255.200/29",
			wantErr: false,
		},
		{
			name:    "invalid_ip_format",
			ip:      "192.168.1",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid_ip_too_many_octets",
			ip:      "192.168.1.1.1",
			want:    "",
			wantErr: true,
		},
		{
			name:    "not_an_ip",
			ip:      "not-an-ip",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty_string",
			ip:      "",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := computePoolCIDR(tt.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("computePoolCIDR() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("computePoolCIDR() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTemplateGeneration(t *testing.T) {
	tests := []struct {
		name     string
		poolName string
		poolCIDR string
		want     string
	}{
		{
			name:     "standard_pool",
			poolName: "test-pool",
			poolCIDR: "192.168.255.200/29",
			want: `apiVersion: cilium.io/v2alpha1
kind: CiliumLoadBalancerIPPool
metadata:
  name: test-pool
spec:
  blocks:
  - cidr: 192.168.255.200/29
`,
		},
		{
			name:     "pool_with_special_chars",
			poolName: "my-special-pool-123",
			poolCIDR: "10.0.255.200/29",
			want: `apiVersion: cilium.io/v2alpha1
kind: CiliumLoadBalancerIPPool
metadata:
  name: my-special-pool-123
spec:
  blocks:
  - cidr: 10.0.255.200/29
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := PoolConfig{
				PoolName: tt.poolName,
				PoolCIDR: tt.poolCIDR,
			}

			tmpl, err := template.New("pool").Parse(poolTemplate)
			if err != nil {
				t.Fatalf("Failed to parse template: %v", err)
			}

			var buf bytes.Buffer
			err = tmpl.Execute(&buf, config)
			if err != nil {
				t.Fatalf("Failed to execute template: %v", err)
			}

			got := buf.String()
			if got != tt.want {
				t.Errorf("Template output mismatch\nGot:\n%s\nWant:\n%s", got, tt.want)
			}
		})
	}
}

func TestEndToEnd(t *testing.T) {
	// Test the complete flow from JSON input to YAML output
	input := `{
		"status": {
			"addresses": [
				{"address": "192.168.1.100", "type": "InternalIP"},
				{"address": "test-node", "type": "Hostname"}
			]
		}
	}`

	// Extract IP
	ip, err := extractInternalIP([]byte(input))
	if err != nil {
		t.Fatalf("Failed to extract IP: %v", err)
	}

	// Compute CIDR
	cidr, err := computePoolCIDR(ip)
	if err != nil {
		t.Fatalf("Failed to compute CIDR: %v", err)
	}

	// Generate YAML
	config := PoolConfig{
		PoolName: "e2e-test-pool",
		PoolCIDR: cidr,
	}

	tmpl, err := template.New("pool").Parse(poolTemplate)
	if err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, config)
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	output := buf.String()

	// Verify output contains expected values
	expectedStrings := []string{
		"apiVersion: cilium.io/v2alpha1",
		"kind: CiliumLoadBalancerIPPool",
		"name: e2e-test-pool",
		"cidr: 192.168.255.200/29",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Output missing expected string: %s\nFull output:\n%s", expected, output)
		}
	}
}