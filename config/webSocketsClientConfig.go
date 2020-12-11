package config

// WebSocketsClientConfig -
type WebSocketsClientConfig struct {
	WebSocketsClient WebSocketsClient
}

// WebSocketsClient -
type WebSocketsClient struct {
	Enabled bool
	Routes  []RouteConfig
}
