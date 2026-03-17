module downstream

go 1.16

require (
	github.com/go-courier/courier v1.5.0
	github.com/go-courier/httptransport v1.18.1
	github.com/go-courier/metax v1.2.1
	github.com/go-courier/statuserror v1.2.1
)

replace github.com/go-courier/httptransport v1.18.1 => ../../
