package network

import (
    "fmt"
    "net"
    "reflect"
    "encoding/binary"

    pb "code.google.com/p/goprotobuf/proto"
    
    "sofa/proto"
)

func init() {
}

type IDispatcher interface {
    Dispatch(cliConn *ClientConnection, request interface{})
    Disconnect(cliConn *ClientConnection)
}

type GameServer struct {
    Dispatcher     IDispatcher
}


func NewGameServer(d IDispatcher) *GameServer {
    return &GameServer{
        Dispatcher : d,
    }
}


func (this *GameServer) Start() {
    ln, err := net.Listen("tcp", ":13603")                                                                            
    if err != nil {
        fmt.Println("err", err)
    }
    fmt.Println("gameServer runing")
    for {
        conn, err := ln.Accept()
        if err != nil {
            fmt.Println("Accept error", err)
            continue
        }
        go this.acceptConn(conn)
    }
}

func (this *GameServer) acceptConn(conn net.Conn) {
    cliConn := NewClientConnection(conn)
    for {
        if buff_body, ok := cliConn.duplexReadBody(); ok {
            this.parse(cliConn, buff_body)
            continue
        }
        this.Dispatcher.Disconnect(cliConn)
        break
    }
    conn.Close()
}

func (this *GameServer) parse(cliConn *ClientConnection, msg []byte) {
    //defer func() {
    //    if r := recover(); r != nil {
    //        fmt.Println("parse err", r)
    //    }
    //}()

    uri := binary.LittleEndian.Uint32(msg[:4])

    ty := proto.URI2PROTO[uri]
    new_ins_value := reflect.New(ty)
    err := pb.Unmarshal(msg[4:], new_ins_value.Interface().(pb.Message))
    if err != nil {
        fmt.Println("pb Unmarshal", err)
        return
    }
    this.Dispatcher.Dispatch(cliConn, new_ins_value.Interface())
}

