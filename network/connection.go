package network

import (
    "io"
    "fmt"
    //"net"
    "bufio"
    "encoding/binary"

)

const (
    ConnStateIn    = iota
    ConnStateDisc  = iota

    MAX_LEN_HEAD   = 1024 * 4
)

type ClientConnection struct {
    rw          *bufio.ReadWriter
    connState   int
}

func NewClientConnection(rw io.ReadWriter) *ClientConnection {
    cliConn := new(ClientConnection)
    cliConn.rw = bufio.NewReadWriter(bufio.NewReader(rw), bufio.NewWriter(rw))
    return cliConn
}

func (this *ClientConnection) Send(buf []byte) {
    if this.connState == ConnStateDisc { return }
    fmt.Println("Send", len(buf))

    head := make([]byte, 4)
    binary.LittleEndian.PutUint32(head, uint32(len(buf)))
    buf = append(head, buf...)

    this.rw.Write(buf)
}

func (this *ClientConnection) sendall() bool {
    if err := this.rw.Flush(); err != nil {
        if this.connState != ConnStateDisc {
            this.connState = ConnStateDisc
        }
        fmt.Print("send err:", err)
        return false
    }
    return true
}

func (this *ClientConnection) duplexRead(buff []byte) bool {
    var read_size int
    for {
        // write
        if !this.sendall() {
            fmt.Println("send err")
            return false
        }

        // read
        n, err := this.rw.Read(buff[read_size:])
        if err != nil {
            if e, ok := err.(*RwError); ok && e.ErrNo == ERR_READ_TIMEOUT {
                read_size = n
                continue
            } else {
                fmt.Println("read err, disconnecting, reason:", err)
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
    if len_head > MAX_LEN_HEAD {
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

