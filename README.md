# OC RPC Health Checker – in Go

This is just a small and reasonably self contained tool to check the healthz endpoint of a OC RPC service.

Unfortunately it depends on ZeroMQ's c-library since they are not done with their pure Go solution yet, so both an alpine and a debian version is built.

You can find the latest binaries on the release page.

### Usage

Example for datastore (port 7246): `./healthz --endpoint=tcp://localhost:7246`

Other options

```
$ ./healthz -h
Usage of ./healthz:
  -endpoint string
    	where to connect (default "tcp://localhost:7200")
  -retries uint
    	how many retries (default 3)
  -timeout uint
    	timeout in ms (default 2000)
```
