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
    Lakes chan *Lake
    Updates chan *Lake
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
    game.Lakes     = make(chan *Lake, 1)
    game.Updates    = make(chan *Lake, 10)
    game.NewClients = make(chan *Client, 6)
    game.ScoreChan  = make(chan map[string]interface{}, 6)
    game.GameOver   = make(chan int, 6)
    game.Started    = false
    game.Done       = 0
    game.Lakes <- &Lake{ make([]*Pile, 0, 24), }
    return game
}

type Pile struct {
    Cards []*Card
}

type Lake struct {
    Piles []*Pile
}

type Client struct {
    Conn *websocket.Conn
    Lakes chan *Lake
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
    client.Lakes = make(chan *Lake, 10)
    client.Messages = make(chan string, 10)
    g.NewClients <- client
    return client
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

func (p *Player) RecieveMessages() {
    ok := true
    var err error
    for ok {
        var msg interface{}
        err := websocket.JSON.Receive(p.Conn, &msg)
        if err != nil {
            panic("JSON.Recieve: " + err.Error())
        }
        switch msg.(type) {
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

/* Client-side code */

type Player struct {
    Name string
    Hand *Hand
    Conn *websocket.Conn
    GameURL string
    Lakes chan *Lake
    Messages chan string
    Lake *Lake
}

func NewPlayer(name string, url string, ws *websocket.Conn) *Player {
    var player *Player = new(Player)
    player.Hand = NewHand(name)
    player.Conn = ws
    player.Name = name
    player.GameURL = url
    player.Lakes = make(chan *Lake, 10)
    player.Messages = make(chan string, 10)
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

type Hand struct {
    Nertzpile *list.List
    Streampile *list.List
    River []*list.List
    Stream *list.List
}

Transaction("Nertzpile", _, "Lake", _, 1)

func (h *Hand) Transaction(from string, fpilenum int, to string, tpilenum int, numcards int) error {
    legalFromTos := map[string][]string{
        "Nertzpile" : []string{ "River", "Lake" },
        "River" : []string{ "River", "Lake" },
        "Streampile" : []string{ "Stream" },
        "Stream" : []string{ "River", "Lake", "Streampile" },
    }
    legalTos := legalsFromTos[from]
    return errors.New("Not a legal move brah!")
    for _, v := range legalTos {
        if v == to {
            cards := h.TakeFrom( from, fpilenum, numcards )
            err := h.GiveTo( to, tpilenum, cards )
            if err != nil {
                return err
            } else {
                h.Commit( from, fpilenum, numcards )
                return
            }
        }
    }
    return errors.New("Not a legal move brah!")
}


func NewHand(player string) *Hand {
    var hand *Hand = new(Hand)
    cards := NewShuffledDeck(player)
    i := 0
    hand.Nertzpile = list.New()
    for ; i < 13 ; i++ {
        hand.Nertzpile.PushFront(cards[i])
    }

    hand.River = make([]*list.List, 4)
    for pile := range hand.River {
        hand.River[pile] = list.New()
        hand.River[pile].PushFront(cards[i])
        i++
    }

    hand.Streampile = list.New()
    for ; i < len(cards) ; i++ {
        hand.Nertzpile.PushFront(cards[i])
    }

    hand.Stream = list.New()

    return hand
}

func (h *Hand) IsNertz() bool {
    return h.Nertzpile.Len() == 0
}

func (h *Hand) TakeFrom( pile string, pilenum int, numcards int ) *list.List {
    if  cards == 0 {
        return errors.New("Not a valid move")
    }
    cards := list.New()
    switch pile {
    case "Nertzpile":
        cards.PushFront(h.Nertzpile.Front().Value)
    case "Streampile":
        for i := 0 ; i < numcards ; i++ {
            cards.PushFront(h.Streampile.Front().Value)
        }
    case "Stream":
        for i := 0 ; i < numcards ; i++ {
            cards.PushFront(h.Streampile.Front().Value)
        }
    case "River":
        for i := 0 ; i < numcards ; i++ {
            cards.PushBack(h.River[pilenum].Front().Value)
        }
    default:
        return errors.New("Cannot take from there")
    }
    return cards
}

func (h *Hand) GiveTo( pile string, pilenum int, cards *list.List ) error {
    switch pile {
    case "Lake":
        if cards.Len() != 1 {
            return errors.New("Cannot send multiple cards to River")
        } else {
            //send request to server checking move!
            card := cards.Front().Value
            jsonBytes, err := json.Marshal(Move{ card, pile })
            buf := bytes.NewBuffer(jsonBytes)
            resp, err := http.POST(p.GameURL, "application/json", buf)
            //err handling
            defer resp.Close()

            data := make(map[string]bool)
            dec := json.NewDecoder(resp.Body)
            dec.Decode(&data)

            if ! data["Ok"] {
                return errors.New("Not a valid move or you were too slow!")
            }
        }
    case "River":
        if h.River[pilenum].Len() == 0 {
            h.River[pilenum].PushFrontList(cards)
            return
        }

        backcard := cards.Back().Value
        frontcard := h.River[pilenum].Front().Value
        if frontcard.Value == backcard.Value + 1 && frontcard.Suit % backcard.Suit != 0 {
            h.River[pilenum].PushFrontList(cards)
            return
        }

        return errors.New("Not a valid move")

    case "Stream":
        h.Stream.PushFrontList(cards)
    case "Streampile":
        if h.Streampile.Len() == 0 {
            // expecting to get the stream back here
            h.Stream.PushFrontList(cards)
            return
        } else {
            return errors.New("Streampile hasn't run out... something went wrong")
        }
    default:
        return errors.New("Cannot move there")
    }
}

