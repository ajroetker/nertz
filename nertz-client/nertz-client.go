package main

import (
    websocket "code.google.com/p/go.net/websocket"
    "fmt"
    "log"
    "os"
    "strconv"
    "bufio"
    "strings"
    "nertz"
)

func credentials() (string, string) {
    reader := bufio.NewReader(os.Stdin)

    fmt.Print("Enter Username: ")
    username, _ := reader.ReadString('\n')

    fmt.Print("Enter Password: ")
    password, _ := reader.ReadString('\n')

    return strings.TrimSpace(username), strings.TrimSpace(password) // ReadString() leaves a trailing newline character
}

func PrintCardStack(cs *list.List, toShow int) {
    stack := "[%v"
    for e := cs.Front() ; e != nil ; e = e.Next() {
        if toShow > 0 {
            card := fmt.Sprintf("%v%v]%%v", e.Value.Value, e.Value.Suit))     
            stack = fmt.Sprintf(stack, card)
            toShow--
        } else {
            stack = fmt.Sprintf(stack, "]%v")
        }   
    }
    stack = fmt.Sprintf(stack, "")
    if stack == "[" {
       fmt.Println("empty stack")
    } else {
        fmt.Println(stack)
    }
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
    wsurl := fmt.Sprintf("ws://%v:%v/ws", host, port)
    gameurl := fmt.Sprintf("http://%v:%v/move", host, port)

    ws, err := websocket.Dial(wsurl, "", origin)
    if err != nil {
        log.Fatal(err)
    }
    defer ws.Close()

    name, _ := credentials()
    player :=  NewPlayer(name, gameurl, ws) *Player {

    fmt.Fprintf(os.Stdout, "Client connected to %v:%v...\n", host, port)

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
