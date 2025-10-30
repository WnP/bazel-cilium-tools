# Cilium Tools Usage

## cilium_generate_loadbalancer_pool

Generates a CiliumLoadBalancerIPPool YAML from cluster node information.

### Usage

```starlark
load("@kubectl_tools//:defs.bzl", "kubectl_get")
load("@cilium_tools//:defs.bzl", "cilium_generate_loadbalancer_pool")

# First, get cluster nodes in JSON format
kubectl_get(
    name = "get_nodes",
    kind = "nodes",
    output = "json",
)

# Then generate the LoadBalancer IP pool
cilium_generate_loadbalancer_pool(
    name = "generate_lb_pool",
    pool_name = "default-pool",
    nodes = ":get_nodes",
    output_file = "cilium-lb-pool.yaml",
)
```

### How it works

1. Takes the output from `kubectl_get` (nodes in JSON format)
2. Uses a Go binary to parse the JSON and extract Internal IP addresses
3. Transforms the first node's IP from `x.y.z.w` to `x.y.255.200/29`
4. Generates a CiliumLoadBalancerIPPool YAML with that CIDR range using Go templates

**Note**: The implementation uses a Go binary (`gen_pool`) for robust JSON parsing and YAML generation, ensuring reliable handling of various Kubernetes node output formats.

### Example output

```yaml
apiVersion: cilium.io/v2alpha1
kind: CiliumLoadBalancerIPPool
metadata:
  name: default-pool
spec:
  blocks:
  - cidr: 172.18.255.200/29
```

## kubectl_get (Updated)

The `kubectl_get` function has been converted from `sh_binary` to `genrule` to support piping output to files.

### Usage

```starlark
load("@kubectl_tools//:defs.bzl", "kubectl_get")

# Get nodes and write to file
kubectl_get(
    name = "get_nodes",
    kind = "nodes",
    output = "json",
    output_file = "nodes.json",  # Optional, defaults to {name}.out
)

# Get pods in specific namespace
kubectl_get(
    name = "get_pods",
    kind = "pods",
    namespace = "kube-system",
    output = "wide",
)
```

The output file can then be used as input to other build rules or processed further.

## Integration Example

See the root `/workspace/BUILD.bazel` for a complete example of how to use both functions together to generate Cilium LoadBalancer IP pools from cluster node information.