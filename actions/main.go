package actions

import (
    "net/http"
    "../conf"
    "../di"
    "../response"
    "../helpers"
    "../web3"
    "../response/errors"
    "math/big"
    goErrors "errors"
    "github.com/ethereum/go-ethereum"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/common"
    "github.com/sirupsen/logrus"
    "strconv"
    "context"
    "encoding/hex"
    "fmt"
)

var errGasLimitExceeded = goErrors.New("Gas limit too big")

func Deposit(w http.ResponseWriter, r *http.Request) {
    step, err := strconv.ParseInt(r.PostFormValue("step"), 10, 64)
    if err != nil || step <= 0 {
        response.JsonError(w, errors.ERR_BAD_PARAM, "step")
        return
    }

    amount, ok := big.NewInt(0).SetString(r.PostFormValue("amount"), 10)
    if !ok || amount.Uint64() <= 100 { // Although this amount is not correct, as uint64 is too small, it's enough for this check
        response.JsonError(w, errors.ERR_BAD_PARAM, "amount")
        return
    }

    if (!common.IsHexAddress(r.PostFormValue("from"))) {
        response.JsonError(w, errors.ERR_BAD_PARAM, "from")
        return
    }

    if (!common.IsHexAddress(r.PostFormValue("to"))) {
        response.JsonError(w, errors.ERR_BAD_PARAM, "to")
        return
    }

    from := common.HexToAddress(r.PostFormValue("from"))
    to := common.HexToAddress(r.PostFormValue("to"))

    diContainer := di.Get()

    packedData, err := diContainer.AbiParser.Abi.Pack("deposit", big.NewInt(step), from, to, amount)
    if err != nil {
        logrus.Warningf("Cannot pack contractData: %s", err.Error())
        response.JsonError(w, errors.ERR_SERVICE)
        return
    }

    txHash, err := makeCall(packedData)
    if err != nil {
        handleError(w, err)
        return
    }

    response.Json(w, map[string]interface{}{
        "tx_hash": txHash,
    })
}

func Pay(w http.ResponseWriter, r *http.Request) {
    step, err := strconv.ParseInt(r.PostFormValue("step"), 10, 64)
    if err != nil || step <= 0 {
        response.JsonError(w, errors.ERR_BAD_PARAM, "step")
        return
    }

    diContainer := di.Get()
    packedData, err := diContainer.AbiParser.Abi.Pack("pay", big.NewInt(step))
    if err != nil {
        logrus.Warningf("Cannot pack contractData: %s", err.Error())
        response.JsonError(w, errors.ERR_SERVICE)
        return
    }

    txHash, err := makeCall(packedData)
    if err != nil {
        handleError(w, err)
        return
    }

    response.Json(w, map[string]interface{}{
        "tx_hash": txHash,
    })
}

func Refund(w http.ResponseWriter, r *http.Request) {
    step, err := strconv.ParseInt(r.PostFormValue("step"), 10, 64)
    if err != nil || step <= 0 {
        response.JsonError(w, errors.ERR_BAD_PARAM, "step")
        return
    }

    diContainer := di.Get()
    packedData, err := diContainer.AbiParser.Abi.Pack("refund", big.NewInt(step))
    if err != nil {
        logrus.Warningf("Cannot pack contractData: %s", err.Error())
        response.JsonError(w, errors.ERR_SERVICE)
        return
    }

    txHash, err := makeCall(packedData)
    if err != nil {
        handleError(w, err)
        return
    }

    response.Json(w, map[string]interface{}{
        "tx_hash": txHash,
    })
}

func makeCall(data []byte) (string, error) {
    diContainer := di.Get()
    contractAddress := common.HexToAddress(conf.EscrowContractAddress)

    rawPrivateKey, err := hex.DecodeString(string(diContainer.PrivateKey))
    if err != nil {
        return "", err
    }

    privateKey := helpers.PrivateKeyFromBytes(rawPrivateKey)
    msg := ethereum.CallMsg{
        From: crypto.PubkeyToAddress(privateKey.PublicKey),
        To: &contractAddress,
        Data: data,
    }

    gasPrice, err := diContainer.EthClient.SuggestGasPrice(context.TODO())
    if err != nil {
        return "", err
    }

    estimatedGasLimit, err := diContainer.EthClient.EstimateGas(context.TODO(), msg)
    if err != nil {
        return "", err
    }

    if estimatedGasLimit.Int64() > int64(conf.MaxContractGasLimit) {
        return "", errGasLimitExceeded
    }

    logrus.Infof("Preparing contract call. GasPrice: %s. GasLimit: %s", gasPrice.String(), estimatedGasLimit.String())

    txHash, err := web3.SendTransaction(
        diContainer.EthConnection,
        diContainer.EthClient,
        contractAddress,
        big.NewInt(0),
        estimatedGasLimit,
        gasPrice,
        privateKey,
        data,
    )

    if err != nil {
        return "", err
    }

    return txHash, nil
}

func handleError(w http.ResponseWriter, err error) {
    logrus.Error(err.Error())

    switch err {
    case errGasLimitExceeded:
        response.JsonError(w, errors.ERR_GAS_LIMIT_EXCEEDED, fmt.Sprintf("Gas required exceeds limit: %d. Probably some contract conditions are unmet.", conf.MaxContractGasLimit))
    default:
        response.JsonError(w, errors.ERR_SERVICE)
    }
}