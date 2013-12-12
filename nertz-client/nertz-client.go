package main

import (
    websocket "code.google.com/p/go.net/websocket"
    "fmt"
    "log"
    "os"
    "strconv"
    "bufio"
    "strings"
)

type Move struct{
    card Card
    pile int
}

type Card struct{
    value int
    player string
}

type Arena struct {
    piles []Card
    moves chan Move
}


func credentials() (string, string) {
    reader := bufio.NewReader(os.Stdin)

    fmt.Print("Enter Username: ")
    username, _ := reader.ReadString('\n')

    fmt.Print("Enter Password: ")
    password, _ := reader.ReadString('\n')

    return strings.TrimSpace(username), strings.TrimSpace(password) // ReadString() leaves a trailing newline character
}

func reader(ws *websocket.Conn, ch chan string) {
    var buf string
    for {
        err := websocket.Message.Receive(ws, &buf)
        if err != nil {
            return
        }
        msg := fmt.Sprintf("Client got: %v\n", buf)
        ch <- msg
    }
}

func main() {

    if len(os.Args) != 3 {
        fmt.Fprintf(os.Stderr,"usage: %v <host> <port>\n", os.Args[0])
        return
    }

    host := os.Args[1]
    port, err := strconv.Atoi(os.Args[2])
    if err != nil {
        log.Fatal(err)
    }

    origin := "http://localhost/"
    url := fmt.Sprintf("ws://%v:%v/ws", host, port)

    println(url)


    ws, err := websocket.Dial(url, "", origin)
    if err != nil {
        log.Fatal(err)
    }
    defer ws.Close()

    player := NewPlayer(url, ws)

    fmt.Fprintf(os.Stdout, "Client connected to %v:%v...\n", host, port)

    ch := make(chan string, 50)

    go reader(ws, ch)
    for {
        select {
        case msg := <-ch:
            fmt.Printf(msg)
        default:
            reader := bufio.NewReader(os.Stdin)
            fmt.Print("Server gets: ")
            msg, err := reader.ReadString('\n')
            msg = strings.TrimSpace(msg)
            if err != nil {
                log.Fatal(err)
            }

            err = websocket.Message.Send(ws, msg)
            if err != nil {
                log.Fatal(err)
            }
        }
    }
}
