# OC RPC Health Checker – in Go

This is just a small and reasonably self contained tool to check the healthz endpoint of a OC RPC service.

Unfortunately it depends on ZeroMQ's c-library since they are not done with their pure Go solution yet, so both an alpine and a debian version is built.

You can find the latest binaries on the release page.
