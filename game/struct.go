package game

import (
    "sync"

    "sofa/proto"
)


type Sid2GameRoom struct {
    sid2scene       map[uint32]*GameRoom        
    umLock          sync.RWMutex
}

func NewSid2GameRoom() *Sid2GameRoom {
    return &Sid2GameRoom{sid2scene:make(map[uint32]*GameRoom)}
}

func (this *Sid2GameRoom) AddGameRoom(sid uint32, scene *GameRoom) {
    this.umLock.Lock()
    defer this.umLock.Unlock()
    this.sid2scene[sid] = scene
}

func (this *Sid2GameRoom) RmGameRoom(sid uint32) {
    this.umLock.Lock()
    defer this.umLock.Unlock()
    delete(this.sid2scene, sid)
}

func (this *Sid2GameRoom) GetGameRoom(sid uint32) (u *GameRoom, ok bool) {
    this.umLock.RLock()
    defer this.umLock.RUnlock()
    u, ok = this.sid2scene[sid]
    return
}

func (this *Sid2GameRoom) Len() int {
    this.umLock.RLock()
    defer this.umLock.RUnlock()
    return len(this.sid2scene)
}

func (this *Sid2GameRoom) GetSids() (sids []uint32) {
    this.umLock.RLock()
    defer this.umLock.RUnlock()
    for _, scene := range this.sid2scene {
        sids = append(sids, scene.Sid)
    }
    return
}

func (this *Sid2GameRoom) GetGameRooms() (scenes []*GameRoom) {
    this.umLock.RLock()
    defer this.umLock.RUnlock()
    for _, u := range this.sid2scene {
        scenes = append(scenes, u)
    }
    return
}

//////////////////////////////////////////////////////////////////////////////////

type Uid2Player struct {
    uid2player      map[uint32]*Player        // uid:*player
    umLock          sync.RWMutex
}

func NewUid2Player() *Uid2Player {
    return &Uid2Player{uid2player:make(map[uint32]*Player)}
}

func (this *Uid2Player) AddPlayer(uid uint32, player *Player) {
    this.umLock.Lock()
    defer this.umLock.Unlock()
    this.uid2player[uid] = player
}

func (this *Uid2Player) RmPlayer(uid uint32) {
    this.umLock.Lock()
    defer this.umLock.Unlock()
    delete(this.uid2player, uid)
}

func (this *Uid2Player) GetPlayer(uid uint32) (u *Player, ok bool) {
    this.umLock.RLock()
    defer this.umLock.RUnlock()
    u, ok = this.uid2player[uid]
    return
}

func (this *Uid2Player) Len() int {
    this.umLock.RLock()
    defer this.umLock.RUnlock()
    return len(this.uid2player)
}

func (this *Uid2Player) GetUids() (uids []uint32) {
    this.umLock.RLock()
    defer this.umLock.RUnlock()
    for _, player := range this.uid2player {
        uids = append(uids, player.Uid)
    }
    return
}

func (this *Uid2Player) GetPlayers() (players []*Player) {
    this.umLock.RLock()
    defer this.umLock.RUnlock()
    for _, u := range this.uid2player {
        players = append(players, u)
    }
    return
}

//////////////////////////////////////////////////////////////////////

type Conn2Player struct {
    cliConn2Player      map[*ClientConnection]*Player
    cLock               sync.RWMutex
}

func NewConn2Player() *Conn2Player {
    return &Conn2Player{cliConn2Player : make(map[*ClientConnection]*Player)}
}

func (this *Conn2Player) AddPlayer(cliConn *ClientConnection, player *Player) {
    this.cLock.Lock()
    defer this.cLock.Unlock()
    this.cliConn2Player[cliConn] = player
}

func (this *Conn2Player) RmPlayer(cliConn *ClientConnection) {
    this.cLock.Lock()
    defer this.cLock.Unlock()
    delete(this.cliConn2Player, cliConn)
}

func (this *Conn2Player) GetPlayer(cliConn *ClientConnection) (u *Player, ok bool) {
    this.cLock.RLock()
    defer this.cLock.RUnlock()
    u, ok = this.cliConn2Player[cliConn]
    return
}

func (this *Conn2Player) Len() int {
    this.cLock.RLock()
    defer this.cLock.RUnlock()
    return len(this.cliConn2Player)
}

//////////////////////////////////////////////////////////////////////

type GameStatus struct {
    Status      proto.GameStatus
    sync.RWMutex
}

func NewGameStatus() *GameStatus {
    return &GameStatus{Status: proto.GameStatus_NotStarted}
}

func (this *GameStatus) SetStatus(st proto.GameStatus) {
    this.Lock()    
    defer this.Unlock()
    this.Status = st
}

func (this *GameStatus) GetStatus() proto.GameStatus {
    this.RLock()
    defer this.RUnlock()
    return this.Status
}

func (this *GameStatus) IsStarted() bool {
    this.RLock()
    defer this.RUnlock()
    return this.Status == proto.GameStatus_Started
}

//////////////////////////////////////////////////////////////////////

type Uid2Seat map[uint32]uint32

type Uid2Winner map[uint32]*proto.UserData

type Seat2UserData map[uint32]*proto.UserData

