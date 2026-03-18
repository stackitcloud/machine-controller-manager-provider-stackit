package spi

// SessionProviderInterface provides an interface to deal with cloud provider session
// Example interfaces are listed below.
type SessionProviderInterface any

// PluginSPIImpl is the real implementation of SPI interface that makes the calls to the provider SDK.
type PluginSPIImpl struct{}
