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

type Credentials struct {
    Username string
    Password string
}

func (g *nertz.Game) WriteScores() {
    for scoreupdate := range <-g.ScoreChan {
        g.Done++
        g.Scores[scoreupdate.Player] -= scoreupdate.Value
        if g.Done == len(g.Clients) {
            close g.ScoreChan
        }
    }
}

func (c *nertz.Client) SendMessages() {
    ok := true
    for ok {
        select {
        case msg := <-c.Messages:
            jsonMsg := map[string]string{ "Message" : msg }
            err := websocket.JSON.Send(c.Conn, jsonMsg)
        case arena, ok := <-c.Arenas:
            if ok {
                err :=  websocket.JSON.Send(c.Conn, arena)
            } else {
                //if the channel is closed it means the game is over!
                jsonMsg := map[string]string{ "Message" : "Nertz" }
                err :=  websocket.JSON.Send(c.Conn, jsonMsg)
            }
        }
        if err != nil {
            panic("JSON.Send: " + err.Error())
        }
    }
}

func (c *nertz.Client) GetCredentials() *Credentials {
    c.Messages <- "Credentials"
    var creds Credentials
    err := websocket.JSON.Receive(c.Conn, &creds)
    if err != nil {
        log.Fatal(err)
    }
    return &creds
}

func MakeMoveHandler(g *Game) func(w http.ResponseWriter, r *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        var resp = make(map[string]bool)
        if !g.Done {
            data := new(Move)
            dec := json.Decoder(r.Body)
            dec.Decode(&data)
            resp["Ok"] = g.MakeMove(m *Move)
            w.Header().Set("Content-Type", "application/json")
            enc := json.Encoder(w)
            enc.Encode(resp)
            return
        } else {
            resp["Ok"] = false
            w.Header().Set("Content-Type", "application/json")
            enc := json.Encoder(w)
            enc.Encode(resp)
            return
        }
    }
}

func (g *nertz.Game) WaitForEnd(c *nertz.Client) {
    var hand nertz.Hand
    err := websocket.JSON.Receive(c.Conn, &hand)
    if err != nil {
        log.Fatal(err)
    }
    if !g.Done {
        if hand.IsNertz() {
            g.GameOver <- 1
        } else {
            //thanks for playing quitter!
        }
    }
    scoreupdate := map[string]int{ c.Name : hand.Nertzpile.Len() }
    g.ScoreChan <- scoreupdate
    <-g.ScoreChan
    err :=  websocket.JSON.Send(c.Conn, g.Scoreboard)
    if err != nil {
        panic("JSON.Send: " + err.Error())
    }
}

func MakeAcceptPlayers(g *nertz.Game) func(ws *websocket.Conn) {
    return func(ws *websocket.Conn) {
        if !g.Started {
            fmt.Fprintf(os.Stdout, "***Nertz server accepted a new player***\n")

            client = g.NewClient(ws)
            //get username :: client.Name = ??
            creds := c.GetCredentials()
            c.Name = creds.Username

            go client.SendMessages()
            g.WaitForEnd(client)
            }
        } else {
            err := websocket.JSON.Send(c.Conn, Request{"Sorry",)
            if err != nil {
                panic("JSON.Send: " + err.Error())
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

    arena := make([]*nertz.Piles, 0, 24)
    game := nertz.NewGame()
    game.Arenas <- arena
    go g.BroadcastMessages()
    go g.WriteScores()

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
