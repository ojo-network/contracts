package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/client"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/proto"

	ra "github.com/ojo-network/cw-relayer/relayer/dep/utils/remote_attestation"
	regtypes "github.com/ojo-network/cw-relayer/relayer/dep/utils/types"

	"github.com/miscreant/miscreant.go"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
)

// WASMContext wraps github.com/cosmos/cosmos-sdk/client/client.Context
type WASMContext struct {
	CLIContext       client.Context
	TestKeyPairPath  string
	TestMasterIOCert regtypes.MasterCertificate
}

type keyPair struct {
	Private string `json:"private"`
	Public  string `json:"public"`
}

// GetTxSenderKeyPair get the local tx encryption id
func (ctx WASMContext) GetTxSenderKeyPair() (privkey []byte, pubkey []byte, er error) {
	var keyPairFilePath string
	if len(ctx.TestKeyPairPath) > 0 {
		keyPairFilePath = ctx.TestKeyPairPath
	} else {
		keyPairFilePath = filepath.Join(ctx.CLIContext.HomeDir, "id_tx_io.json")
	}

	if _, err := os.Stat(keyPairFilePath); os.IsNotExist(err) {
		var privkey [32]byte
		rand.Read(privkey[:]) //nolint:errcheck

		var pubkey [32]byte
		curve25519.ScalarBaseMult(&pubkey, &privkey)

		keyPair := keyPair{
			Private: hex.EncodeToString(privkey[:]),
			Public:  hex.EncodeToString(pubkey[:]),
		}

		keyPairJSONBytes, err := json.MarshalIndent(keyPair, "", "    ")
		if err != nil {
			return nil, nil, err
		}

		err = os.WriteFile(keyPairFilePath, keyPairJSONBytes, 0o600)
		if err != nil {
			return nil, nil, err
		}

		return privkey[:], pubkey[:], nil
	}

	keyPairJSONBytes, err := os.ReadFile(keyPairFilePath)
	if err != nil {
		return nil, nil, err
	}

	var keyPair keyPair

	err = json.Unmarshal(keyPairJSONBytes, &keyPair)
	if err != nil {
		return nil, nil, err
	}

	privkey, err = hex.DecodeString(keyPair.Private)
	if err != nil {
		return nil, nil, err
	}
	pubkey, err = hex.DecodeString(keyPair.Public)
	if err != nil {
		return nil, nil, err
	}

	// TODO verify pubkey

	return privkey, pubkey, nil
}

var hkdfSalt = []byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x02, 0x4b, 0xea, 0xd8, 0xdf, 0x69, 0x99,
	0x08, 0x52, 0xc2, 0x02, 0xdb, 0x0e, 0x00, 0x97,
	0xc1, 0xa1, 0x2e, 0xa6, 0x37, 0xd7, 0xe9, 0x6d,
}

func GetConsensusIoPubKey(ctx WASMContext) ([]byte, error) {
	var masterIoKey regtypes.Key
	if ctx.TestMasterIOCert.Bytes != nil { // TODO check length?
		masterIoKey.Key = ctx.TestMasterIOCert.Bytes
	} else {
		res, _, err := ctx.CLIContext.Query("/secret.registration.v1beta1.Query/TxKey")
		if err != nil {
			return nil, err
		}

		err = encoding.GetCodec(proto.Name).Unmarshal(res, &masterIoKey)
		if err != nil {
			return nil, err
		}
	}

	ioPubkey, err := ra.UNSAFE_VerifyRaCert(masterIoKey.Key)
	if err != nil {
		return nil, err
	}

	return ioPubkey, nil
}

func (ctx WASMContext) getTxEncryptionKey(txSenderPrivKey, nonce, iopubkey []byte) ([]byte, error) {
	txEncryptionIkm, err := curve25519.X25519(txSenderPrivKey, iopubkey)
	if err != nil {
		fmt.Println("Failed to get tx encryption key")
		return nil, err
	}

	kdfFunc := hkdf.New(sha256.New, append(txEncryptionIkm, nonce...), hkdfSalt, []byte{})

	txEncryptionKey := make([]byte, 32)
	if _, err := io.ReadFull(kdfFunc, txEncryptionKey); err != nil {
		return nil, err
	}

	return txEncryptionKey, nil
}

// Encrypt encrypts
func (ctx WASMContext) Encrypt(iopubkey, plaintext []byte) ([]byte, error) {
	txSenderPrivKey, txSenderPubKey, err := ctx.GetTxSenderKeyPair()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	nonce := make([]byte, 32)
	_, err = rand.Read(nonce)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	txEncryptionKey, err := ctx.getTxEncryptionKey(txSenderPrivKey, nonce, iopubkey)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return encryptData(txEncryptionKey, txSenderPubKey, plaintext, nonce)
}

func encryptData(aesEncryptionKey []byte, txSenderPubKey []byte, plaintext []byte, nonce []byte) ([]byte, error) {
	cipher, err := miscreant.NewAESCMACSIV(aesEncryptionKey)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	ciphertext, err := cipher.Seal(nil, plaintext, []byte{})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// ciphertext = nonce(32) || wallet_pubkey(32) || ciphertext
	ciphertext = append(nonce, append(txSenderPubKey, ciphertext...)...) //nolint:gocritic

	return ciphertext, nil
}
