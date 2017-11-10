package helpers

import (
    "math/big"
    "crypto/aes"
    "encoding/base64"
    "io"
    "crypto/cipher"
    "errors"
    "crypto/ecdsa"
    "crypto/rand"
    "github.com/ethereum/go-ethereum/crypto"
)

var WeiPrecision = big.NewFloat(0.000000000000000001)

func Wei2Float(wei *big.Int) string {
    f := big.NewFloat(0).SetInt(wei)
    output := f.Mul(f, WeiPrecision).String()

    return output
}

func Encrypt(key, text []byte) (string, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return "", err
    }

    b := base64.StdEncoding.EncodeToString(text)
    ciphertext := make([]byte, aes.BlockSize + len(b))
    iv := ciphertext[:aes.BlockSize]
    if _, err := io.ReadFull(rand.Reader, iv); err != nil {
        return "", err
    }

    cfb := cipher.NewCFBEncrypter(block, iv)
    cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))

    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func Decrypt(key []byte, textBase64 string) ([]byte, error) {
    text, err := base64.StdEncoding.DecodeString(textBase64)
    if err != nil {
        return nil, err
    }

    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    if len(text) < aes.BlockSize {
        return nil, errors.New("ciphertext too short")
    }

    iv := text[:aes.BlockSize]
    text = text[aes.BlockSize:]
    cfb := cipher.NewCFBDecrypter(block, iv)
    cfb.XORKeyStream(text, text)
    data, err := base64.StdEncoding.DecodeString(string(text))
    if err != nil {
        return nil, err
    }

    return data, nil
}

func PrivateKeyFromBytes(raw []byte) *ecdsa.PrivateKey {
    privateKeyECDSA := new(ecdsa.PrivateKey)
    privateKeyECDSA.PublicKey.Curve = crypto.S256()
    privateKeyECDSA.D = big.NewInt(32).SetBytes(raw)
    privateKeyECDSA.PublicKey.X, privateKeyECDSA.PublicKey.Y = crypto.S256().ScalarBaseMult(privateKeyECDSA.D.Bytes())

    return privateKeyECDSA
}