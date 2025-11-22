package kadmin

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"
)

type CertValidationFunc func(certFile string) error

func CertValidator(certFile string) error {
	if _, err := os.Stat(certFile); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("CA Certificate file does not exist at path: %s", certFile)
	}

	data, err := os.ReadFile(certFile)
	if err != nil {
		return fmt.Errorf("CA Certificate file could not be read at path: %s", certFile)
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "CERTIFICATE" {
		return errors.New("file does not contain a PEM-encoded certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("invalid X.509 certificate: %w", err)
	}

	now := time.Now()
	if now.Before(cert.NotBefore) || now.After(cert.NotAfter) {
		return fmt.Errorf(
			"certificate is not valid at current time (valid from %s to %s)",
			cert.NotBefore, cert.NotAfter,
		)
	}

	return nil
}
