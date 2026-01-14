package config

// Config holds all configuration for Omnipoll
type Config struct {
	SQLServer SQLServerConfig `json:"sqlServer" yaml:"sqlServer"`
	MQTT      MQTTConfig      `json:"mqtt" yaml:"mqtt"`
	MongoDB   MongoDBConfig   `json:"mongodb" yaml:"mongodb"`
	Polling   PollingConfig   `json:"polling" yaml:"polling"`
	Admin     AdminConfig     `json:"admin" yaml:"admin"`
}

type SQLServerConfig struct {
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	Database string `json:"database" yaml:"database"`
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"` // Encrypted at rest
}

type MQTTConfig struct {
	Broker      string `json:"broker" yaml:"broker"`
	Port        int    `json:"port" yaml:"port"`
	Topic       string `json:"topic" yaml:"topic"`
	TopicPrefix string `json:"topicPrefix" yaml:"topicPrefix"` // e.g., "feeding/mowi"
	ClientID    string `json:"clientId" yaml:"clientId"`
	User        string `json:"user" yaml:"user"`
	Password    string `json:"password" yaml:"password"` // Encrypted at rest
	QoS         byte   `json:"qos" yaml:"qos"`
	UseTLS      bool   `json:"useTLS" yaml:"useTLS"`
}

type MongoDBConfig struct {
	URI        string `json:"uri" yaml:"uri"`
	Database   string `json:"database" yaml:"database"`
	Collection string `json:"collection" yaml:"collection"`
}

type PollingConfig struct {
	IntervalMS int `json:"intervalMs" yaml:"intervalMs"`
	BatchSize  int `json:"batchSize" yaml:"batchSize"`
}

type AdminConfig struct {
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"` // Encrypted at rest
}
