package nertz

import (
    "container/list"
    "log"
    websocket "code.google.com/p/go.net/websocket"
)

type Card struct{
    Value int
    Suit int
    Player string
}

/* Server-side code */

type Move struct{
    Card *Card
    Pile int
}

type Game struct {
    Clients []*Client
    Arenas chan *Arena
    Updates chan *Arena
    NewClients chan *Client
    ScoreChan chan map[string]interface{}
    GameOver chan int
    Started bool
    Done int
    Scoreboard map[string]int
}

func NewGame() *Game {
    var game *Game  = new(Game)
    game.Clients    = make([]*Client, 0, 6)
    game.Arenas     = make(chan *Arena, 1)
    game.Updates    = make(chan *Arena, 10)
    game.NewClients = make(chan *Client, 6)
    game.ScoreChan  = make(chan map[string]interface{}, 6)
    game.GameOver   = make(chan int, 6)
    game.Started    = false
    game.Done       = 0
    game.Arenas <- &Arena{ make([]*Pile, 0, 24), }
    return game
}

type Pile struct {
    Cards []*Card
}

type Arena struct {
    Piles []*Pile
}

type Client struct {
    Conn *websocket.Conn
    Arenas chan *Arena
    Messages chan string
    Name string
}

func (g *Game) AddNewClients() {
    for cli := range g.NewClients {
        g.Clients[len(g.Clients)] = cli
    }
}

func (g *Game) NewClient(ws *websocket.Conn) *Client {
    var client *Client = new(Client)
    client.Conn = ws
    client.Arenas = make(chan *Arena)
    client.Messages = make(chan string)
    g.NewClients <- client
    return client
}

func (g *Game) TallyUp() {
    a := <-g.Arenas
    var scores map[string]int
    for _, pile := range a.Piles {
        for _, card := range pile.Cards {
            scores[card.Player]++
        }
    }
    g.Scoreboard = scores
}

func (g *Game) BroadcastMessages() {
    for {
        select {
        case arena := <-g.Updates:
            for _, c := range g.Clients {
                c.Arenas <- arena
            }
        case <-g.GameOver:
            g.TallyUp()
            close(g.Arenas)
            for _, c := range g.Clients {
                close(c.Arenas)
            }
            return
        }
    }
}

func (g *Game) WaitForEnd(c *Client) {
    var hand Hand
    err := websocket.JSON.Receive(c.Conn, &hand)
    if err != nil {
        log.Fatal(err)
    }
    if g.Done == 0  {
        if hand.IsNertz() {
            g.GameOver <- 1
        } else {

            scoreupdate := map[string]interface{}{
                "Player" : c.Name,
                "Value"  : hand.Nertzpile.Len(),
            }

            g.ScoreChan <- scoreupdate
            return
            //thanks for playing quitter!
        }
    }
    scoreupdate := map[string]interface{}{
        "Player" : c.Name,
        "Value"  : hand.Nertzpile.Len(),
    }
    g.ScoreChan <- scoreupdate
    <-g.ScoreChan
    err = websocket.JSON.Send(c.Conn, g.Scoreboard)
    if err != nil {
        panic("JSON.Send: " + err.Error())
    }
}

type Credentials struct {
    Username string
    Password string
}

func (g *Game) WriteScores() {
    <-g.GameOver
    for scoreupdate := range g.ScoreChan {
        g.Done++
        g.Scoreboard[scoreupdate["Player"].(string)] -= scoreupdate["Value"].(int)
        if g.Done == len(g.Clients) {
            close(g.ScoreChan)
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
        case arena, ok := <-c.Arenas:
            if ok {
                err =  websocket.JSON.Send(c.Conn, arena)
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

func (c *Client) GetCredentials() *Credentials {
    c.Messages <- "Credentials"
    var creds Credentials
    err := websocket.JSON.Receive(c.Conn, &creds)
    if err != nil {
        log.Fatal(err)
    }
    return &creds
}

func (s *Game) MakeMove(move *Move) bool {
    a := <-s.Arenas
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
    s.Arenas <- a
    s.Updates <- a
    return resp
}

/* Client-side code */

type Hand struct {
    Nertzpile *list.List
    Streampile *list.List
    Lake []*list.List
    Moves chan *PlayerMove
    Responses chan bool
}

func (h *Hand) IsNertz() bool {
    return h.Nertzpile.Len() == 0
}

type PlayerMove struct {
    From string
    To map[string]string
    Cards *list.List
}

/*
func (h *Hand) MakeMoves() {
    for m := range h.Moves {
        
    }

}
func (h *Hand) Valid(m *PlayerMove) bool {
    if m.To["Pile"] == "Arena" && m.Cards.Len() == 1 {
        num := m.To["Number"]
        if num == nil {
            // send an error message
        }
        pile, err := strconv.Atoi(num)
        if err != nil {
            log.Fatal(err)
        }
        //send request to server checking move!
        //card := m.Cards.Front().Value
        //Lets not use a socket here, let's just use a request to another url
        //  the url can have the encoded params
        //err := websocket.JSON.Send(Move{ cards, pile})
        //websocket.JSON.Recieve() ===> message of whether it worked!
        // return reponse := true/false
    } else {
         return false
    }
}
*/
