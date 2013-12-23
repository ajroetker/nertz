package main

import (
    websocket "code.google.com/p/go.net/websocket"
    "fmt"
    "log"
    "os"
    "strconv"
    "net/http"
    "encoding/json"
    "github.com/ajroetker/nertz"
)

func MakeReadyHandler(g *nertz.Game) func(w http.ResponseWriter, r *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        var resp = make(map[string]string)
        if ! g.Started {
            resp["Message"] = "Waiting on the other players..."
            w.Header().Set("Content-Type", "application/json")
            enc := json.NewEncoder(w)
            enc.Encode(resp)

            val := <-g.ReadyPlayers
            if val == len(g.Clients) - 1 {
                g.Begin <- 1
            }
            g.ReadyPlayers <- val + 1
            return
        } else {
            resp["Message"] = "Already Started!"
            w.Header().Set("Content-Type", "application/json")
            enc := json.NewEncoder(w)
            enc.Encode(resp)
            return
        }
    }
}

func MakeMoveHandler(g *nertz.Game) func(w http.ResponseWriter, r *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        var resp = make(map[string]bool)
        if ! g.Started {
            data := new(nertz.Move)
            dec := json.NewDecoder(r.Body)
            dec.Decode(&data)
            //TODO Document this in the README.md
            resp["Ok"] = g.MakeMove(data)
            w.Header().Set("Content-Type", "application/json")
            enc := json.NewEncoder(w)
            enc.Encode(resp)
            return
        } else {
            resp["Ok"] = false
            w.Header().Set("Content-Type", "application/json")
            enc := json.NewEncoder(w)
            enc.Encode(resp)
            return
        }
    }
}

func MakeAcceptPlayers(g *nertz.Game) func(ws *websocket.Conn) {
    return func(ws *websocket.Conn) {
        if ! g.Started {
            fmt.Fprintf(os.Stdout, "***Nertz server accepted a new player***\n")

            client := g.NewClient(ws)
            creds := client.GetCredentials()
            client.Name = creds.Username

            g.WaitForStart(client)
            go client.SendMessages()
            g.WaitForEnd(client)
        } else {
            jsonMsg := map[string]string{ "Message" : "In Progress" }
            err :=  websocket.JSON.Send(ws, jsonMsg)
            if err != nil {
                panic("JSON.Send: " + err.Error())
            }
        }
    }
}

func main() {
    if len(os.Args) != 2 {
        fmt.Fprintf(os.Stderr, "usage: %s <port>\n", os.Args[0])
        return
    }

    port, err := strconv.Atoi(os.Args[1]);
    if err != nil {
        log.Fatal(err)
    }
    listenAt := fmt.Sprintf(":%v", port)

    game := nertz.NewGame(6)
    go game.BroadcastMessages()
    go game.WriteScores()

    GameHandler  := MakeAcceptPlayers(game)
    MoveHandler  := MakeMoveHandler(game)
    ReadyHandler := MakeReadyHandler(game)

    http.Handle("/", websocket.Handler(GameHandler))
    http.Handle("/move", http.HandlerFunc(MoveHandler))
    http.Handle("/ready", http.HandlerFunc(ReadyHandler))
    fmt.Fprintf(os.Stdout, "Nertz server listening on port %v...\n", port)
    err = http.ListenAndServe(listenAt, nil)
    if err != nil {
        panic("ListenAndServe: " + err.Error())
    }
}
