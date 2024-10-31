package utils

import (
	"bytes"
	ecdsa2 "crypto/ecdsa"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/libsv/go-bk/bec"
)

func CreateMagicMessage(message string) string {
	const varIntProtoVer uint32 = 0
	// Signed message are prepended with this magicMessage
	// Taken from https://bitcoin.stackexchange.com/a/77325
	const magicMessage = "\x18Bitcoin Signed Message:\n"
	buffer := bytes.Buffer{}
	buffer.Grow(wire.VarIntSerializeSize(uint64(len(message))))
	// If we cannot write the VarInt, just panic since that should never happen
	if err := wire.WriteVarInt(&buffer, varIntProtoVer, uint64(len(message))); err != nil {
		panic(err)
	}
	return magicMessage + buffer.String() + message
}

func VerifySignature(signatureEncoded []byte, publicKey *btcec.PublicKey, messageHash []byte) error {
	if publicKey == nil || !publicKey.IsOnCurve() {
		return fmt.Errorf("public key was not correctly instantiated")
	}
	// Parse the signature so we can verify it
	parsedSignature, err := ParseCompact(signatureEncoded)
	if err != nil {
		return err
	}
	// Actually verify the message
	if verified := parsedSignature.Verify(messageHash, publicKey); !verified {
		return fmt.Errorf("signature could not be verified")
	}
	return nil
}

const (
	// compactSigSize is the size of a compact signature.  It consists of a
	// compact signature recovery code byte followed by the R and S components
	// serialized as 32-byte big-endian values. 1+32*2 = 65.
	// for the R and S components. 1+32+32=65.
	compactSigSize = 65
	// compactSigMagicOffset is a value used when creating the compact signature
	// recovery code inherited from Bitcoin and has no meaning, but has been
	// retained for compatibility.  For historical purposes, it was originally
	// picked to avoid a binary representation that would allow compact
	// signatures to be mistaken for other components.
	compactSigMagicOffset = 27
	// compactSigCompPubKey is a value used when creating the compact signature
	// recovery code to indicate the original public key was compressed.
	compactSigCompPubKey = 4
)

func ParseCompact(signature []byte) (*ecdsa.Signature, error) {
	if len(signature) != compactSigSize {
		return nil, fmt.Errorf("invalid compact signature size")
	}
	// Parse and validate the compact signature recovery code.
	const (
		minValidCode = compactSigMagicOffset
		maxValidCode = compactSigMagicOffset + compactSigCompPubKey + 3
	)
	if signature[0] < minValidCode || signature[0] > maxValidCode {
		return nil, fmt.Errorf("invalid compact signature recovery code")
	}
	// Parse and validate the R and S signature components.
	//
	// Fail if r and s are not in [1, N-1].
	var r, s btcec.ModNScalar
	if overflow := r.SetByteSlice(signature[1:33]); overflow {
		return nil, fmt.Errorf("signature R is >= curve order")
	}
	if r.IsZero() {
		return nil, fmt.Errorf("signature R is 0")
	}
	if overflow := s.SetByteSlice(signature[33:]); overflow {
		return nil, fmt.Errorf("signature S is >= curve order")
	}
	if s.IsZero() {
		return nil, fmt.Errorf("ignature S is 0")
	}
	return ecdsa.NewSignature(&r, &s), nil
}

func SignMessageAlter(privateKey, msg string) ([]byte, []byte) {
	// Decode a hex-encoded private key.
	pkBytes, err := hex.DecodeString(privateKey)
	if err != nil {
		fmt.Printf("DecodeString privateKey failed, err=%v", err)
		return nil, nil
	}
	message := msg
	//message := CreateMagicMessage(msg)
	messageHash := chainhash.DoubleHashB([]byte(message))
	//signature := ecdsa.Sign(priKey, messageHash)
	priKey, pubKeys := btcec.PrivKeyFromBytes(pkBytes)
	//signature, err := priKey.Sign(messageHash)
	//if err != nil {
	//  fmt.Printf("priKey sign failed, err=%v", err)
	//  return nil, nil
	//}
	signature := ecdsa.Sign(priKey, messageHash)
	fmt.Printf("btcec pub: %v\n", hex.EncodeToString(pubKeys.SerializeCompressed()))
	fmt.Printf("btcec sig: %v\n", hex.EncodeToString(signature.Serialize()))
	//fmt.Printf("btcec sig base64: %v\n", base64.StdEncoding.EncodeToString(signature.Serialize()))
	//fmt.Printf("btcec sig len: %v\n", len(base64.StdEncoding.EncodeToString(signature.Serialize())))
	x, y := bec.S256().ScalarBaseMult(pkBytes)
	ecdsaPubKey := ecdsa2.PublicKey{
		Curve: bec.S256(),
		X:     x,
		Y:     y,
	}
	ecdsaPrivateKey := &bec.PrivateKey{PublicKey: ecdsaPubKey, D: new(big.Int).SetBytes(pkBytes)}
	sigBytes, err := bec.SignCompact(bec.S256(), ecdsaPrivateKey, messageHash, true)
	pubKey, _, err := bec.RecoverCompact(bec.S256(), sigBytes, messageHash)
	// Verify the signature for the message using the public key.
	//verified := signature.Verify(messageHash, pubKey)
	return sigBytes, pubKey.SerialiseCompressed()
}

var ErrPrivateKeyMissing = errors.New("private key is missing")

func SignMessage(privateKey string, message string, sigRefCompressedKey bool) (string, string, error) {
	if len(privateKey) == 0 {
		return "", "", ErrPrivateKeyMissing
	}
	var buf bytes.Buffer
	var err error
	hBSV := "Bitcoin Signed Message:\n"
	if err = wire.WriteVarString(&buf, 0, hBSV); err != nil {
		return "", "", err
	}
	if err = wire.WriteVarString(&buf, 0, message); err != nil {
		return "", "", err
	}
	// Create the hash
	messageHash := chainhash.DoubleHashB(buf.Bytes())
	// Get the private key
	var ecdsaPrivateKey *bec.PrivateKey
	if ecdsaPrivateKey, err = PrivateKeyFromString(privateKey); err != nil {
		return "", "", err
	}
	// Sign
	var sigBytes []byte
	if sigBytes, err = bec.SignCompact(bec.S256(), ecdsaPrivateKey, messageHash, sigRefCompressedKey); err != nil {
		return "", "", err
	}
	pubKey, _, err := bec.RecoverCompact(bec.S256(), sigBytes, messageHash)
	if err != nil {
		return "", "", err
	}
	return base64.StdEncoding.EncodeToString(sigBytes), hex.EncodeToString(pubKey.SerialiseCompressed()), nil
}
func PrivateKeyFromString(privateKey string) (*bec.PrivateKey, error) {
	if len(privateKey) == 0 {
		return nil, ErrPrivateKeyMissing
	}
	privateKeyBytes, err := hex.DecodeString(privateKey)
	if err != nil {
		return nil, err
	}
	x, y := bec.S256().ScalarBaseMult(privateKeyBytes)
	ecdsaPubKey := ecdsa2.PublicKey{
		Curve: bec.S256(),
		X:     x,
		Y:     y,
	}
	return &bec.PrivateKey{PublicKey: ecdsaPubKey, D: new(big.Int).SetBytes(privateKeyBytes)}, nil
}
