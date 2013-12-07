package nertz

import (
    "log"
    "container/list"
)

type Card struct{
    Value int
    Suit int
    Player string
}

/* Client :: httpreq <- move
   Server :: if Valid(move) => Arena.Moves <- move && httpresp <- true
             else httpresp <- false
             continue listening for requests
*/

/* Server-side code */

type Move struct{
    Card *Card
    Pile int
}

type Response struct {
    Ok bool
}

type Game struct {
    Clients []*Client
    Arenas chan *Arena
    Updates chan *Arena
    Messages chan string
    NewClients chan *Client
    Quiters chan *Client
    GameOver chan int
    Done int
}

func NewGame() {
    var game *Game  = new(Game)
    game.Clients    = make([]*Client, 0, 6)
    game.Arenas     = make(chan *Arena)
    game.Updates    = make(chan *Arena, 10)
    game.Messages   = make(chan string, 10)
    game.NewClients = make(chan *Client, 6)
    game.Quiters    = make(chan *Client, 6)
    game.GameOver   = make(chan int)
    game.Done       = false
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
    GameOver chan int
    Name string
}

func (g *Game) AddNewClients() {
    for cli := range <-g.NewClients {
        g.Clients[len(cli)] = cli
    }
}

func (g *Game) NewClient(ws *websocket.Conn) *Client {
    var client *Client = new(Client)
    client.Conn = ws
    client.Arenas = make(chan *Arena)
    client.Messages = make(chan string)
    client.GameOver = make(chan int)
    g.NewClients <- client
    return client
}

func (g *Game) TallyUp() map[string]int {
    a := <-g.Arenas
    var scores map[string]int
    for _, pile := range a.Piles {
        for _, card := range a.Cards {
            scores[card.Player]++
        }
    }
    return scores
}

func (g *Game) BroadcastMessages() {
    for {
        select {
        case msg := <-g.Messages:
            for _, c := range g.Clients {
                c.Messages <- msg
            }
        case arena := <-g.Updates:
            for _, c := range g.Clients {
                c.Arenas <- arena
            }
        case <-g.GameOver:
            g.Done = true
            scores := g.TallyUp()
            for _, c := range g.Clients {
                c.GameOver <- scores[c.Name]
            }
        }
    }
}

func (s *Game) MakeMove(m *Move) bool {
    a := <-s.Arenas
    size := len(a.Piles[move.Pile].Cards)
    if size == 0 && move.Card.Value == 1 {
        a.Piles[move.Pile].Cards[size] = move.Card
        resp := true
    } else {
        top := a.Piles[move.Pile].Cards[size-1].value
        suit := a.Piles[move.Pile].Cards[0].Suit
        if move.Card.Value != top + 1 || suit != move.Card.Value || top == 13 {
            resp := false
        } else {
            a.Piles[move.Pile].Cards[size] = move.Card
            resp := true
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
    return Nertzpile.Len() == 0
}

type PlayerMove struct {
    From string
    To map[string]string
    Cards *list.List
}

func (h *Hand) MakeMoves() {
    for m := range <-h.Moves {
        
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
