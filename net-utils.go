package service

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
)

type TLSConfig struct {
	Cert           string   `json:"cert" yaml:"cert"`                         // [Optional] SSL Certificate content or file path
	Key            string   `json:"key" yaml:"key"`                           // [Optional] SSL Key content or file path
	CAs            []string `json:"ca" yaml:"ca"`                             // [Optional] CA Certificates content or file path for verifying client certificates, only used when SSL is enabled
	ClientAuthType string   `json:"client_auth_type" yaml:"client_auth_type"` // [Optional] Client authentication type, can be "none", "optional", "required", "must"
}

type NetServiceConfig struct {
	Address string     `json:"address" yaml:"address"` // Listening address, e.g., ":8080"
	TLS     *TLSConfig `json:"tls" yaml:"tls"`         // [Optional] TLS configuration, if provided, the service will use TLS for secure communication
}

func (cfg *NetServiceConfig) ListenTCP() (IAdapter, error) {
	if cfg.TLS != nil {
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS12,
		}

		cert, _, err := loadX509KeyPair(cfg.TLS.Cert, cfg.TLS.Key)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{*cert}

		switch cfg.TLS.ClientAuthType {
		case "none":
			tlsConfig.ClientAuth = tls.NoClientCert
		case "optional":
			tlsConfig.ClientAuth = tls.VerifyClientCertIfGiven
		case "required":
			tlsConfig.ClientAuth = tls.RequireAnyClientCert
		case "must":
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		default:
			tlsConfig.ClientAuth = tls.NoClientCert
		}

		if tlsConfig.ClientAuth != tls.NoClientCert {
			tlsConfig.ClientCAs = x509.NewCertPool()
			for _, ca := range cfg.TLS.CAs {
				cacert, err := loadData(ca)
				if err != nil {
					return nil, fmt.Errorf("failed to load CA certificate: %w", err)
				}
				tlsConfig.ClientCAs.AppendCertsFromPEM(cacert)
			}
		}

		listener, err := tls.Listen("tcp", cfg.Address, tlsConfig)
		if err != nil {
			return nil, ErrPortBindingFailed
		}

		return listener, nil
	} else {
		return net.Listen("tcp", cfg.Address)
	}
}
