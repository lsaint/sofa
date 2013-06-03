package game

import (
    "fmt"

    pb "code.google.com/p/goprotobuf/proto"

    "sofa/proto"
)




//
type GameRoom struct {
    *Uid2Player
    Sid         uint32
    status      *proto.GameStatus
}

func NewGameRoom(sid uint32) *GameRoom {
    room := &GameRoom{
        Uid2Player      :   NewUid2Player(),
        Sid             :   sid,
    }
    go room.eventLoop()
    return room
}

func (this *GameRoom) eventLoop() {
}

func (this *GameRoom) ComeIn(player *Player) bool {
    this.AddPlayer(player.Uid, player)
    return true
}

func (this *GameRoom) Broadcast(msg ...pb.Message) {
    for _, player := range this.GetPlayers() {
        player.SendMsg(msg...)
    }
}

func (this *GameRoom) Leave(player *Player) {
    fmt.Println(player.Uid, "Leave")
    this.RmPlayer(player.Uid)
}

