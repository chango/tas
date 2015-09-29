package tas

type TASConfig struct {
	ZMQPort     string // Port to listen for ZMQ traffic (from agents)
	ZMQAddress  string // Address to listen for ZMQ traffic (from agents)
	HTTPPort    string // HTTP Port to listen on for querying/stats
	HTTPAddress string // HTTP Address to listen on for querying/stats
}

// Returns a default TAS server configuration that uses the default ports
func NewDefaultTASConfig() (c *TASConfig) {
	c = &TASConfig{
		ZMQPort:     "7450",
		ZMQAddress:  "*",
		HTTPPort:    "7451",
		HTTPAddress: "0.0.0.0",
	}
	return
}
