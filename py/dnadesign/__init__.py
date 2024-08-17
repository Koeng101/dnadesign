r'''
# What is dnadesign

dnadesign is a batteries-included software suite for engineering biology.

We're building a practical, modern, and ambitious library for engineering biology. While the library itself is written in Go, specifically for its simplicity and maintainability, it is accessible using Python, the programming language that biologists would be most familiar with.

'''

# Required so that pdoc doesn't attempt to parse libdnadesign
__all__ = [
    "__init__",
    "cffi_bindings",
    "parsers"
    ]
