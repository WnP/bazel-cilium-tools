"""Platform-aware cilium binary download implementation."""

def _cilium_download_impl(repository_ctx):
    """Download cilium binary for the current platform."""

    # Detect platform
    os_name = repository_ctx.os.name.lower()
    if os_name.startswith("mac"):
        os_name = "darwin"
    elif os_name.startswith("windows"):
        os_name = "windows"
    else:
        os_name = "linux"

    # Detect architecture
    arch = repository_ctx.os.arch.lower()
    if arch == "x86_64" or arch == "amd64":
        arch = "amd64"
    elif arch == "aarch64" or arch == "arm64":
        arch = "arm64"
    else:
        fail("Unsupported architecture: {}".format(arch))

    version = repository_ctx.attr.version

    # cilium-cli binary URL (distributed as tar.gz)
    url = "https://github.com/cilium/cilium-cli/releases/download/v{}/cilium-{}-{}.tar.gz".format(
        version, os_name, arch
    )

    # Download and extract tar.gz
    repository_ctx.download_and_extract(
        url = url,
        stripPrefix = "",  # cilium binary is in the root of the archive
    )

    # Create BUILD file
    # Export the binary file so it can be used as a source
    build_content = """# Generated cilium binary
exports_files(["cilium"], visibility = ["//visibility:public"])

alias(
    name = "cilium_binary",
    actual = ":cilium",
    visibility = ["//visibility:public"],
)
"""

    repository_ctx.file("BUILD.bazel", build_content)

cilium_download = repository_rule(
    implementation = _cilium_download_impl,
    attrs = {
        "version": attr.string(
            doc = "cilium-cli version to download",
            mandatory = True,
        ),
    },
    doc = "Downloads cilium-cli binary for the current platform",
)
