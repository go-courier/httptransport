module downstream

go 1.26

require (
	github.com/go-courier/courier v1.5.0
	github.com/go-courier/httptransport v1.18.1
	github.com/go-courier/metax v1.3.0
	github.com/go-courier/statuserror v1.2.1
)

require (
	github.com/go-courier/enumeration v1.3.1 // indirect
	github.com/go-courier/x v0.1.2 // indirect
	github.com/julienschmidt/httprouter v1.3.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/mod v0.10.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
	golang.org/x/tools v0.9.1 // indirect
)

replace github.com/go-courier/httptransport v1.18.1 => ../../
