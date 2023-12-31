package xipher

import (
	"crypto/rand"
	"fmt"

	"gopkg.shib.me/xipher/internal/ecc"
	"gopkg.shib.me/xipher/internal/symmcipher"
)

type PrivateKey struct {
	password     *[]byte
	spec         *kdfSpec
	key          []byte
	symEncrypter *symmcipher.Cipher
	publicKey    *PublicKey
	specKeyMap   map[string][]byte
}

// NewPrivateKeyForPassword creates a new private key for the given password.
func NewPrivateKeyForPassword(password []byte) (*PrivateKey, error) {
	spec, err := newSpec()
	if err != nil {
		return nil, err
	}
	return newPrivateKeyForPwdAndSpec(password, spec)
}

// NewPrivateKeyForPasswordAndSpec creates a new private key for the given password and kdf spec.
func NewPrivateKeyForPasswordAndSpec(password []byte, iterations, memory, threads uint8) (*PrivateKey, error) {
	spec, err := newSpec()
	if err != nil {
		return nil, err
	}
	spec.setIterations(iterations).setMemory(memory).setThreads(threads)
	return newPrivateKeyForPwdAndSpec(password, spec)
}

func newPrivateKeyForPwdAndSpec(password []byte, spec *kdfSpec) (*PrivateKey, error) {
	if len(password) == 0 {
		return nil, errInvalidPassword
	}
	privateKey := pwdXipherMap[string(password)]
	if privateKey == nil {
		privateKey = &PrivateKey{
			password: &password,
			spec:     spec,
		}
		privateKey.key = spec.getCipherKey(*privateKey.password)
		pwdXipherMap[string(*privateKey.password)] = privateKey
		privateKey.specKeyMap = make(map[string][]byte)
		privateKey.specKeyMap[string(privateKey.spec.bytes())] = privateKey.key
	}
	return privateKey, nil
}

// NewPrivateKey creates a new random private key.
func NewPrivateKey() (*PrivateKey, error) {
	key := make([]byte, cipherKeyLength)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	privateKey := &PrivateKey{
		key: key,
	}
	keyXipherMap[string(key)] = privateKey
	return privateKey, nil
}

// ParsePrivateKey parses the given bytes and returns a corresponding private key. the given bytes must be 32 bytes long.
func ParsePrivateKey(key []byte) (*PrivateKey, error) {
	if len(key) != PrivateKeyLength {
		return nil, fmt.Errorf("invalid private key length: expected %d, got %d", PrivateKeyLength, len(key))
	}
	privateKey := keyXipherMap[string(key)]
	if privateKey == nil {
		privateKey = &PrivateKey{
			key: key,
		}
		keyXipherMap[string(key)] = privateKey
	}
	return privateKey, nil
}

func (privateKey *PrivateKey) isPwdBased() bool {
	return privateKey.password != nil && privateKey.spec != nil
}

// Bytes returns the private key as bytes only if it is not password based.
func (privateKey *PrivateKey) Bytes() ([]byte, error) {
	if privateKey.password != nil || privateKey.spec != nil {
		return nil, errPrivKeyUnavailableForPwd
	}
	return privateKey.key, nil
}

// PublicKey returns the public key corresponding to the private key.
func (privateKey *PrivateKey) PublicKey() (*PublicKey, error) {
	if privateKey.publicKey == nil {
		eccPrivKey, err := ecc.GetPrivateKey(privateKey.key)
		if err != nil {
			return nil, err
		}
		eccPubKey, err := eccPrivKey.PublicKey()
		if err != nil {
			return nil, err
		}
		privateKey.publicKey = &PublicKey{
			publicKey: eccPubKey,
			spec:      privateKey.spec,
		}
	}
	return privateKey.publicKey, nil
}

type PublicKey struct {
	publicKey *ecc.PublicKey
	spec      *kdfSpec
}

// ParsePublicKey parses the given bytes and returns a corresponding public key. the given bytes must be 51 bytes long.
func ParsePublicKey(pubKeyBytes []byte) (*PublicKey, error) {
	if len(pubKeyBytes) != PublicKeyLength {
		return nil, fmt.Errorf("invalid public key length: expected %d, got %d", PublicKeyLength, len(pubKeyBytes))
	}
	eccPubKey, err := ecc.GetPublicKey(pubKeyBytes[:cipherKeyLength])
	if err != nil {
		return nil, err
	}
	publicKey := &PublicKey{
		publicKey: eccPubKey,
	}
	specBytes := pubKeyBytes[cipherKeyLength:]
	if [kdfSpecLength]byte(specBytes) != [kdfSpecLength]byte{} {
		publicKey.spec, err = parseKdfSpec(specBytes)
		if err != nil {
			return nil, err
		}
	}
	return publicKey, nil
}

func (publicKey *PublicKey) isPwdBased() bool {
	return publicKey.spec != nil
}

// Bytes returns the public key as bytes.
func (publicKey *PublicKey) Bytes() []byte {
	if publicKey.spec != nil {
		return append(publicKey.publicKey.Bytes(), publicKey.spec.bytes()...)
	} else {
		return append(publicKey.publicKey.Bytes(), make([]byte, kdfSpecLength)...)
	}
}
