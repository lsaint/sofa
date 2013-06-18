package game

import (
    "fmt"
    "reflect"
    "encoding/binary"

    pb "code.google.com/p/goprotobuf/proto"

    "sofa/proto"
    "sofa/network"
)




//
type Player struct {
    *network.ClientConnection
    *proto.UserData
    Uid         uint32
    Room       *GameRoom
}

func NewPlayer(cliConn *network.ClientConnection, uid uint32) *Player {
    return &Player{ ClientConnection:cliConn, Uid:uid}
}

func (this *Player) SendMsg(msgs ...pb.Message) {
    for _, msg := range msgs {
        if data, err := pb.Marshal(msg); err == nil {
            ty := reflect.ValueOf(msg).Elem().Type()
            if uri, ok := proto.PROTO2URI[ty]; ok {
                uri_field := make([]byte, 4)
                binary.LittleEndian.PutUint32(uri_field, uri)
                data = append(uri_field, data...)
                this.Send(data)
            } else {
                fmt.Println("SendMsg uri not exist", err, "type->", ty)
            }
        } else {
            fmt.Println("SendMsg Marshal err", err)
        }
    }
}



