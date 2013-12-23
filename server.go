package nertz

import (
    "log"
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
        g.Clients[len(g.Clients)] = cli
    }
}

func (g *Game) WaitForStart(c *Client) {
    <-g.Begin
    c.Messages <- "Let's Begin!"
    return
}

func (g *Game) BroadcastMessages() {
    for {
        select {
        case lake := <-g.Updates:
            for _, c := range g.Clients {
                c.Lakes <- lake
            }
        case <-g.GameOver:
            g.TallyUp()
            close(g.Lakes)
            for _, c := range g.Clients {
                close(c.Lakes)
            }
            return
        }
    }
}

func (c *Client) SendMessages() {
    ok := true
    var err error
    for ok {
        select {
        case msg := <-c.Messages:
            jsonMsg := map[string]string{ "Message" : msg }
            err = websocket.JSON.Send(c.Conn, jsonMsg)
        case lake, ok := <-c.Lakes:
            if ok {
                err =  websocket.JSON.Send(c.Conn, lake)
            } else {
                //if the channel is closed it means the game is over!
                jsonMsg := map[string]string{ "Message" : "Nertz" }
                err =  websocket.JSON.Send(c.Conn, jsonMsg)
            }
        }
        if err != nil {
            panic("JSON.Send: " + err.Error())
        }
    }
}

func (s *Game) MakeMove(move *Move) bool {
    a := <-s.Lakes
    size := len(a.Piles[move.Pile].Cards)
    var resp bool
    if size == 0 && move.Card.Value == 1 {
        a.Piles[move.Pile].Cards[size] = move.Card
        resp = true
    } else {
        top := a.Piles[move.Pile].Cards[size-1].Value
        suit := a.Piles[move.Pile].Cards[0].Suit
        if move.Card.Value != top + 1 || suit != move.Card.Value || top == 13 {
            resp = false
        } else {
            a.Piles[move.Pile].Cards[size] = move.Card
            resp = true
        }
    }
    s.Lakes <- a
    s.Updates <- a
    return resp
}

func (g *Game) TallyUp() {
    a := <-g.Lakes
    var scores map[string]int
    for _, pile := range a.Piles {
        for _, card := range pile.Cards {
            scores[card.Player]++
        }
    }
    g.Scoreboard = scores
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
    if done == 0 {
        if scoreupdate["Nertz"].(bool) {
            g.GameOver <- 1
        } else {
            //thanks for playing quitter!
            scoreupdate["Value"] = g.Scoreboard[c.Name]
            err = websocket.JSON.Send(c.Conn, scoreupdate)
            if err != nil {
                panic("JSON.Send: " + err.Error())
            }
            return
        }
    }
    //Wait for the channel to close
    // which signals that we've collected all the scores
    <-g.ScoreChan
    err = websocket.JSON.Send(c.Conn, g.Scoreboard)
    if err != nil {
        panic("JSON.Send: " + err.Error())
    }
}

func (g *Game) WriteScores() {
    <-g.GameOver
    for scoreupdate := range g.ScoreChan {
        g.Scoreboard[scoreupdate["Player"].(string)] -= 2*scoreupdate["Value"].(int)
        done := <-g.Done
        if done == len(g.Clients) {
            close(g.ScoreChan)
        } else {
            g.Done <- done
        }
    }
}

