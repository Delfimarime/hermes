package model

const (
	ReceiverType    = "RECEIVER"
	TransmitterType = "TRANSMITTER"
	TransceiverType = "TRANSCEIVER"
)

type Smpp struct {
	Id          string
	Name        string
	Description string
	PoweredBy   string
	Contact     []Person
	Type        string
	Settings    Settings
	Alias       string
}

type Settings struct {
	SourceAddr  string
	ServiceType string
	Host        Host
	Bind        *Bind
	Merge       *Merge
	Enquire     *Enquire
	Response    *Response
	Delivery    *Delivery
}

type Delivery struct {
	AwaitReport bool
}

type Host struct {
	Address  string
	Username string
	Password string
}

type Bind struct {
	Timeout int64
}

type Enquire struct {
	Link        int64
	LinkTimeout int64
}

type Response struct {
	Timeout int64
}

type Merge struct {
	Interval        int64
	CleanupInterval int64
}
