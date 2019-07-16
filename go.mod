module github.com/edgexfoundry/app-filter-mind

go 1.12

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/antlr/antlr4 v0.0.0-20190714172556-b627fffdd1e8 // indirect
	github.com/caibirdme/yql v0.0.0-20190420141751-229981055127
	github.com/eclipse/paho.mqtt.golang v1.2.0
	github.com/edgexfoundry/app-functions-sdk-go v0.1.1-0.20190709232221-e414d04e7129
	github.com/edgexfoundry/go-mod-core-contracts v0.1.0
	github.com/gorilla/mux v1.7.2
)

replace (
	golang.org/x/crypto v0.0.0-20181029021203-45a5f77698d3 => github.com/golang/crypto v0.0.0-20181029021203-45a5f77698d3
	golang.org/x/net v0.0.0-20181023162649-9b4f9f5ad519 => github.com/golang/net v0.0.0-20181023162649-9b4f9f5ad519
	golang.org/x/net v0.0.0-20181201002055-351d144fa1fc => github.com/golang/net v0.0.0-20181201002055-351d144fa1fc
	golang.org/x/sync v0.0.0-20181221193216-37e7f081c4d4 => github.com/golang/sync v0.0.0-20181221193216-37e7f081c4d4
	golang.org/x/sys v0.0.0-20180823144017-11551d06cbcc => github.com/golang/sys v0.0.0-20180823144017-11551d06cbcc
	golang.org/x/sys v0.0.0-20181026203630-95b1ffbd15a5 => github.com/golang/sys v0.0.0-20181026203630-95b1ffbd15a5
)
