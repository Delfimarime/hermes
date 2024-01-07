package config

type Configuration struct {
	Smsc          Smsc
	Logger        Logger
	OpenTelemetry OpenTelemetry
}

type Smsc struct {
	StartupTimeout int64
}

type OpenTelemetry struct {
	ServiceName string
	Spans       OpenTelemetrySpanExporter
	Metrics     OpenTelemetryMetricExporter
}

type OpenTelemetryExporter struct {
	TLS             *TLS
	Timeout         int64
	ReconnectPeriod int64
	Endpoint        string
	ExportMethod    string
}

type OpenTelemetrySpanExporter struct {
	MaxQueueSize int
	OpenTelemetryExporter
}

type OpenTelemetryMetricExporter struct {
	OpenTelemetryExporter
	Reader *OpenTelemetryMetricExporterReader
}

type OpenTelemetryMetricExporterReader struct {
	Period int64
}

type TLS struct {
	File string
}

type Logger struct {
	Level string
}
