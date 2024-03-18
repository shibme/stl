package asx

import (
	"crypto/rand"
	"crypto/sha256"

	"dev.shib.me/xipher/internal/ecc"
	"dev.shib.me/xipher/internal/kyb"
)

// PrivateKey represents a private key.
type PrivateKey struct {
	key        *[]byte
	eccPrivKey *ecc.PrivateKey
	kybPrivKey *kyb.PrivateKey
	pubKeyECC  *PublicKey
	pubKeyKyb  *PublicKey
}

// PublicKey represents a public key.
type PublicKey struct {
	ePub *ecc.PublicKey
	kPub *kyb.PublicKey
}

// Bytes returns the bytes of the private key.
func (privateKey *PrivateKey) Bytes() []byte {
	return *privateKey.key
}

// NewPrivateKey generates a new random private key.
func NewPrivateKey() (*PrivateKey, error) {
	key := make([]byte, PrivateKeyLength)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return ParsePrivateKey(key)
}

// ParsePrivateKey returns the instance private key for given bytes. Please use exactly 64 bytes.
func ParsePrivateKey(key []byte) (*PrivateKey, error) {
	if len(key) != PrivateKeyLength {
		return nil, errInvalidPrivateKeyLength
	}
	return &PrivateKey{
		key: &key,
	}, nil
}

func (privateKey *PrivateKey) getEccPrivKey() (*ecc.PrivateKey, error) {
	if privateKey.eccPrivKey == nil {
		eccPrivKeyBytes := sha256.Sum256(*privateKey.key)
		eccPrivKey, err := ecc.ParsePrivateKey(eccPrivKeyBytes[:])
		if err != nil {
			return nil, err
		}
		privateKey.eccPrivKey = eccPrivKey
	}
	return privateKey.eccPrivKey, nil
}

func (privateKey *PrivateKey) getKybPrivKey() (*kyb.PrivateKey, error) {
	if privateKey.kybPrivKey == nil {
		kybPrivKey, err := kyb.NewPrivateKeyForSeed(*privateKey.key)
		if err != nil {
			return nil, err
		}
		privateKey.kybPrivKey = kybPrivKey
	}
	return privateKey.kybPrivKey, nil
}

// PublicKey returns the ecc public key corresponding to the private key. The public key is derived from the private key.
func (privateKey *PrivateKey) PublicKeyECC() (*PublicKey, error) {
	if privateKey.pubKeyECC == nil {
		eccPrivKeyBytes := sha256.Sum256(*privateKey.key)
		eccPrivKey, err := ecc.ParsePrivateKey(eccPrivKeyBytes[:])
		if err != nil {
			return nil, err
		}
		eccPubKey, err := eccPrivKey.PublicKey()
		if err != nil {
			return nil, err
		}
		privateKey.pubKeyECC = &PublicKey{
			ePub: eccPubKey,
		}
	}
	return privateKey.pubKeyECC, nil
}

// PublicKey returns the kyber public key corresponding to the private key. The public key is derived from the private key.
func (privateKey *PrivateKey) PublicKeyKyber() (*PublicKey, error) {
	if privateKey.pubKeyKyb == nil {
		kybPrivKey, err := kyb.NewPrivateKeyForSeed(*privateKey.key)
		if err != nil {
			return nil, err
		}
		kybPubKey, err := kybPrivKey.PublicKey()
		if err != nil {
			return nil, err
		}
		privateKey.pubKeyKyb = &PublicKey{
			kPub: kybPubKey,
		}
	}
	return privateKey.pubKeyKyb, nil
}

// Bytes returns the public key as bytes.
func (publicKey *PublicKey) Bytes() ([]byte, error) {
	if publicKey.ePub != nil {
		return append([]byte{AlgoECC}, publicKey.ePub.Bytes()...), nil
	} else if publicKey.kPub != nil {
		kybPubKeyBytes, err := publicKey.kPub.Bytes()
		if err != nil {
			return nil, err
		}
		return append([]byte{AlgoKyber}, kybPubKeyBytes...), nil
	} else {
		return nil, errInvalidPublicKey
	}
}

// GetPublicKey returns the instance of public key for given bytes.
func ParsePublicKey(key []byte) (*PublicKey, error) {
	if len(key) < MinPublicKeyLength {
		return nil, errInvalidPublicKeyLength
	}
	switch key[0] {
	case AlgoECC:
		eccPubKey, err := ecc.ParsePublicKey(key[1:])
		if err != nil {
			return nil, err
		}
		return &PublicKey{
			ePub: eccPubKey,
		}, nil
	case AlgoKyber:
		kybPubKey, err := kyb.ParsePublicKey(key[1:])
		if err != nil {
			return nil, err
		}
		return &PublicKey{
			kPub: kybPubKey,
		}, nil
	default:
		return nil, errInvalidPublicKey
	}
}

// func (publicKey *PublicKey) getEncrypter() (*encrypter, error) {
// 	if publicKey.encrypter == nil {
// 		ephPrivKey := make([]byte, KeyLength)
// 		if _, err := rand.Read(ephPrivKey); err != nil {
// 			return nil, err
// 		}
// 		ephPubKey, err := curve25519.X25519(ephPrivKey, curve25519.Basepoint)
// 		if err != nil {
// 			return nil, err
// 		}
// 		sharedKey, err := curve25519.X25519(ephPrivKey, *publicKey.key)
// 		if err != nil {
// 			return nil, err
// 		}
// 		cipher, err := xcp.New(sharedKey)
// 		if err != nil {
// 			return nil, err
// 		}
// 		publicKey.encrypter = &encrypter{
// 			ephPubKey: ephPubKey,
// 			cipher:    cipher,
// 		}
// 	}
// 	return publicKey.encrypter, nil
// }
