package main

import (
    websocket "code.google.com/p/go.net/websocket"
    "fmt"
    "log"
    "os"
    "strconv"
    "bufio"
    "strings"
    "time"
    "github.com/ajroetker/nertz"
)

func Credentials() (string, string) {
    reader := bufio.NewReader(os.Stdin)

    fmt.Print("Enter Username: ")
    username, _ := reader.ReadString('\n')

    //fmt.Print("Enter Password: ")
    //password, _ := reader.ReadString('\n')

    //return strings.TrimSpace(username), strings.TrimSpace(password) // ReadString() leaves a trailing newline character
    return strings.TrimSpace(username), ""
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
    url := fmt.Sprintf("http://%v:%v", host, port)

    fmt.Fprintf(os.Stdout, "\nConnecting to the server at %v:%v...\n\n", host, port)
    ws, err := websocket.Dial(wsurl, "", origin)
    if err != nil {
        log.Fatal(err)
    }
    defer ws.Close()

    name, password := Credentials()
    player := nertz.NewPlayer(name, password, url, ws)

    go player.ReceiveMessages()
    time.Sleep(1000 * time.Millisecond)
    player.HandleMessages()
}
