package web3

import (
    "github.com/ethereum/go-ethereum/rpc"
    "strconv"
    "crypto/ecdsa"
    "encoding/json"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/stellar/go/support/errors"
    "github.com/ethereum/go-ethereum/ethclient"
    "context"
    "github.com/ethereum/go-ethereum/rlp"
    "github.com/ethereum/go-ethereum/crypto"
    "math/big"
)

type tx struct {
    From        string
    To          string
    BlockHash   string
    BlockNumber string
}

func BlockNumber(c *rpc.Client) (uint64, error) {
    var raw interface{}
    err := c.Call(&raw, "eth_blockNumber")
    if err != nil {
        return 0, err
    }

    n, err := strconv.ParseUint(raw.(string), 0, 64)
    if err != nil {
        return 0, err
    }

    return uint64(n), nil
}

func GetTransaction(c *rpc.Client, hash common.Hash) (tx *tx, err error) {
    var raw json.RawMessage

    err = c.Call(&raw, "eth_getTransactionByHash", hash.String())
    if err != nil {
        return tx, err
    }

    if raw == nil {
        return tx, errors.New("Tx not found")
    }

    if err := json.Unmarshal(raw, &tx); err != nil {
        return tx, err
    }

    return tx, nil
}

func GetHighestBlocknum(c *rpc.Client, e *ethclient.Client) (uint64, error) {
    syncInfo, err := e.SyncProgress(context.TODO())
    if err != nil {
        return 0, err
    }

    // Check maybe sync is over
    if syncInfo != nil {
        return syncInfo.CurrentBlock, nil
    }

    blocknum, err := BlockNumber(c)
    if err != nil {
        return 0, err
    }

    return blocknum, nil
}

func SendTransaction(c *rpc.Client, e *ethclient.Client, to common.Address, amount *big.Int, gasLimit *big.Int, gasPrice *big.Int, privKey *ecdsa.PrivateKey, data ...[]byte) (string, error) {
    from := crypto.PubkeyToAddress(privKey.PublicKey)
    nonce, err := e.NonceAt(context.TODO(), from, nil)
    if err != nil {
        return "", err
    }

    signer := types.HomesteadSigner{}

    var inputData []byte
    if len(data) > 0 {
        inputData = data[0]
    }

    tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, inputData)
    signature, err := crypto.Sign(signer.Hash(tx).Bytes(), privKey)
    if err != nil {
        return "", err
    }

    txSigned, err := tx.WithSignature(signer, signature)
    if err != nil {
        return "", err
    }

    txBytes, err := rlp.EncodeToBytes(txSigned)
    if err != nil {
        return "", err
    }

    var raw interface{}
    err = c.Call(&raw, "eth_sendRawTransaction", common.ToHex(txBytes))
    if err != nil {
        return "", err
    }

    return string(raw.(string)), nil
}