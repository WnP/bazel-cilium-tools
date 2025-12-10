# Cilium Tools Bazel Module

## Overview

The `cilium_tools` module provides a simple, hermetic Bazel integration for the Cilium CLI, enabling easy management of Cilium CNI (Container Network Interface) in Kubernetes clusters. This module offers:

- Platform-aware Cilium CLI binary downloads
- Simple `cilium_wait` macro for checking Cilium cluster status
- LoadBalancer IP pool generation from cluster node information
- Configurable Cilium CLI version via module extension

## Features

- 🚀 Hermetic Cilium CLI binary downloads
- 🔧 Flexible cluster status checking
- 🌐 Automatic LoadBalancer IP pool generation using Go binary
- 📦 Version-configurable module extension
- 🧪 Comprehensive test coverage including Go unit tests

## Prerequisites

- Bazel 7.0.0+
- Kubernetes cluster with Cilium installed
- Configured kubeconfig

## Installation

Add the following to your `MODULE.bazel`:

```starlark
bazel_dep(name = "cilium_tools", version = "1.0.0")

# Optional: Configure Cilium CLI version
cilium_extension = use_extension("@cilium_tools//:extensions.bzl", "cilium")
cilium_extension.version(version = "0.16.20")  # Optional, defaults to 0.16.20
use_repo(cilium_extension, "cilium_binary")
```

## Public API Macros

### `cilium_generate_loadbalancer_pool` Function

Generates CiliumLoadBalancerIPPool YAML from Kubernetes node information using a built-in Go binary:

```starlark
load("@cilium_tools//:defs.bzl", "cilium_generate_loadbalancer_pool")
load("@kubectl_tools//:defs.bzl", "kubectl_get")

# Get node IP information
kubectl_get(
    name = "get_node_ips",
    kind = "nodes",
    output = "json"
)

# Generate Cilium LoadBalancer pool
cilium_generate_loadbalancer_pool(
    name = "lb_pool",
    pool_name = "default-pool",
    nodes = ":get_node_ips",
)
```

#### Implementation Details

- Uses an embedded Go binary for robust JSON parsing
- Extracts InternalIP from the first node in the cluster
- Transforms IPs (e.g., 192.168.1.5) to a /29 subnet (192.168.255.200/29)
- Generates LoadBalancer pool YAML using Go templates
- Fully self-contained - no external dependencies required

### `cilium_wait` Macro

The `cilium_wait` macro creates a `sh_binary` target that runs `cilium status --wait` with the specified Kubernetes context.

#### Basic Usage

```starlark
load("@cilium_tools//:defs.bzl", "cilium_wait")

cilium_wait(
    name = "check_cilium_status",
    context = "my-cluster",
)
```

### Advanced Configuration

```starlark
cilium_wait(
    name = "check_cilium_status",
    context = "my-cluster-context",      # Mandatory
    kubeconfig = "/path/to/kubeconfig",  # Optional
    namespace = "kube-system",           # Optional
)
```

### Running the Target

```bash
bazel run //:check_cilium_status
```

## Module Extension: Version Configuration

Configure the Cilium CLI version in your `MODULE.bazel`:

```starlark
cilium_extension = use_extension("@cilium_tools//:extensions.bzl", "cilium")
cilium_extension.version(version = "0.16.20")
```

## Integration with Buildfarm Project

Use `cilium_wait` to prepare your Kubernetes environment:

```starlark
# In BUILD.bazel
load("@cilium_tools//:defs.bzl", "cilium_wait")

cilium_wait(
    name = "wait_for_cilium",
    kubeconfig = "$(location //:kind_kubeconfig)",
)

# Integration example with LoadBalancer pool generation
genrule(
    name = "create_loadbalancer_manifests",
    srcs = [":get_node_ips", ":lb_pool"],
    outs = ["loadbalancer_manifests.yaml"],
    cmd = "cat $(location :get_node_ips) $(location :lb_pool) > $@"
)

sh_binary(
    name = "deploy_buildfarm",
    srcs = ["deploy.sh"],
    data = [":wait_for_cilium", ":create_loadbalancer_manifests"],
)
```

### Integration Workflow

1. Use `kubectl_get` to fetch node IP information
2. Use `cilium_generate_loadbalancer_pool` to create load balancer configuration
3. Use `genrule` to combine manifests
4. Deploy Buildfarm with complete networking configuration

## Testing

Run module tests:

```bash
# Run all Starlark tests
bazel test @cilium_tools//private/tests:all_tests

# Run Go unit tests for the gen_pool binary
bazel test @cilium_tools//cmd/gen_pool:gen_pool_test
```

## Limitations

- Currently supports Linux, macOS, and Windows
- Requires pre-installed Cilium in the Kubernetes cluster
- Minimal error handling in the macro

## Contributing

1. Maintain functional programming principles
2. Add comprehensive unit tests
3. Follow Starlark compatibility guidelines

## License

[Insert appropriate license]

## Support

For issues or feature requests, please file a GitHub issue.