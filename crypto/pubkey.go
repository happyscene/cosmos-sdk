package crypto

import (
	"github.com/pkg/errors"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/multisig"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

// public key sentinel errors
var (
	ErrInvalidPubKeySecp256k1Length = errors.New("invalid PubKeySecp256k1 length")
	ErrInvalidPubKeySecp256k1       = errors.New("incompatible PubKeySecp256k1")

	ErrInvalidPubKeyEd25519Length = errors.New("invalid PubKeyEd25519 length")
	ErrInvalidPubKeyEd25519       = errors.New("incompatible PubKeyEd25519")

	ErrInvalidPubKeyMultisigThreshold = errors.New("incompatible PubKeyMultisigThreshold")
)

// GetPubKeySecp256k1 returns a Tendermint secp256k1 public key from the
// PublicKey message type. It will return an error if the size of the public key
// is invalid or the underlying Pub field is invalid.
//
// NOTE: Do not use or call bytes on the result when serializing.
func (m *PublicKey) GetPubKeySecp256k1() (secp256k1.PubKeySecp256k1, error) {
	pk := secp256k1.PubKeySecp256k1{}

	if x, ok := m.GetPub().(*PublicKey_Secp256K1); ok {
		if len(x.Secp256K1) != secp256k1.PubKeySecp256k1Size {
			return pk, ErrInvalidPubKeySecp256k1Length
		}

		copy(pk[:], x.Secp256K1)
		return pk, nil
	}

	return pk, ErrInvalidPubKeySecp256k1
}

// GetPubKeyEd25519 returns a Tendermint Ed25519 public key from the PublicKey
// message type. It will return an error if the size of the public key
// is invalid or the underlying Pub field is invalid.
//
// NOTE: Do not use or call bytes on the result when serializing.
func (m *PublicKey) GetPubKeyEd25519() (ed25519.PubKeyEd25519, error) {
	pk := ed25519.PubKeyEd25519{}

	if x, ok := m.GetPub().(*PublicKey_Ed25519); ok {
		if len(x.Ed25519) != ed25519.PubKeyEd25519Size {
			return pk, ErrInvalidPubKeyEd25519Length
		}

		copy(pk[:], x.Ed25519)
		return pk, nil
	}

	return pk, ErrInvalidPubKeyEd25519
}

// GetPubKeyMultisigThreshold returns a Tendermint multi-sig threshold public key
// from the PublicKey message type. It will return an error if the size of the
// public key is invalid or the underlying Pub field is invalid.
//
// NOTE: Do not use or call bytes on the result when serializing.
func (m *PublicKey) GetPubKeyMultisigThreshold() (multisig.PubKeyMultisigThreshold, error) {
	mspk := multisig.PubKeyMultisigThreshold{}

	if x, ok := m.GetPub().(*PublicKey_Multisig); ok {
		mspk.K = uint(x.Multisig.K)
		mspk.PubKeys = make([]tmcrypto.PubKey, len(x.Multisig.Pubkeys))

		for i, pk := range x.Multisig.Pubkeys {
			mspk.PubKeys[i] = pk.TendermintPubKey()
		}

		return mspk, nil
	}

	return mspk, ErrInvalidPubKeyMultisigThreshold
}

// TendermintPubKey returns a Tendermint PubKey from the PublicKey proto message
// type.
func (m *PublicKey) TendermintPubKey() tmcrypto.PubKey {
	switch m.GetPub().(type) {
	case *PublicKey_Secp256K1:
		pk, err := m.GetPubKeySecp256k1()
		if err != nil {
			return nil
		}

		return pk

	case *PublicKey_Ed25519:
		pk, err := m.GetPubKeyEd25519()
		if err != nil {
			return nil
		}

		return pk

	case *PublicKey_Multisig:
		mspk, err := m.GetPubKeyMultisigThreshold()
		if err != nil {
			return nil
		}

		return mspk

	default:
		return nil
	}
}

// Address returns the address of a Tendermint PubKey.
func (m *PublicKey) Address() tmcrypto.Address {
	if pk := m.TendermintPubKey(); pk != nil {
		return pk.Address()
	}

	return nil
}
