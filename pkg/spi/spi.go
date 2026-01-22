package spi

// SessionProviderInterface provides an interface to deal with cloud provider session
// Example interfaces are listed below.
type SessionProviderInterface interface {
	// NewSession(*corev1.Secret, string) (*session.Session, error)
	// NewEC2API(*session.Session) ec2iface.EC2API
}

// PluginSPIImpl is the real implementation of SPI interface that makes the calls to the provider SDK.
type PluginSPIImpl struct{}
