// L'sofa

package main

import (
    "fmt"
    "os"
    "syscall"
    "os/signal"
    "runtime"
    "net/http"
    "log"
    _ "net/http/pprof"

    "sofa/game"
)


func handleSig() {
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)

    for sig := range c {
        fmt.Println("__handle__signal__", sig)
        return
    }
}


func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())

    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()

    authServer := game.AuthServer{}
    go authServer.Start()
    
    mgr := game.NewGameMgr()
    gs := game.NewGameServer(mgr)
    go gs.Start()

    handleSig()
}

