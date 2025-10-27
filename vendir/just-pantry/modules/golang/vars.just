## Variables for Golang development environment
# These variables are used to manage the Go development environment, including paths and module information.

go-path := `go env GOPATH`
go-bin := `go env GOBIN`
go-mod := `go list -m`
go-mod-path := `echo $(go list -m) | cut -d/ -f2-`
go-mod-short := `go list -m | awk -F/ '{print $NF}'`
