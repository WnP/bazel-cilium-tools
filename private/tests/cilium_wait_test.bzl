"""Simple tests for cilium_wait macro."""

load("//:defs.bzl", "cilium_wait")

def cilium_wait_test_suite():
    """Test suite for cilium_wait macro.

    Creates test targets that verify the macro generates correct sh_binary targets.
    These targets can be inspected with 'bazel query' to verify the macro works.
    """

    # Test basic usage - should create sh_binary with minimal args
    cilium_wait(
        name = "test_basic_wait",
        context = "test-context",
        tags = ["manual", "test"],
    )

    # Test with kubeconfig parameter
    cilium_wait(
        name = "test_with_kubeconfig",
        context = "test-context",
        kubeconfig = "test-kubeconfig.yaml",
        tags = ["manual", "test"],
    )

    # Test with context parameter
    cilium_wait(
        name = "test_with_context",
        context = "test-context",
        tags = ["manual", "test"],
    )

    # Test with namespace parameter
    cilium_wait(
        name = "test_with_namespace",
        context = "test-context",
        namespace = "test-namespace",
        tags = ["manual", "test"],
    )

    # Test with all parameters
    cilium_wait(
        name = "test_with_all_params",
        kubeconfig = "test-kubeconfig.yaml",
        context = "test-context",
        namespace = "test-namespace",
        tags = ["manual", "test"],
    )

    # Test that additional kwargs are passed through
    cilium_wait(
        name = "test_with_kwargs",
        context = "test-context",
        visibility = ["//visibility:private"],
        tags = ["manual", "test"],
    )
