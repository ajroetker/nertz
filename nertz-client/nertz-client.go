package main

import (
    websocket "code.google.com/p/go.net/websocket"
    "fmt"
    "log"
    "os"
    "strconv"
    "bufio"
    "strings"
    "github.com/ajroetker/nertz"
)

func Credentials() (string, string) {
    reader := bufio.NewReader(os.Stdin)

    fmt.Print("Enter Username: ")
    username, _ := reader.ReadString('\n')

    fmt.Print("Enter Password: ")
    password, _ := reader.ReadString('\n')

    return strings.TrimSpace(username), strings.TrimSpace(password) // ReadString() leaves a trailing newline character
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

    name, password := Credentials()
    err = websocket.JSON.Send(ws, nertz.Credentials{ name, password, })
    if err != nil {
        panic("JSON.Send: " + err.Error())
    }
    player := nertz.NewPlayer(name, gameurl, ws)

    fmt.Fprintf(os.Stdout, "Client connected to %v:%v...\n", host, port)
    go player.HandleMessages()
    player.RecieveMessages()
}
