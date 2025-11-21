/*
Package external contains functions that interact with outside programs.

The primary way that external interacts with common bioinformatics tools is
through the command line, so the bioinformatics programs must be installed on
your local computer.

We would like to port these programs to be in `lib`, using WASM for reliable
builds, but until then, they live here.
*/
package external
