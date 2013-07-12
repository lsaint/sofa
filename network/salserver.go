package network

/*
#cgo CFLAGS: -I /home/lsaint/go/src/gosal/open_sal/sdk
#cgo LDFLAGS: -L /home/lsaint/go/src/gosal/ -lopenyy_sal
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "sal-c.h"
#include "sal-events-c.h"
*/
import "C"

import (
    "fmt"
    "unsafe"
    "reflect"
    "time"
    "sync"
    "encoding/binary"

    pb "code.google.com/p/goprotobuf/proto"
    
    "sofa/proto"
)

const (
    ERR_MSG_TOO_LONG    = iota
    ERR_READ_TIMEOUT    = iota
    ERR_WRITE_0         = iota
)

type _IDispatcher interface {
    Dispatch(cliConn *ClientConnection, request interface{})
    Disconnect(cliConn *ClientConnection)
}

type SalServer struct {
    sal             *C.openyy_SAL_t
    WrChan          chan *WriteMsg
    Ucc             *UCC
    Ping            map[uint32]time.Duration
    Dispatcher      _IDispatcher
}

func NewSalServer(d _IDispatcher) *SalServer {
    sal := C.openyy_SAL_New(10088, C.CString("ssssssssssssssssssss"))
    if C.openyy_SAL_Init(sal) < 0 {
        panic("Init sal error")
    }
    return &SalServer{sal: sal, 
                WrChan: make(chan *WriteMsg, 10240),
                Ucc: NewUCC(),
                Dispatcher: d}
}

func (this *SalServer) Start() {
    if C.openyy_SAL_Start(this.sal) < 0 {
        panic("Start sal error")
    }
    fmt.Println("openyy_SAL_Start sucess")

    ev_chan := make(chan *C.openyy_SALEvent_t)
    done := make(chan bool)

    go this.waitEvent(ev_chan, done)
    this.handleEvent(ev_chan, done)
}

func (this *SalServer) waitEvent(c chan *C.openyy_SALEvent_t, d chan bool) {
    for {
        ev := &C.openyy_SALEvent_t{}
        if C.openyy_SAL_WaitEvent(this.sal, &ev, -1) < 0 {
            fmt.Println("Wait sal event error")
            return
        }
        c <- ev
        <-d
        C.openyy_SALEvent_Free(ev)
    }
}

func (this *SalServer) handleEvent(c chan *C.openyy_SALEvent_t, d chan bool) {
    for ev := range c {
        ev_type := C.openyy_SALEvent_GetType(ev) 
        switch ev_type {
            case C._SAL_LOG_EVENT_TP:
                //this.handleLog(ev)

            case C._SAL_SUBSCRIBE_HASH_CHANNEL_RES_EVENT_TP:
                this.handleSubscribeRep(ev)

            case C._SAL_LOGIN_RES_EVENT_TP:
                this.handleLoginRep(ev)

            case C._SAL_USER_MSG_EVENT_TP:
                this.handleUserMsg(ev)

            default:
                fmt.Println("Sal event:", ev_type)
        }
        d <- true
    }
}

func (this *SalServer) handleLog(ev *C.openyy_SALEvent_t) {
    var tm C.time_t
    var level_name, msg *C.char
    C.openyy_SALLogEvent_Datas(ev, &tm, &level_name, &msg)
    fmt.Println("[LOG]", tm, C.GoString(level_name), C.GoString(msg))
}

func (this *SalServer) handleSubscribeRep(ev *C.openyy_SALEvent_t) {
    var sub_chns, unsub_chns *C.uint
    var sub_count, unsub_count, min, max C.uint
    C.openyy_SALSubscribeHashChannelResEvent_Datas(ev, &sub_chns, &sub_count, 
        &unsub_chns, &unsub_count, &min, &max)
    fmt.Println("subscribe count:", sub_chns, sub_count, min, max)
    
    var theGoSlice []*C.uint
    sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&theGoSlice)))
    sliceHeader.Cap = int(sub_count)
    sliceHeader.Len = int(sub_count)
    sliceHeader.Data = uintptr(unsafe.Pointer(&sub_chns))
    fmt.Println(theGoSlice, *theGoSlice[0])
}

func (this *SalServer) handleLoginRep(ev *C.openyy_SALEvent_t) {
    var info *C.char
    C.openyy_SALLoginResEvent_Datas(ev, &info)
    fmt.Println("[LOGIN_REP]", C.GoString(info))
    //C.openyy_SAL_SubscribeHashChRange(this.sal, 0, 43670700, C._SAL_HASH_CH_RANGE_NONE);
    C.openyy_SAL_SubscribeHashChRange(this.sal, 0, 0, 43670710);

    go func() {
        for w := range this.WrChan {
            fmt.Println("openyy_SAL_SendMsgToUser", w.TopCh, w.Uid)
            msg := (*C.char)(unsafe.Pointer(&w.Msg[0]))
            C.openyy_SAL_SendMsgToUser(this.sal, w.TopCh, w.Uid, msg, C.uint(len(w.Msg)))
        }
    }()

    go func() {
        for {
            now := time.Now().Unix()
            this.Ucc.Lock()
            for u, cc := range this.Ucc.uid2cc {
                if now - cc.T > 60 {
                    delete(this.Ucc.uid2cc, u)    
                    this.Dispatcher.Disconnect(cc)
                    fmt.Println("delete", u)
                }
            }
            this.Ucc.Unlock()
            time.Sleep(60 * time.Second)
        }
    }()
}

func (this *SalServer) handleUserMsg(ev *C.openyy_SALEvent_t) {
    var top_ch, uid, msg_size C.uint
    var msg *C.char
    C.openyy_SALUserMsgEvent_Datas(ev, nil, &top_ch, &uid, &msg, &msg_size);
    fmt.Println("[USR_MSG]", top_ch, uid, msg_size)

    cc, ok := this.Ucc.Get(uint32(uid))
    if !ok {
        rw := &SalReadWriter{uid, top_ch, make(chan []byte, 512), this.WrChan}
        cc = NewClientConnection(rw)
        this.Ucc.Add(uint32(uid), cc)
        go this.acceptConn(cc)
    }
    cc.T = time.Now().Unix()
    if msg_size < 8 {return}

    b := C.GoBytes(unsafe.Pointer(msg), C.int(msg_size))
    select {
        case cc._rw.RdChan <- b:

        default:
            fmt.Println("user recive buff overflow")
    }

    //b := make([]byte, 4)
    //binary.LittleEndian.PutUint32(b, 3737)
    ////m := C.CString("lsaint")
    //m := (*C.char)(unsafe.Pointer(&b[0]))
    //C.openyy_SAL_BcMsgToTopCh(this.sal, 43670710, 0, m, 4)
}

func (this *SalServer) acceptConn(cc *ClientConnection) {
    for {
        if buff_body, ok := cc.duplexReadBody(); ok {
            this.parse(cc, buff_body)
            continue
        }
        //this.Dispatcher.Disconnect(cliConn)
        break
    }
}

func (this *SalServer) parse(cliConn *ClientConnection, msg []byte) {
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

///

type WriteMsg struct {
    Uid         C.uint
    TopCh       C.uint   
    Msg         []byte
}

///

type SalReadWriter struct {
    Uid         C.uint
    TopCh       C.uint
    RdChan      chan []byte
    WrChan      chan *WriteMsg
}

func (this *SalReadWriter) Read(p []byte) (n int, err error) {
    select {
        case b := <-this.RdChan:
            n := len(b)
            if n > len(p) {
                return 0, &RwError{ERR_MSG_TOO_LONG}
            }
            copy(p[0:n], b)
            return n, nil
        case <-time.After(200 * time.Millisecond):
            return 0, &RwError{ERR_READ_TIMEOUT}
    }
}

func (this *SalReadWriter) Write(p []byte) (n int, err error) {
    fmt.Println("rw.Write", len(p))
    n = len(p)
    if n == 0 {
        return 0, &RwError{ERR_WRITE_0}
    }
    //msg := (*C.char)(unsafe.Pointer(&p[0]))
    this.WrChan <- &WriteMsg{this.Uid, this.TopCh, p}
    return n, nil
}

///

type RwError struct {
    ErrNo   int             
}

func (this *RwError) Error() string {
    return fmt.Sprintf("Rw ErrNo:%v\n", this.ErrNo)
}

///

type UCC struct {
    uid2cc  map[uint32]*ClientConnection
    sync.RWMutex
}

func NewUCC() *UCC {
    return &UCC{uid2cc: make(map[uint32]*ClientConnection)}
}

func (this *UCC) Add(uid uint32, cc *ClientConnection) {
    this.Lock()
    defer this.Unlock()
    this.uid2cc[uid] = cc
}

func (this *UCC) Get(uid uint32) (cc *ClientConnection, ok bool){
    this.RLock()
    defer this.RUnlock()
    cc, ok = this.uid2cc[uid]
    return
}

func (this *UCC) Set(uid uint32) {
    this.Lock()
    defer this.Unlock()
    this.uid2cc[uid].T = time.Now().Unix()
}

