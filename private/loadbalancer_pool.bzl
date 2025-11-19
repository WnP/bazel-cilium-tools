"""Custom Bazel rule for generating CiliumLoadBalancerIPPool YAML files.

This rule processes kubectl node output and generates LoadBalancer pool configuration
using a Go binary for robust JSON parsing and YAML generation.
"""

def _loadbalancer_pool_impl(ctx):
    """Implementation of the loadbalancer_pool rule.

    Reads kubectl nodes output, extracts Internal IP, computes CIDR,
    and generates CiliumLoadBalancerIPPool YAML using a Go binary.
    """
    # Use the Go binary to process the nodes file and generate YAML
    args = [
        "-nodes", ctx.file.nodes.path,
        "-pool-name", ctx.attr.pool_name,
        "-output", ctx.outputs.yaml.path,
        "-pool-offset", str(ctx.attr.pool_offset),
        "-pool-mask", str(ctx.attr.pool_mask),
    ]

    ctx.actions.run(
        inputs = [ctx.file.nodes],
        outputs = [ctx.outputs.yaml],
        executable = ctx.executable._gen_pool_binary,
        arguments = args,
        mnemonic = "GenerateLoadBalancerPool",
        progress_message = "Generating CiliumLoadBalancerIPPool for %s" % ctx.label,
    )

    return [DefaultInfo(files = depset([ctx.outputs.yaml]))]

loadbalancer_pool = rule(
    implementation = _loadbalancer_pool_impl,
    attrs = {
        "nodes": attr.label(
            mandatory = True,
            allow_single_file = True,
            doc = "Input file containing kubectl get nodes output",
        ),
        "pool_name": attr.string(
            mandatory = True,
            doc = "Name for the CiliumLoadBalancerIPPool resource",
        ),
        "pool_offset": attr.int(
            default = 200,
            doc = "Third octet of the pool IP address (default: 200)",
        ),
        "pool_mask": attr.int(
            default = 29,
            doc = "CIDR mask for the pool (default: 29)",
        ),
        "_gen_pool_binary": attr.label(
            default = "//cmd/gen_pool",
            executable = True,
            cfg = "exec",
            doc = "Go binary for generating pool YAML",
        ),
    },
    outputs = {"yaml": "%{name}.yaml"},
    doc = """Generate a CiliumLoadBalancerIPPool YAML file from kubectl nodes output.

    This rule:
    1. Reads kubectl get nodes output (JSON format only)
    2. Uses a Go binary to extract the Internal IP address
    3. Computes a pool CIDR by transforming x.y.z.w -> x.y.255.<pool_offset>/<pool_mask>
    4. Generates a CiliumLoadBalancerIPPool YAML using Go templates

    Example:
        loadbalancer_pool(
            name = "my_pool",
            nodes = ":get_nodes_output",
            pool_name = "production-pool",
            pool_offset = 200,
            pool_mask = 29,
        )
    """,
)