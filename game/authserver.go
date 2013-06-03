package game

import (
    "net"
    "fmt"

    "sofa/con"
)

type AuthServer struct {
}

func (aus *AuthServer) Start() {
    ln, err := net.Listen("tcp", ":13604")
    if err != nil {
        fmt.Println("authServer err")
    }
    fmt.Println("authServer runing")
    for {
        conn, err := ln.Accept()
        if err != nil {
            continue
        }
        go func(c net.Conn) {
            c.Write([]byte(con.XML_REP))
            c.Write([]byte("\x00"))
        }(conn)
    }
}
