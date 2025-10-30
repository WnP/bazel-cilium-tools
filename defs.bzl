"""Public API for cilium_tools module.

Provides simple macros for cilium operations via sh_binary.
"""

load("//private:loadbalancer_pool.bzl", "loadbalancer_pool")

def cilium_wait(name, kubeconfig = None, context = None, namespace = None, **kwargs):
    """Create a sh_binary target that waits for Cilium to be ready.

    Args:
        name: Name of the target
        kubeconfig: Path to kubeconfig file (optional, uses $KUBECONFIG env var if not specified)
        context: kubectl context to use (optional)
        namespace: Kubernetes namespace (optional)
        **kwargs: Additional arguments passed to sh_binary
    
    Note: To use the KUBECONFIG environment variable, run with:
        bazel run --action_env=KUBECONFIG //:target_name
    """
    # Build cilium status command arguments
    args = ["status", "--wait"]

    if kubeconfig:
        args.extend(["--kubeconfig", kubeconfig])

    if context:
        args.extend(["--context", context])

    if namespace:
        args.extend(["--namespace", namespace])

    # Create sh_binary target
    native.sh_binary(
        name = name,
        srcs = ["@cilium_binary//:cilium"],
        args = args,
        data = [kubeconfig] if kubeconfig else [],
        **kwargs
    )

def cilium_generate_loadbalancer_pool(name, pool_name, nodes, **kwargs):
    """Generate a CiliumLoadBalancerIPPool YAML from node information.

    This macro uses a custom Bazel rule to process kubectl get nodes output
    and generate a CiliumLoadBalancerIPPool resource with IPs derived from the
    nodes' Internal IPs, transforming x.y.z.w -> x.y.255.200/29.

    The implementation uses a Go binary for robust JSON parsing and YAML
    generation, ensuring reliable handling of various Kubernetes node formats.

    Args:
        name: Name of the target
        pool_name: Name for the CiliumLoadBalancerIPPool resource
        nodes: Input file containing kubectl get nodes output (typically JSON format)
        **kwargs: Additional arguments passed to the underlying rule

    Example:
        kubectl_get(
            name = "get_nodes",
            kind = "nodes",
            output = "json",
        )
        
        cilium_generate_loadbalancer_pool(
            name = "my_pool",
            pool_name = "production-pool",
            nodes = ":get_nodes",
        )
    """
    # Use the custom rule for clean implementation
    loadbalancer_pool(
        name = name,
        pool_name = pool_name,
        nodes = nodes,
        **kwargs
    )

