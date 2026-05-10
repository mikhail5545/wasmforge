/*
 * Copyright (c) 2026. Mikhail Kulik.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package materials

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"time"

	materialmodel "github.com/mikhail5545/wasmforge/internal/models/storage/crypto/material"
	"github.com/mikhail5545/wasmforge/internal/storage/core"
	"go.uber.org/zap"
)

type (
	ValidationInput struct {
		ExpectedKind materialmodel.CryptoMaterialKind
		Parts        []ValidationPart
	}

	ValidationPart struct {
		Name      string
		Reader    io.Reader
		SizeHint  int64
		ObjectRef core.ObjectRef
		Metadata  map[string]any
	}

	ValidationResult struct {
		Kind materialmodel.CryptoMaterialKind

		Certs        []CertInfo
		PublicKeys   []PublicKeyInfo
		PrivateKeys  []PrivateKeyInfo
		TrustBundles []TrustBundle
		CAPairs      []CAPair

		HasPrivateKeyMaterial bool
		RequiresEncryption    bool
	}

	PrivateKeyInfo struct {
		Algorithm string
		Details   string
		PublicKey PublicKeyInfo
		Encrypted bool
		BlockType string
		PartName  string
		ObjectRef core.ObjectRef
		Metadata  map[string]any
	}

	CertInfo struct {
		CertSHA256Hex string
		Subject       string
		Issuer        string
		NotBefore     time.Time
		NotAfter      time.Time
		IsCA          bool
		SerialHex     string
		PublicKey     PublicKeyInfo

		Cert *x509.Certificate

		PartName  string
		ObjectRef core.ObjectRef
		Metadata  map[string]any
	}

	CAPair struct {
		Cert CertInfo
		Key  PrivateKeyInfo

		IsCA    bool
		Matched bool
	}

	TrustBundle struct {
		Certs       []CertInfo
		CertCount   int
		AllCA       bool
		Fingerprint string
	}

	PublicKeyInfo struct {
		HasPrivateKey                 bool
		Fingerprint                   string
		Algorithm                     string
		Details                       string // e.g. RSA-2048, ECDSA-P256
		SubjectPublicKeyInfoSha256Hex string
		PartName                      string
		ObjectRef                     core.ObjectRef
		Metadata                      map[string]any
	}
)

// Validator validates input certificates and public keys before they are stored in the object store.
// Supported types are:
//
//   - CERTIFICATE
//   - PRIVATE KEY
//   - RSA PRIVATE KEY
//   - EC PRIVATE KEY
//
// Including provided PEM block in the certificate manager allows to validate both certificates and public keys using the same code.
// Extracting from:
//
//   - PUBLIC CERTIFICATE
//   - CA BUNDLE
//   - TRUST BUNDLE
//   - KEY PAIR
//
// All other combinations/types will be rejected and will cause error responses.
type Validator struct {
	logger *zap.Logger
}

func NewValidator(logger *zap.Logger) *Validator {
	return &Validator{
		logger: logger.With(zap.String("domain", "storage"), zap.String("component", "certificate_validator")),
	}
}

// Validate validates the provided input parts, which can be multiple PEM blocks, and determines the certificate kind
// (PublicCert, KeyPair, TrustBundle, CABundle) based on the content.
//
// It returns a ValidationResult containing parsed information about the certificates and keys
// found in the input. If the input is invalid or does not match expected formats, it returns an error.
func (v *Validator) Validate(input ValidationInput) (*ValidationResult, error) {
	if len(input.Parts) == 0 {
		return nil, core.NewInvalidObjectRef("no input for validation")
	}

	var out ValidationResult
	var allPemBytes []byte

	for _, part := range input.Parts {
		pemBytes, err := io.ReadAll(io.LimitReader(part.Reader, part.SizeHint+1))
		if err != nil {
			return nil, fmt.Errorf("failed to read part '%s': %w", part.Name, err)
		}
		allPemBytes = append(allPemBytes, pemBytes...)
		allPemBytes = append(allPemBytes, '\n')

		var anyBlockFound bool
		for {
			var block *pem.Block
			block, pemBytes = pem.Decode(pemBytes)
			if block == nil {
				break
			}
			anyBlockFound = true

			switch block.Type {
			case "PUBLIC KEY", "RSA PUBLIC KEY":
				info, err := v.ValidatePublicKey(block)
				if err != nil {
					return nil, core.NewInvalidObjectFormatError(err)
				}
				info.PartName = part.Name
				info.ObjectRef = part.ObjectRef
				info.Metadata = part.Metadata
				out.PublicKeys = append(out.PublicKeys, info)
			case "PRIVATE KEY", "RSA PRIVATE KEY", "EC PRIVATE KEY", "ENCRYPTED PRIVATE KEY":
				info, err := v.ValidatePrivateKey(block)
				if err != nil {
					return nil, core.NewInvalidObjectFormatError(err)
				}
				info.PartName = part.Name
				info.ObjectRef = part.ObjectRef
				info.Metadata = part.Metadata
				out.PrivateKeys = append(out.PrivateKeys, info)
			case "CERTIFICATE":
				info, err := v.ValidateCertificate(block)
				if err != nil {
					return nil, core.NewInvalidObjectFormatError(err)
				}
				info.PartName = part.Name
				info.ObjectRef = part.ObjectRef
				info.Metadata = part.Metadata
				out.Certs = append(out.Certs, info)
			default:
				return nil, core.NewInvalidObjectFormatError("unsupported block type")
			}
		}
		if !anyBlockFound {
			return nil, core.NewInvalidObjectFormatError("no valid PEM blocks found in part " + part.Name)
		}
	}

	trustBundle, built := v.BuildTrustBundle(&out, allPemBytes)
	if built {
		out.TrustBundles = append(out.TrustBundles, trustBundle)
	}

	caPairs, err := v.BuildCAPairs(&out)
	if err != nil {
		return nil, core.NewInvalidObjectFormatError(err)
	}
	out.CAPairs = caPairs

	if err := v.determineKind(&out); err != nil {
		return nil, core.NewInvalidObjectFormatError(err)
	}

	return &out, nil
}

func (v *Validator) determineKind(result *ValidationResult) error {
	var foundKinds []materialmodel.CryptoMaterialKind

	numCerts := len(result.Certs)
	numPubKeys := len(result.PublicKeys)
	numPrivKeys := len(result.PrivateKeys)
	numCAPairs := len(result.CAPairs)

	// Check for PublicCert (exactly 1 certificate, no keys)
	if numCerts == 1 && numPubKeys == 0 && numPrivKeys == 0 {
		foundKinds = append(foundKinds, materialmodel.CryptoMaterialKindPublicCert)
	}

	// Check for KeyPair (exactly 1 private key, exactly 1 matching certificate or public key)
	if numPrivKeys == 1 {
		if (numCerts == 1 && numPubKeys == 0 && numCAPairs == 1) || (numPubKeys == 1 && numCerts == 0) || (numCerts == 0 && numPubKeys == 0) {
			foundKinds = append(foundKinds, materialmodel.CryptoMaterialKindKeyPair)
		}
	}

	// Check for TrustBundle (multiple certificates, no keys, not all CA)
	if numCerts > 1 && numPubKeys == 0 && numPrivKeys == 0 && len(result.TrustBundles) > 0 {
		if !result.TrustBundles[0].AllCA {
			foundKinds = append(foundKinds, materialmodel.CryptoMaterialKindTrustBundle)
		}
	}

	// Check for CABundle (multiple certificates, all CA, no keys)
	if numCerts > 1 && numPubKeys == 0 && numPrivKeys == 0 && len(result.TrustBundles) > 0 {
		if result.TrustBundles[0].AllCA {
			foundKinds = append(foundKinds, materialmodel.CryptoMaterialKindCABundle)
		}
	}

	if len(foundKinds) == 0 {
		return fmt.Errorf("could not determine exactly one certificate kind for the provided input")
	}
	if len(foundKinds) > 1 {
		return core.NewAmbiguousInputError("matches multiple certificate kinds")
	}

	result.Kind = foundKinds[0]

	if numPrivKeys > 0 {
		result.HasPrivateKeyMaterial = true
		for _, pk := range result.PrivateKeys {
			if !pk.Encrypted {
				result.RequiresEncryption = true
				break
			}
		}
	}

	return nil
}

func (v *Validator) ValidatePrivateKey(block *pem.Block) (PrivateKeyInfo, error) {
	var priv crypto.PrivateKey
	var parseErr error
	var encrypted bool

	switch block.Type {
	case "RSA PRIVATE KEY":
		encrypted = false
		priv, parseErr = x509.ParsePKCS1PrivateKey(block.Bytes)
	case "PRIVATE KEY":
		encrypted = false
		priv, parseErr = x509.ParsePKCS8PrivateKey(block.Bytes)
	case "EC PRIVATE KEY":
		encrypted = false
		priv, parseErr = x509.ParseECPrivateKey(block.Bytes)
	case "ENCRYPTED PRIVATE KEY":
		encrypted = true
	default:
		return PrivateKeyInfo{}, fmt.Errorf("unsupported PEM block type (Private Key): %w", parseErr)
	}
	if parseErr != nil {
		return PrivateKeyInfo{}, fmt.Errorf("failed to parse private key: %w", parseErr)
	}

	if encrypted {
		return PrivateKeyInfo{
			Algorithm: "Unknown",
			Details:   "",
			Encrypted: true,
			PublicKey: PublicKeyInfo{
				HasPrivateKey:                 true,
				Algorithm:                     "Unknown",
				Details:                       "",
				SubjectPublicKeyInfoSha256Hex: "",
			},
			BlockType: block.Type,
		}, nil
	}

	info, err := privateKeyInfo(block.Type, priv, encrypted)
	if err != nil {
		return PrivateKeyInfo{}, fmt.Errorf("failed to extract private key info: %w", err)
	}
	return info, nil
}

func (v *Validator) ValidateCertificate(block *pem.Block) (CertInfo, error) {
	if block.Type != "CERTIFICATE" {
		return CertInfo{}, core.NewInvalidObjectFormatError("unsupported PEM block type for certificate: " + block.Type)
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return CertInfo{}, core.NewInvalidObjectFormatError("failed to parse certificate: " + err.Error())
	}

	pkInfo, err := publicKeyInfo(cert.PublicKey)
	if err != nil {
		return CertInfo{}, fmt.Errorf("failed to parse public key info in certificate: " + err.Error())
	}

	res := CertInfo{
		CertSHA256Hex: hex.EncodeToString(cert.Raw),
		Issuer:        cert.Issuer.String(),
		Subject:       cert.Subject.String(),
		SerialHex:     serialToHex(cert.SerialNumber),
		NotBefore:     cert.NotBefore,
		NotAfter:      cert.NotAfter,
		IsCA:          cert.IsCA,
		PublicKey:     pkInfo,
		Cert:          cert,
	}

	return res, nil
}

func (v *Validator) ValidatePublicKey(block *pem.Block) (PublicKeyInfo, error) {
	var pub crypto.PublicKey
	var parseErr error

	switch block.Type {
	case "PUBLIC KEY":
		// PKIX / SubjectPublicKeyInfo
		pub, parseErr = x509.ParsePKIXPublicKey(block.Bytes)
	case "RSA PUBLIC KEY":
		pub, parseErr = x509.ParsePKCS1PublicKey(block.Bytes)
	default:
		return PublicKeyInfo{}, core.NewInvalidObjectFormatError("unsupported PEM block type for public key: " + block.Type)
	}
	if parseErr != nil {
		return PublicKeyInfo{}, core.NewInvalidObjectFormatError("failed to parse public key: " + parseErr.Error())
	}

	info, err := publicKeyInfo(pub)
	if err != nil {
		return PublicKeyInfo{}, fmt.Errorf("failed to parse public key info: %w", err)
	}
	res := PublicKeyInfo{
		Algorithm:                     info.Algorithm,
		Details:                       info.Details,
		SubjectPublicKeyInfoSha256Hex: info.SubjectPublicKeyInfoSha256Hex,
	}

	return res, nil
}

func (v *Validator) BuildTrustBundle(result *ValidationResult, bundlePEM []byte) (TrustBundle, bool) {
	pool := x509.NewCertPool()

	if ok := pool.AppendCertsFromPEM(bundlePEM); !ok {
		return TrustBundle{}, false
	}
	return TrustBundle{
		Certs:       result.Certs,
		CertCount:   len(result.Certs),
		AllCA:       allCertsCA(result.Certs),
		Fingerprint: sha256Hex(bundlePEM),
	}, true
}

// MatchCAPair tries to match a cert with a private key by comparing SPKI fingerprints
func (v *Validator) MatchCAPair(cert CertInfo, key PrivateKeyInfo) bool {
	if cert.PublicKey.SubjectPublicKeyInfoSha256Hex == "" || key.PublicKey.SubjectPublicKeyInfoSha256Hex == "" {
		return false
	}
	return bytes.Equal(
		[]byte(cert.PublicKey.SubjectPublicKeyInfoSha256Hex),
		[]byte(key.PublicKey.SubjectPublicKeyInfoSha256Hex),
	)
}

// BuildCAPairs forms all matching cert<->key pairs found in validation result.
func (v *Validator) BuildCAPairs(result *ValidationResult) ([]CAPair, error) {
	var pairs []CAPair
	for _, c := range result.Certs {
		for _, k := range result.PrivateKeys {
			if k.Encrypted {
				continue
			}
			matched := v.MatchCAPair(c, k)
			if matched {

				pairs = append(pairs, CAPair{
					IsCA:    c.IsCA,
					Cert:    c,
					Key:     k,
					Matched: true,
				})
			}
		}
	}
	return pairs, nil
}

func allCertsCA(certs []CertInfo) bool {
	for _, cert := range certs {
		if !cert.IsCA {
			return false
		}
	}
	return true
}

func privateKeyInfo(blockType string, key any, encrypted bool) (PrivateKeyInfo, error) {
	pub, err := publicFromPrivate(key)
	if err != nil {
		return PrivateKeyInfo{}, fmt.Errorf("failed to derive public key from private key: %w", err)
	}
	pkInfo, err := publicKeyInfo(pub)
	if err != nil {
		return PrivateKeyInfo{}, fmt.Errorf("failed to parse public key info (from private): %w", err)
	}
	alg, det := keyAlgDetails(key)

	return PrivateKeyInfo{
		Algorithm: alg,
		Details:   det,
		PublicKey: pkInfo,
		Encrypted: encrypted,
		BlockType: blockType,
	}, nil
}

func publicFromPrivate(key any) (crypto.PublicKey, error) {
	switch k := key.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey, nil
	case *ecdsa.PrivateKey:
		return &k.PublicKey, nil
	case ed25519.PrivateKey:
		return k.Public().(ed25519.PublicKey), nil
	default:
		return nil, fmt.Errorf("unsupported private key type: %T", k)
	}
}

func keyAlgDetails(key any) (alg, det string) {
	switch k := key.(type) {
	case *rsa.PrivateKey:
		return "RSA", fmt.Sprintf("RSA-%d", k.N.BitLen())
	case *ecdsa.PrivateKey:
		return "ECDSA", "ECDSA-" + curveName(k.Curve)
	case ed25519.PrivateKey:
		return "Ed25519", "Ed25519"
	default:
		return fmt.Sprintf("Unknown(%T)", key), ""
	}
}

func publicKeyInfo(pub any) (PublicKeyInfo, error) {
	subjectPublicKeyInfoDER, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return PublicKeyInfo{}, fmt.Errorf("failed to marshal public key: " + err.Error())
	}

	info := PublicKeyInfo{SubjectPublicKeyInfoSha256Hex: sha256Hex(subjectPublicKeyInfoDER)}

	switch k := pub.(type) {
	case *rsa.PublicKey:
		info.Algorithm = "RSA"
		info.Details = fmt.Sprintf("RSA-%d", k.N.BitLen())
	case *ecdsa.PublicKey:
		info.Algorithm = "ECDSA"
		info.Details = fmt.Sprintf("ECDSA-%s", curveName(k.Curve))
	case ed25519.PublicKey:
		info.Algorithm = "Ed25519"
		info.Details = "Ed25519"
	default:
		return PublicKeyInfo{}, fmt.Errorf("unsupported public key algorithm: %T", pub)
	}
	return info, nil
}

func curveName(c elliptic.Curve) string {
	switch c {
	case elliptic.P256():
		return "P-256"
	case elliptic.P384():
		return "P-384"
	case elliptic.P521():
		return "P-521"
	default:
		// fallback
		return c.Params().Name
	}
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func serialToHex(n *big.Int) string {
	if n == nil {
		return ""
	}
	return fmt.Sprintf("%x", n)
}
