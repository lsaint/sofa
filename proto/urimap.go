package proto

import (
    "reflect"
)

type Uri2Proto map[uint32]reflect.Type

var URI2PROTO Uri2Proto = Uri2Proto {
    1       :   reflect.TypeOf(C2SLogin{}),
    2       :   reflect.TypeOf(C2SStartGame{}),
    3       :   reflect.TypeOf(C2STug{}),
    4       :   reflect.TypeOf(C2SStopGame{}),

    1001    :   reflect.TypeOf(S2CLoginRep{}),
    1002    :   reflect.TypeOf(S2CStartGameRep{}),
    1003    :   reflect.TypeOf(S2CTugRep{}),
    1004    :   reflect.TypeOf(S2CStopGameRep{}),

    2001    :   reflect.TypeOf(S2CNotifyWinners{}),
    2002    :   reflect.TypeOf(S2CNotifyGameStart{}),
    2004    :   reflect.TypeOf(S2CNotifyGameStop{}),
}


type Proto2Uri map[reflect.Type]uint32
var PROTO2URI Proto2Uri

func init() {
    PROTO2URI = make(map[reflect.Type]uint32)
    for uri, proto_type := range URI2PROTO {
        PROTO2URI[proto_type] = uri
    }
}

