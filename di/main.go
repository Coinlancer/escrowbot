package di

import (
    "errors"
    "../abi"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/ethereum/go-ethereum/rpc"
)

type Container struct {
    EthClient     *ethclient.Client
    EthConnection *rpc.Client
    AbiParser     *abi.Container
    PrivateKey    []byte
}

var di *Container

func Get() *Container {
    if di == nil {
        panic("DI cotainer is not set")
    }

    return di
}

func Set(container *Container) error {
    if di != nil {
        return errors.New("DI container is already set")
    }

    di = container

    return nil
}