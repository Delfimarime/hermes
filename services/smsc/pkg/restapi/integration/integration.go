package integration

type Integration struct {
	AzureVault
	HashicorpVault
	AwsSecretManager
	Name string
	Type string
}

type AzureVault struct {
	Url string
}

type AwsSecretManager struct {
	Region string
}

type HashicorpVault struct {
	Token string
	Path  string
}
