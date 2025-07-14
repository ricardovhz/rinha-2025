module github.com/ricardovhz/rinha-2025

go 1.24.0

require (
	github.com/gocql/gocql v0.0.0-00010101000000-000000000000
	github.com/gofiber/fiber/v3 v3.0.0-beta.4
	github.com/mattn/go-sqlite3 v1.14.28
	github.com/scylladb/scylla-go-driver v0.0.0-20221013134735-ce81923df69a
	github.com/stretchr/testify v1.10.0
)

require (
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fxamacker/cbor/v2 v2.7.0 // indirect
	github.com/gofiber/schema v1.2.0 // indirect
	github.com/gofiber/utils/v2 v2.0.0-beta.7 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/philhofer/fwd v1.1.3-0.20240916144458-20a13a1f6b7c // indirect
	github.com/pierrec/lz4/v4 v4.1.14 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/tinylib/msgp v1.2.5 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.63.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/gocql/gocql => github.com/scylladb/gocql v1.15.1
