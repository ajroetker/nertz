package nertz

import (
    "container/list"
    "bytes"
    "net/http"
    "errors"
    "log"
    "math/rand"
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
    client.Arenas = make(chan *Arena, 10)
    client.Messages = make(chan string, 10)
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

type Player struct {
    Name string
    Hand *Hand
    Conn *websocket.Conn
    GameURL string
    Arenas chan *Arena
    Messages chan string
    Moves chan *PlayerMove
}

func NewPlayer(name string, url string, ws *websocket.Conn) *Player {
    var player *Player = new(Player)
    player.Hand = NewHand()
    player.Conn = ws
    player.Name = name
    player.GameURL = url
    player.Arenas = make(chan *Arena, 10)
    player.Messages = make(chan string, 10)
    player.Moves = make(chan string, 10)
    return player
}

func NewShuffledDeck(player string) []*Card {
    deck := make([]*Card, 52)
    for i := 1; i <= 4; i++ {
        for j := 1; j<= 13; j++ {
            deck[i*j] = &Card{ j, i, player, }
        }
    }
    for i := 51; i > 0; i-- {
        j   := rand.Intn(i+1)
        tmp := deck[i]
        deck[i] = deck[j]
        deck[j] = tmp
    }
    return deck
}

func NewHand() *Hand {
    var hand *Hand = new(Hand)
    cards := NewShuffledDeck(player string)
    i := 0
    hand.Nertzpile = list.New()
    for ; i < 13 ; i++ {
        hand.Nertzpile.PushFront(cards[i])
    }

    hand.Lake = make([]*list.List, 4)
    for pile := range hand.Lake {
        hand.Lake[pile] = list.New()
        hand.Lake[pile].PushFront(cards[i])
        i++
    }

    hand.Streampile = list.New()
    for ; i < len(cards) ; i++ {
        hand.Nertzpile.PushFront(cards[i])
    }

    return hand
}

type Hand struct {
    Nertzpile *list.List
    Streampile *list.List
    Lake []*list.List
}

func (h *Hand) IsNertz() bool {
    return h.Nertzpile.Len() == 0
}

type PlayerMove struct {
    From map[string]interface{}
    To map[string]interface{}
    Cards *list.List
}

func (h *Hand) TakeFrom( pile string, numcards int ) *list.List {
    if  cards == 0 {
        return errors.New("Not a valid move")
    }
    cards := list.New()
    switch pile {
    case "Nertzpile":

    case "Streampile":
    case "Lake":
    default:
        return errors.New("Cannot take from there")
    }
}

func (h *Hand) GiveTo( pile string, pilenum int, cards *list.List ) error {
    switch pile {
    case "Arena":
        if cards.Len() != 1 {
            return errors.New("Cannot send multiple cards to River")
        } else {
            //send request to server checking move!
            card := cards.Front().Value
            jsonBytes, err := json.Marshal(Move{ card, pile })
            buf := bytes.NewBuffer(jsonBytes)
            resp, err := http.POST(p.GameURL + "/move", "application/json", buf)
            //err handling
            defer resp.Close()

            data := make(map[string]bool)
            dec := json.NewDecoder(resp.Body)
            dec.Decode(&data)

            if ! data["Ok"] {
                return errors.New("Not a valid move")
            }
        }
    case "Lake":
        if h.Lake[pilenum].Len() == 0 {
            h.Lake[pilenum].PushFrontList(cards)
            return
        }

        backcard := cards.Back().Value
        frontcard := h.Lake[pilenum].Front().Value
        if frontcard.Value == backcard.Value + 1 && frontcard.Suit % backcard.Suit != 0 {
            h.Lake[pilenum].PushFrontList(cards)
            return
        }

        return errors.New("Not a valid move")

    default:
        return errors.New("Cannot move there")
    }
}

func (p *Player) Valid(pm *PlayerMove) bool {

    num, ok := m.To["Number"]
    pile, err := strconv.Atoi(num)
    if !ok || err != nil {
        fmt.Println("Move did not contain a pile number or was not an integer")
        return false
    }
    switch m.To["Pile"] {
    case "Arena":
        if m.Cards.Len() != 1 {
            fmt.Println("Cannot send multiple cards to River")
            h.UndoMove(m)
        } else {
            //send request to server checking move!
            card := m.Cards.Front().Value
            jsonBytes, err := json.Marshal(Move{ card, pile })
            buf := bytes.NewBuffer(jsonBytes)
            resp, err := http.POST(p.GameURL + "/move", "application/json", buf)
            //err handling
            defer resp.Close()

            data := make(map[string]bool)
            dec := json.NewDecoder(resp.Body)
            dec.Decode(&data)

            if ! data["Ok"] {
                fmt.Println("Not a valid move")
                h.UndoMove(m)
            }
        }
    case "Lake":

    }
}

func (p *Player) MakeMoves() {
    legalTo := map[string]bool{
        "Lake"  : true,
        "Arena" : true,
    }
    legalFrom := map[string]bool{
        "Lake"       : true,
        "Streampile" : true,
        "Nertzpile"  : true,
    }
    for m := range p.Moves {
        if ! legalFrom[m.From] || m.Cards.Len() == 0 {
            fmt.Println("Don't know what you just did weirdo!")
            break
        }
        if ! legalTo[m.To] {
            fmt.Println("Not a valid move")
            h.UndoMove(m)
        }

        num, ok := m.To["Number"]
        pile, err := strconv.Atoi(num)
        if !ok || err != nil {
            fmt.Println("Move did not contain a pile number or was not an integer")
            h.UndoMove(m)
        }
        switch m.To["Pile"] {
        case "Arena":
            if m.Cards.Len() != 1 {
                fmt.Println("Cannot send multiple cards to River")
                h.UndoMove(m)
            } else {
                //send request to server checking move!
                card := m.Cards.Front().Value
                jsonBytes, err := json.Marshal(Move{ card, pile })
                buf := bytes.NewBuffer(jsonBytes)
                resp, err := http.POST(p.GameURL + "/move", "application/json", buf)
                //err handling
                defer resp.Close()

                data := make(map[string]bool)
                dec := json.NewDecoder(resp.Body)
                dec.Decode(&data)

                if ! data["Ok"] {
                    fmt.Println("Not a valid move")
                    h.UndoMove(m)
                }
            }
        case "Lake":
            
        }

    }
}
