package main

import (
    "./conf"
    "./abi"
    "./di"
    "./actions"
    "./helpers"
    "./response"
    "./response/errors"
    "flag"
    "net/http"
    "os"
    "fmt"
    "syscall"
    "crypto/md5"
    "encoding/hex"
    "golang.org/x/crypto/ssh/terminal"
    "github.com/urfave/negroni"
    "github.com/sirupsen/logrus"
    "github.com/ethereum/go-ethereum/rpc"
    "github.com/ethereum/go-ethereum/ethclient"
    //"github.com/rs/cors"
)

const pathPrefix = "/api"

func main() {
    logrus.SetOutput(os.Stdout)
    logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})
    logrus.SetLevel(logrus.DebugLevel)

    boolPtr := flag.Bool("encode-key", false, "encode private key")
    flag.Parse()

    if *boolPtr {
        encodeKey()
    } else {
        start()
    }
}

func start() {
    conn, err := rpc.Dial(conf.EthHost)
    if err != nil {
        panic("Failed to connect to Ethereum node: " + err.Error())
    }

    ec := ethclient.NewClient(conn)

    abiParser, err := abi.New()
    if err != nil {
        panic("Cannot init ABI parser: " + err.Error())
    }

    fmt.Printf("Enter Password:")
    bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
    if err != nil {
        panic(err.Error())
    }
    fmt.Println("")

    pwdHash := md5.Sum(bytePassword)
    rawPrivateKey, err := helpers.Decrypt(pwdHash[:], conf.OwnerPrivKeyEncrypted)
    if err != nil {
        panic("Cannot decode private key.")
    }

    // Set DI container
    diContainer := &di.Container{
        EthConnection: conn,
        EthClient:     ec,
        AbiParser:     abiParser,
        PrivateKey:    rawPrivateKey,
    }

    err = di.Set(diContainer)
    if err != nil {
        logrus.Fatalf("Error setting DI: %s", err.Error())
    }

    router := http.NewServeMux()
    router.HandleFunc("/deposit", actions.Deposit)
    router.HandleFunc("/refund", actions.Refund)
    router.HandleFunc("/pay", actions.Pay)
    router.HandleFunc("/confirmations", actions.Confirmations)

    //corsMiddleware := cors.New(cors.Options{
    //    AllowedOrigins: []string{"*"},
    //    AllowedMethods: []string{"POST", "GET", "OPTIONS", "PUT", "DELETE"},
    //    AllowedHeaders: []string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"},
    //})

    n := negroni.New(negroni.HandlerFunc(middlewareAuth))
    //n.Use(corsMiddleware)
    n.UseHandler(router)

    logrus.Info("Starting server on: " + conf.HttpPort)
    http.ListenAndServe(":" + conf.HttpPort, n)
}

func encodeKey() {
    fmt.Printf("Enter Private Key:")
    privKey, err := terminal.ReadPassword(int(syscall.Stdin))
    fmt.Println("")
    if err != nil {
        panic(err.Error())
    }

    fmt.Printf("Enter Password:")
    pwd, err := terminal.ReadPassword(int(syscall.Stdin))
    pwdHash := md5.Sum(pwd)
    fmt.Println("")
    if err != nil {
        panic(err.Error())
    }

    encrypted, err := helpers.Encrypt(pwdHash[:], privKey)
    if err != nil {
        panic(err.Error())
    }

    fmt.Println(encrypted)
}

func middlewareAuth(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
    user, pwd, ok := r.BasicAuth()

    hasher := md5.New()
    hasher.Write([]byte(user + pwd))

    authHash := hex.EncodeToString(hasher.Sum(nil))
    if !ok || authHash != conf.BasicAuthHash {
        response.JsonError(w, errors.ERR_NOT_ALLOWED)
        return
    }

    next(w, r)
}