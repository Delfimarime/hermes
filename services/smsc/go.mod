module github.com/delfimarime/hermes/services/smsc

replace github.com/delfimarime/hermes/services/smsc => ./

go 1.21

require (
	github.com/fiorix/go-smpp v0.0.0-20210403173735-2894b96e70ba
	github.com/google/uuid v1.5.0
)

require golang.org/x/text v0.3.6 // indirect
