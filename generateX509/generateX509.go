/*
 * Example: https://golang.org/src/crypto/tls/generate_cert.go
 *
 * Create time: 2021/09/06
 * Update time: 2021/09/16
 */
package generateX509

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"
)

type Info struct {
	Hostname   string    /* 主机名，默认为本机主机名 */
	ValidFrom  time.Time /* 创建日期，默认为当前时间 */
	ValidFor   int       /* 有效期(天)，默认为365天 */
	IsCA       bool      /* 这个证书是否应该是它自己的证书颁发机构，默认为否 */
	RsaBits    int       /* RSA私钥长度，默认为2048 */
	EcdsaCurve string    /* 用于生成密钥的 ECDSA 曲线。 有效值为 P224、P256（推荐）、P384、P521，默认为无 */
	Ed25519Key bool      /* 生成Ed25519私钥，默认为否 */
}

func (x *Info) defaultHostname() {
	if x.Hostname == "" {
		hostname, err := os.Hostname()
		if err != nil {
			x.Hostname = "unknown"
			return
		}
		x.Hostname = hostname
	}
}
func (x *Info) defaultValidFrom() {
	if x.ValidFrom.IsZero() {
		x.ValidFrom = time.Now()
		return
	}
}
func (x *Info) defaultValidFor() {
	if x.ValidFor == 0 {
		x.ValidFor = 365
	}
}
func (x *Info) defaultIsCA() {}
func (x *Info) defaultRsaBits() {
	if x.RsaBits == 0 {
		x.RsaBits = 2048
	}
}
func (x *Info) defaultEcdsaCurve() {}
func (x *Info) defaultEd25519Key() {}

// Generate 生成公私钥密钥对
func (x *Info) Generate() (*tls.Certificate, error) {

	// 设置默认值
	x.defaultHostname()
	x.defaultValidFrom()
	x.defaultValidFor()
	x.defaultIsCA()
	x.defaultRsaBits()
	x.defaultEcdsaCurve()
	x.defaultEd25519Key()

	var privy interface{}
	var err error
	switch x.EcdsaCurve {
	case "":
		if x.Ed25519Key {
			_, privy, err = ed25519.GenerateKey(rand.Reader)
		} else {
			privy, err = rsa.GenerateKey(rand.Reader, x.RsaBits)
		}
	case "P224":
		privy, err = ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	case "P256":
		privy, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case "P384":
		privy, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case "P521":
		privy, err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	default:
		return nil, fmt.Errorf("unrecognized elliptic curve: %q", x.EcdsaCurve)
	}
	if err != nil {
		return nil, err
	}

	// ECDSA, ED25519 and RSA subject keys should have the DigitalSignature
	// KeyUsage bits set in the x509.Certificate template
	keyUsage := x509.KeyUsageDigitalSignature
	// Only RSA subject keys should have the KeyEncipherment KeyUsage bits set. In
	// the context of TLS this KeyUsage is particular to RSA key exchange and
	// authentication.
	if _, isRSA := privy.(*rsa.PrivateKey); isRSA {
		keyUsage |= x509.KeyUsageKeyEncipherment
	}

	notBefore := x.ValidFrom
	notAfter := notBefore.Add(time.Duration(x.ValidFor) * 24 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:         x.Hostname,
			Organization:       []string{x.Hostname},
			OrganizationalUnit: []string{x.Hostname},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              keyUsage,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,

		DNSNames: []string{x.Hostname},
	}
	if x.IsCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	publicKey := func(privy interface{}) interface{} {
		switch k := privy.(type) {
		case *rsa.PrivateKey:
			return &k.PublicKey
		case *ecdsa.PrivateKey:
			return &k.PublicKey
		case ed25519.PrivateKey:
			return k.Public().(ed25519.PublicKey)
		default:
			return nil
		}
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(privy), privy)
	if err != nil {
		return nil, err
	}

	return &tls.Certificate{
		Certificate: [][]byte{derBytes},
		PrivateKey:  privy,
	}, nil
}

// Pem 转换为pem格式公私钥对
func (x *Info) Pem(pk *tls.Certificate) ([]byte, []byte, error) {
	var crt bytes.Buffer
	if err := pem.Encode(&crt, &pem.Block{Type: "CERTIFICATE", Bytes: pk.Certificate[0]}); err != nil {
		return nil, nil, err
	}

	var key bytes.Buffer
	privyBytes, _ := x509.MarshalPKCS8PrivateKey(pk.PrivateKey)
	if err := pem.Encode(&key, &pem.Block{Type: "PRIVATE KEY", Bytes: privyBytes}); err != nil {
		return nil, nil, err
	}

	return crt.Bytes(), key.Bytes(), nil
}
