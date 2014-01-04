package nertz

import (
    "fmt"
    "os"
    "log"
    "time"
    websocket "code.google.com/p/go.net/websocket"
)

/* Server-side code */

func (c *Client) GetCredentials() *Credentials {
    c.Messages <- "Credentials"
    var creds Credentials
    err := websocket.JSON.Receive(c.Conn, &creds)
    if err != nil {
        log.Fatal(err)
    }
    return &creds
}

func (g *Game) AddNewClients() {
    for cli := range g.NewClients {
        g.Clients = append(g.Clients, cli)
    }
}

func (g *Game) BroadcastMessages() {
    for {
        select {
        case <-g.Begin:
            g.Started = true
            g.Init(len(g.Clients))
            time.Sleep(100 * time.Millisecond)
            fmt.Fprintf( os.Stdout, "--------------------------------------------\n             Let the game begin!            \n--------------------------------------------\n")
            for _, c := range g.Clients {
                if ! c.Done {
                    c.Messages <- "Let's Begin!"
                }
            }
        case lake := <-g.Updates:
            for _, c := range g.Clients {
                if ! c.Done {
                    c.Lakes <- lake
                }
            }
        case <-g.GameOver:
            g.TallyUp()
            close(g.Lakes)
            fmt.Fprintf(os.Stdout, "--------------------------------------------\n            Looks like nertz!\n--------------------------------------------\n")
            for _, c := range g.Clients {
                if ! c.Done {
                    close(c.Lakes)
                }
            }
            return
        }
    }
}

func (c *Client) SendMessages() {
    ok := true
    var lake Lake
    var err error
    for ok {
        select {
        case msg := <-c.Messages:
            jsonMsg := map[string]string{ "Message" : msg }
            err = websocket.JSON.Send(c.Conn, jsonMsg)
            if err != nil {
                panic("JSON.Send: " + err.Error())
            }
        case lake, ok = <-c.Lakes:
            if ok {
                err = websocket.JSON.Send(c.Conn, lake)
                if err != nil {
                    panic("JSON.Send: " + err.Error())
                }
            } else {
                //if the channel is closed it means the game is over!
                jsonMsg := map[string]string{ "Message" : "Nertz" }
                err = websocket.JSON.Send(c.Conn, jsonMsg)
                if err != nil {
                    panic("JSON.Send: " + err.Error())
                }
                return
            }
        }
    }
}

func (s *Game) MakeMove(move *Move) bool {
    a := <-s.Lakes
    size := len(a.Piles[move.Pile].Cards)
    var resp bool
    if size == 0 && move.Card.Value == 1 {
        if move.Card.Value == 1 {
            a.Piles[move.Pile].Cards = append( a.Piles[move.Pile].Cards, move.Card )
            resp = true
        } else {
            resp = false
        }
    } else {
        top := a.Piles[move.Pile].Cards[size-1].Value
        suit := a.Piles[move.Pile].Cards[size-1].Suit
        if move.Card.Value != top + 1 || suit != move.Card.Suit || top == 13 {
            resp = false
        } else {
            a.Piles[move.Pile].Cards = append( a.Piles[move.Pile].Cards, move.Card )
            resp = true
        }
    }
    s.Lakes <- a
    s.Updates <- a
    return resp
}

func (g *Game) TallyUp() {
    a := <-g.Lakes
    for _, pile := range a.Piles {
        for _, card := range pile.Cards {
            g.Scoreboard[card.Player]++
        }
    }
}


func (g *Game) WaitForEnd(c *Client) {
    var scoreupdate map[string]interface{}
    err := websocket.JSON.Receive(c.Conn, &scoreupdate)
    if err != nil {
        log.Fatal(err)
    }
    scoreupdate["Player"] = c.Name
    g.ScoreChan <- scoreupdate
    done := <-g.Done
    g.Done <- done + 1
    c.Done = true
    if scoreupdate["Nertz"].(bool) {
        g.Over = true
        g.GameOver <- 1
    } else {
        //thanks for playing quitter!
        fmt.Fprintf(os.Stdout, "--------------------------------------------\n            %v quit player nertz\n--------------------------------------------\n", c.Name)
        scoreupdate["Value"] = float64(g.Scoreboard[c.Name])
        err = websocket.JSON.Send(c.Conn, map[string]int{ "Value" : g.Scoreboard[c.Name] })
        if err != nil {
            panic("JSON.Send: " + err.Error())
        }
        return
    }
    <-c.TalliedUp
    err = websocket.JSON.Send(c.Conn, g.Scoreboard)
    if err != nil {
        panic("JSON.Send: " + err.Error())
    }
}

func (g *Game) WriteScores() {
    for scoreupdate := range g.ScoreChan {
        g.Scoreboard[scoreupdate["Player"].(string)] -= 2*int(scoreupdate["Value"].(float64))
        done := <-g.Done
        if done == len(g.Clients) {
            for _, cli := range g.Clients {
                cli.TalliedUp <- 1
            }
        } else {
            g.Done <- done
        }
    }
}

