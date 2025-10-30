"""Module extension for configuring cilium binary download."""

load("//private:cilium_download.bzl", "cilium_download")

def _cilium_extension_impl(module_ctx):
    """Implementation of cilium extension."""

    # Default cilium version
    cilium_version = "0.16.20"

    # Process version configuration from modules
    for mod in module_ctx.modules:
        for tag in mod.tags.version:
            cilium_version = tag.version

    # Download cilium binary
    cilium_download(
        name = "cilium_binary",
        version = cilium_version,
    )

# Tag for configuring cilium version
_version_tag = tag_class(
    attrs = {
        "version": attr.string(
            doc = "cilium-cli version to download",
            default = "0.16.20",
        ),
    },
)

# Module extension definition
cilium = module_extension(
    implementation = _cilium_extension_impl,
    tag_classes = {
        "version": _version_tag,
    },
)
