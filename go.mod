module github.com/ONSdigital/log.go/v2

go 1.24

require (
	github.com/ONSdigital/dp-net/v3 v3.2.0
	github.com/hokaccha/go-prettyjson v0.0.0-20211117102719-0474bc63780f
	github.com/smartystreets/goconvey v1.8.1
	go.opentelemetry.io/otel v1.35.0
	go.opentelemetry.io/otel/trace v1.35.0
)

require (
	github.com/ONSdigital/dp-api-clients-go/v2 v2.266.0 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gopherjs/gopherjs v1.17.2 // indirect
	github.com/jtolds/gls v4.20.0+incompatible // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/smarty/assertions v1.16.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel/metric v1.35.0 // indirect
	golang.org/x/crypto v0.40.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
)

// necessary to mitigate CVE-2022-27191, CVE-2024-45337
tool golang.org/x/crypto/ssh
