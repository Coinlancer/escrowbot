package actions

import (
    "net/http"
    "../response"
    "../response/errors"
    "../di"
    "../web3"
    "github.com/ethereum/go-ethereum/common"
    "github.com/sirupsen/logrus"
    "strconv"
)

const txHashLen = 66

func Confirmations(w http.ResponseWriter, r *http.Request) {
    tx := r.PostFormValue("tx")
    if len(tx) != txHashLen {
        response.JsonError(w, errors.ERR_BAD_PARAM, "tx")
        return
    }

    txHash := common.HexToHash(tx)
    diContainer := di.Get()

    w3tx, err := web3.GetTransaction(diContainer.EthConnection, txHash)
    if err != nil {
        logrus.Error(err)
        response.JsonError(w, errors.ERR_SERVICE)
        return
    }

    var confirmations int64
    if w3tx.BlockNumber != "" {
        n, err := strconv.ParseUint(w3tx.BlockNumber, 0, 64)
        if err != nil {
            logrus.Error(err)
            response.JsonError(w, errors.ERR_SERVICE)
            return
        }

        highestBlock, err := web3.GetHighestBlocknum(diContainer.EthConnection, diContainer.EthClient)
        if err != nil {
            logrus.Error(err)
            response.JsonError(w, errors.ERR_SERVICE)
            return
        }

        confirmations = int64(highestBlock) - int64(n)
    }

    response.Json(w, map[string]interface{}{
        "tx_hash": txHash.String(),
        "is_pending": w3tx.BlockNumber == "",
        "confirmations": confirmations,
    })
}
