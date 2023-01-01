package config

// SASLGSSAPIConfig represents the Kafka Kerberos config
type SASLGSSAPIConfig struct {
	AuthType           string `yaml:"authType"`
	KeyTabPath         string `yaml:"keyTabPath"`
	KerberosConfigPath string `yaml:"kerberosConfigPath"`
	ServiceName        string `yaml:"serviceName"`
	Username           string `yaml:"username"`
	Password           string `yaml:"password"`
	Realm              string `yaml:"realm"`
}
