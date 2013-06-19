package game

import (
    "fmt"
    "reflect"

    //pb "code.google.com/p/goprotobuf/proto"

    "sofa/network"
    "sofa/proto"
)


type GameMgr struct {
     *Conn2Player    
     *Sid2GameRoom
}

func NewGameMgr() *GameMgr {
    return &GameMgr{
        Conn2Player     :   NewConn2Player(),
        Sid2GameRoom   :   NewSid2GameRoom(),
    }
}

func (this *GameMgr) Dispatch(cliConn *network.ClientConnection, request interface{}) {
    pname := reflect.ValueOf(request).Type().Elem().Name()[3:]
    //fmt.Println("onProto name", pname)
    player, ok := this.GetPlayer(cliConn)

    if pname == "Login" && !ok {
        this.OnLogin(cliConn, request)
        return
    }
    if !ok {
        fmt.Println("player not exist")
        return
    }

    if handle := reflect.ValueOf(this).MethodByName("On" + pname); handle.IsValid() {
        handle.Call([]reflect.Value{reflect.ValueOf(player), 
                    reflect.ValueOf(player.Room), reflect.ValueOf(request)})
        return
    }

    handle := reflect.ValueOf(player.Room).MethodByName("On" + pname)
    handle.Call([]reflect.Value{reflect.ValueOf(player), reflect.ValueOf(request)})
}

func (this *GameMgr) Disconnect(cliConn *network.ClientConnection) {
    if player, ok := this.GetPlayer(cliConn); ok {
        player.Room.Leave(player)
        this.RmPlayer(cliConn)
    }
}

func (this *GameMgr) OnLogin(cliConn *network.ClientConnection, request interface{}) {
    req := request.(*proto.C2SLogin)
    fmt.Println("onlogin", req)
    rep := &proto.S2CLoginRep{Ret: proto.Result_OK.Enum()}
    player := NewPlayer(cliConn, req.GetUid())
    sid := req.GetChannel()
    if sid == 0 { return }
    this.AddPlayer(cliConn, player)

    var room *GameRoom
    var ok bool
    if room, ok = this.GetGameRoom(sid); !ok {
        room = NewGameRoom(sid)
        this.AddGameRoom(sid, room)
    }
    player.UserData = req.UserInfo
    player.Room = room
    room.ComeIn(player)

    rep.Status = room.GetStatus().Enum()
    player.SendMsg(rep)
    fmt.Println("len:", this.Conn2Player.Len())
}

//func (this *GameMgr) OnStartGame(player *Player, room *GameRoom, request interface{}) {
//    req := request.(*proto.C2SStartGame)
//    room.Start(player, req)
//}

func (this *GameMgr) OnLogout(player *Player, room *GameRoom, request interface{}) {
    this.Disconnect(player.ClientConnection)
}

