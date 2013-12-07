package main

import (
    websocket "code.google.com/p/go.net/websocket"
    "fmt"
    "log"
    "os"
    "strconv"
    "net/http"
    "json"
    "github.com/ajroetker/nertz"
)

func MakeMoveHandler(g *Game) func(w http.ResponseWriter, r *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        if !g.Done {
            data := new(Move)
            dec := json.Decoder(r.Body)
            dec.Decode(&data)
            resp := g.MakeMove(m *Move)
            w.Header().Set("Content-Type", "application/json")
            enc := json.Encoder(w)
            enc.Encode(Response{resp,})
            return
        } else {
            w.Header().Set("Content-Type", "application/json")
            enc := json.Encoder(w)
            enc.Encode(Response{false,})
            return
        }
    }
}

func WriteArena(c *nertz.Client) {
    for arena := range c.Arenas {
        err :=  websocket.JSON.Send(c.Conn, arena)
        if err != nil {
            panic("JSON.Send: " + err.Error())
        }
    }
}

func WaitForEnd(c *nertz.Client) {
    var hand nertz.Hand
    err := websocket.JSON.Receive(c.Conn, &hand)
    if err != nil {
        log.Fatal(err)
    }
    if hand.IsNertz() {

    } else {
        
    }
}

func MakeAcceptPlayers(g *nertz.Game) func(ws *websocket.Conn) {
    return func(ws *websocket.Conn) {
        fmt.Fprintf(os.Stdout, "***Nertz server accepted a new player***\n")

        client = g.NewClient(ws)
        //get username :: client.Name = ??

        go WriteArena(client)
        WaitForEnd(client)
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

    arena := make([]*nertz.Piles, 0, 24)
    game := nertz.NewGame()
    game.Arenas <- arena
    go g.BroadcastMessages()

    GameHandler := MakeAcceptPlayers(game)
    MoveHandler := nertz.MakeMoveHandler(game)

    http.Handle("/", websocket.Handler(GameHandler))
    http.Handle("/move", http.Handler(MoveHandler))
    fmt.Fprintf(os.Stdout, "Nertz server listening on port %v...\n", port)
    err = http.ListenAndServe(listenAt, nil)
    if err != nil {
        panic("ListenAndServe: " + err.Error())
    }
}
