package integration

type AwsSecretManager struct {
	Key     string `json:"key"`
	Region  string `json:"region"`
	Profile string `json:"profile"`
}
