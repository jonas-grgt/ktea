package kadmin

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"ktea/config"
	"ktea/styles"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go/modules/redpanda"
)

func TestSaramaKadminWithMTLS(t *testing.T) {
	ctx := context.Background()

	caCertPEM, caKeyPEM, err := generateCA()
	if err != nil {
		t.Fatalf("Failed to generate CA: %v", err)
	}

	serverCertPEM, serverKeyPEM, err := generateSignedCertificate(caCertPEM, caKeyPEM, "localhost", false)
	if err != nil {
		t.Fatalf("Failed to generate server certificate: %v", err)
	}

	redpandaContainer, err := redpanda.Run(
		ctx,
		"docker.redpanda.com/redpandadata/redpanda:v23.3.10",
		redpanda.WithTLS(serverCertPEM, serverKeyPEM),
	)
	if err != nil {
		t.Fatalf("Failed to start Redpanda container: %v", err)
	}

	t.Cleanup(func() {
		if err := redpandaContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate Redpanda container: %v", err)
		}
	})

	brokerAddress, err := redpandaContainer.KafkaSeedBroker(ctx)
	if err != nil {
		t.Fatalf("Failed to get Kafka seed broker: %v", err)
	}

	t.Logf("Broker address: %s", brokerAddress)

	clientCertPEM, clientKeyPEM, err := generateSignedCertificate(caCertPEM, caKeyPEM, "client", true)
	if err != nil {
		t.Fatalf("Failed to generate client certificate: %v", err)
	}

	certFile, err := os.CreateTemp("", "client-cert-*.pem")
	if err != nil {
		t.Fatalf("Failed to create temp cert file: %v", err)
	}
	defer os.Remove(certFile.Name())
	certFile.Write(clientCertPEM)
	certFile.Close()

	keyFile, err := os.CreateTemp("", "client-key-*.pem")
	if err != nil {
		t.Fatalf("Failed to create temp key file: %v", err)
	}
	defer os.Remove(keyFile.Name())
	keyFile.Write(clientKeyPEM)
	keyFile.Close()

	caFile, err := os.CreateTemp("", "ca-cert-*.pem")
	if err != nil {
		t.Fatalf("Failed to create temp CA file: %v", err)
	}
	defer os.Remove(caFile.Name())
	caFile.Write(caCertPEM)
	caFile.Close()

	t.Run("mTLS with client certificate", func(t *testing.T) {

		cluster := &config.Cluster{
			Name:             "test-mtls",
			Color:            styles.ColorGreen,
			Active:           true,
			BootstrapServers: []string{brokerAddress},
			TLSConfig: config.TLSConfig{
				Enable:     true,
				SkipVerify: false,
				CACertPath: caFile.Name(),
				ClientCert: certFile.Name(),
				ClientKey:  keyFile.Name(),
			},
		}

		cfg := ToSaramaCfg(cluster)

		if !cfg.Net.TLS.Enable {
			t.Fatal("Expected TLS to be enabled")
		}

		if cfg.Net.TLS.Config == nil {
			t.Fatal("Expected TLS config to be set")
		}

		if len(cfg.Net.TLS.Config.Certificates) != 1 {
			t.Fatalf("Expected 1 certificate, got %d", len(cfg.Net.TLS.Config.Certificates))
		}

		if cfg.Net.TLS.Config.RootCAs == nil {
			t.Fatal("Expected RootCAs to be set")
		}

		_, err = NewSaramaKadmin(cluster)
		if err != nil {
			t.Fatalf("Failed to connect with mTLS: %v", err)
		}
	})

	t.Run("TLS without client certificate", func(t *testing.T) {
		cluster := &config.Cluster{
			Name:             "test-tls",
			Color:            styles.ColorGreen,
			Active:           true,
			BootstrapServers: []string{brokerAddress},
			TLSConfig: config.TLSConfig{
				Enable:     true,
				SkipVerify: true,
				CACertPath: "",
				ClientCert: "",
				ClientKey:  "",
			},
		}

		cfg := ToSaramaCfg(cluster)

		if !cfg.Net.TLS.Enable {
			t.Fatal("Expected TLS to be enabled")
		}

		if cfg.Net.TLS.Config == nil {
			t.Fatal("Expected TLS config to be set")
		}

		if len(cfg.Net.TLS.Config.Certificates) != 0 {
			t.Fatalf("Expected 0 certificates, got %d", len(cfg.Net.TLS.Config.Certificates))
		}

		_, err := NewSaramaKadmin(cluster)
		if err != nil {
			t.Fatalf("Failed to connect with TLS: %v", err)
		}
	})
}

func generateCA() ([]byte, []byte, error) {
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	caTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test CA"},
			CommonName:   "Test CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	caCertDER, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, err
	}

	caCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCertDER})
	caKeyPEM, err := x509.MarshalECPrivateKey(caKey)
	if err != nil {
		return nil, nil, err
	}
	caKeyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: caKeyPEM})

	return caCertPEM, caKeyPEM, nil
}

func generateSignedCertificate(caCertPEM, caKeyPEM []byte, name string, isClient bool) ([]byte, []byte, error) {
	caCertBlock, _ := pem.Decode(caCertPEM)
	caCert, err := x509.ParseCertificate(caCertBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}

	caKeyBlock, _ := pem.Decode(caKeyPEM)
	caKey, err := x509.ParseECPrivateKey(caKeyBlock.Bytes)
	if err != nil {
		return nil, nil, err
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Test"},
			CommonName:   name,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost", "redpanda"},
	}

	if isClient {
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
	} else {
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, caCert, &key.PublicKey, caKey)
	if err != nil {
		return nil, nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	keyBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, nil, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})

	return certPEM, keyPEM, nil
}
