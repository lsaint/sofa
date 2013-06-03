package game

import (
    "fmt"
    "net"
    "bufio"
    "encoding/binary"
    "time"

    "sofa/con"
)

const (
    ConnStateIn    = iota
    ConnStateDisc  = iota
)

type ClientConnection struct {
    conn        net.Conn
    reader      *bufio.Reader
    connState   int
    channel     chan []byte 
}

func NewClientConnection(c net.Conn) *ClientConnection {
    cliConn := new(ClientConnection)
    cliConn.conn = c
    cliConn.reader = bufio.NewReader(c)
    cliConn.channel = make(chan []byte, 10240)
    return cliConn
}

func (this *ClientConnection) send(buf []byte) {
    if this.connState == ConnStateDisc { return }

    head := make([]byte, 4)
    binary.LittleEndian.PutUint32(head, uint32(len(buf)))
    buf = append(head, buf...)

    select {                                                                                                          
        case this.channel <- buf:

        default:
    }
}

func (this *ClientConnection) blockSend(b []byte) bool {
    for len(b) > 0 {
        n, err := this.conn.Write(b)
        if err == nil {
            b = b[n:]
        } else if e, ok := err.(*net.OpError); ok && e.Temporary() {
            continue
        } else {
            if this.connState != ConnStateDisc {
                this.connState = ConnStateDisc
            }
            fmt.Println("blockSend disconnect")
            return false
        }
    }
    return true
}

func (this *ClientConnection) sendall() bool {
    for moreData := true; moreData; {
        select {
            case b := <-this.channel:
                if !this.blockSend(b) {
                    return false
                }
            default:
                moreData = false
        }
    }
    return true
}

func (this *ClientConnection) duplexRead(buff []byte) bool {
    var read_size int
    for {
        // write
        if !this.sendall() {
            return false
        }

        // read
        this.conn.SetReadDeadline(time.Now().Add(1e8))
        n, err := this.reader.Read(buff[read_size:])
        if err != nil {
            if e, ok := err.(*net.OpError); ok && e.Temporary() {
                read_size = n
                continue
            } else {
                fmt.Println("read err, disconnect", err)
                return false
            }
        }

        if n == 0 { return true }
        if n < len(buff) {
            read_size += n
            continue
        }
        return true
    }
    return false
}

func (this *ClientConnection) duplexReadBody() (ret []byte,  ok bool) {
    buff_head := make([]byte, 4)
    if !this.duplexRead(buff_head) {
        return
    }
    len_head := binary.LittleEndian.Uint32(buff_head)
    if len_head > con.MAX_LEN_HEAD {
        fmt.Println("message len too long", len_head)
        return
    }
    ret = make([]byte, len_head)
    if !this.duplexRead(ret) {
        return
    }
    ok = true
    return
}

