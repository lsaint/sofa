package game

import (
    "fmt"
    "sync"
    "strconv"

    pb "code.google.com/p/goprotobuf/proto"

    "sofa/proto"
    "sofa/con"
)

//
type Uid2Seat map[uint32]uint32

type Uid2Winner map[uint32]*proto.UserData

type Seat2UserData map[uint32]*proto.UserData

//
type GameRoom struct {
    *Uid2Player
    Uid2Seat
    Seat2UserData
    Uid2Winner
    Sid         uint32
    CurNum      uint32
    Status      proto.GameStatus
    sLock       sync.RWMutex       
    GameParam   *proto.C2SStartGame
    ChanTug     chan *Player
}

func NewGameRoom(sid uint32) *GameRoom {
    room := &GameRoom{
        Uid2Player      :   NewUid2Player(),
        Sid             :   sid,
    }
    room.reset()
    go room.eventLoop()
    return room
}

func (this *GameRoom) reset() {
    this.CurNum = 0
    this.Status = proto.GameStatus_NotStarted
    this.Uid2Seat = make(map[uint32]uint32)
    this.Seat2UserData = make(map[uint32]*proto.UserData)
    this.Uid2Winner = make(map[uint32]*proto.UserData)
    this.ChanTug = make(chan *Player)
}

func (this *GameRoom) eventLoop() {
    for {
        select {
            case player := <-this.ChanTug:
                if player != nil {
                    this.OnTug(player)
                }
        }
    }
}

func (this *GameRoom) ComeIn(player *Player) bool {
    this.AddPlayer(player.Uid, player)
    if seat, exist := this.Uid2Seat[player.Uid]; exist {
        player.SeatNum = pb.Uint32(seat)
        if _, exist := this.Uid2Winner[player.Uid]; exist {
            player.IsWin = pb.Bool(true)
        }
        player.SendMsg(this.PackTugMsg(player))
    }

    player.SendMsg(this.PackWinnersMsg())
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

func (this* GameRoom) CheckGameParam(param *proto.C2SStartGame) bool {
    switch this.GameParam.GetType() {
        case proto.C2SStartGame_SpeType:
            if len(this.GameParam.GetSpe().GetNumbers()) == 0 {
                return false
            }
        case proto.C2SStartGame_SecType:
            lower := this.GameParam.GetSec().GetLower()
            upper := this.GameParam.GetSec().GetUpper()
            if upper < lower || upper == 0 {
                return false
            }
    }
    return true
}

func (this *GameRoom) CheckAuth(player *Player) bool {
    return player.GetRole() > con.ROLE_LEVEL 
}

func (this *GameRoom) Start(player *Player, param *proto.C2SStartGame) {
    fmt.Println("START", param)
    this.sLock.Lock()
    defer this.sLock.Unlock()
    if this.Status == proto.GameStatus_Started || 
                !this.CheckGameParam(param) || !this.CheckAuth(player) {
        rep := &proto.S2CStartGameRep{Ret: proto.Result_FL.Enum()}
        player.SendMsg(rep)
        return
    }
    this.Status = proto.GameStatus_Started
    this.GameParam = param
    rep := &proto.S2CNotifyGameStart{UserInfo: &proto.UserData{Name: player.Name}}
    this.Broadcast(rep)
}

func (this *GameRoom) Stop(player *Player) {
    fmt.Println("STOP")
    this.sLock.Lock()
    defer this.sLock.Unlock()
    if this.Status != proto.GameStatus_Started {
        rep := &proto.S2CStopGameRep{Ret: proto.Result_FL.Enum()}
        player.SendMsg(rep)
        return
    }
    fmt.Println("stop sucess")
    rep := &proto.S2CNotifyGameStop{UserInfo: &proto.UserData{Name: player.Name}}
    this.Broadcast(rep)
    close(this.ChanTug)
    this.reset()
}

func (this *GameRoom) Tug(player *Player) {
    this.ChanTug <- player
}

func (this *GameRoom) OnTug(player *Player) {
    fmt.Println("room.OnTug")
    if this.Status != proto.GameStatus_Started {
        fmt.Println("tug: game not started")
        return
    }
    if this.Uid2Seat[player.Uid] == 0 {
        this.CurNum++
        player.SeatNum = pb.Uint32(this.CurNum)
        this.Uid2Seat[player.Uid] = this.CurNum
        this.Seat2UserData[this.CurNum] = player.UserData
    } else {
        return
    }
    fmt.Println(player.Uid, "Tug", player.GetSeatNum())
    this.Bingo(player)
    player.SendMsg(this.PackTugMsg(player))
    this.SendBesideTugMsg(player)
}

// seat list 
// |_| |_| ME |_|  |_|
func (this *GameRoom) PackTugMsg(player *Player) *proto.S2CTugRep {
    rep := new(proto.S2CTugRep)
    rep.SeatList = make([]*proto.UserData, 0, 5)
    sn, lower := player.GetSeatNum(), uint32(0)
    upper := sn + 2
    if sn > 2 {
        lower = sn - 2
    }
    for i:=lower; i<upper; i++ {
        if ud, exist := this.Seat2UserData[i]; exist {
            rep.SeatList = append(rep.SeatList, ud)
        }
    }

    return rep
}

func (this *GameRoom) SendBesideTugMsg(player *Player) {
    rep := new(proto.S2CTugRep)
    rep.SeatList = make([]*proto.UserData, 0, 1)
    sn := player.GetSeatNum()
    if sn == 1 { return }
    for i:=sn-1; i>0; i-- {
        ud := this.Seat2UserData[i]
        if p, exist := this.GetPlayer(ud.GetUid()); exist {
            rep.SeatList = append(rep.SeatList, player.UserData)
            p.SendMsg(rep)
        }
        if sn > 5 && (sn-i) > 2 {
            // 前5个位置都会把自己的信息发给前面的人
            // 5个之后的只发给自己前面两位
            return
        }
    }
}

func (this *GameRoom) PackWinnersMsg() *proto.S2CNotifyWinners {
    rep := new(proto.S2CNotifyWinners)
    rep.WinnerList = make([]*proto.UserData, 0, len(this.Uid2Winner))
    for _, user_info := range this.Uid2Winner {
        rep.WinnerList = append(rep.WinnerList, &proto.UserData{
                                                    SeatNum: user_info.SeatNum,
                                                    Name: user_info.Name,
                                                    Imid: user_info.Imid})
    }
    return rep
}

func (this *GameRoom) Bingo(player *Player) {
    setWinner := func(player *Player) {
        player.IsWin = pb.Bool(true)
        this.Uid2Winner[player.Uid] = player.UserData
        this.Broadcast(this.PackWinnersMsg())
        fmt.Println("win", player.Uid)
    }

    sn := player.GetSeatNum()
    player.IsWin = pb.Bool(false)
    switch this.GameParam.GetType() {
        case proto.C2SStartGame_NilType:
            fmt.Println("nil type")
        case proto.C2SStartGame_SpeType:
            fmt.Println("spe type")
            for _, n := range this.GameParam.GetSpe().GetNumbers() {
                if sn == n {
                    setWinner(player)
                }
            }
        case proto.C2SStartGame_SecType:
            fmt.Println("sec type")
            if sn >= this.GameParam.GetSec().GetLower() &&
               sn <= this.GameParam.GetSec().GetUpper() {
                setWinner(player)
            }
        case proto.C2SStartGame_SufType:
            suf := this.GameParam.GetSuf().GetSuffix()
            fmt.Println("suffix", suf)
            ssuf, ssn := strconv.Itoa(int(suf)), strconv.FormatInt(int64(sn), 10)
            lssuf, lssn := len(ssuf), len(ssn)
            if lssuf >= lssn && ssuf == ssn[lssuf-lssn:] {
                setWinner(player)
            }
    }
}

