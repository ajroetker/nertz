package nertz

import (
    "encoding/json"
    "container/list"
    "fmt"
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
    Card Card
    Pile int
}

type Game struct {
    Clients []*Client
    Lakes chan Lake
    Updates chan Lake
    NewClients chan *Client
    ScoreChan chan map[string]interface{}
    GameOver chan int
    ReadyPlayers chan int
    Begin chan int
    Started bool
    Done chan int
    Scoreboard map[string]int
}

func NewGame(players int) *Game {
    var game *Game    = new(Game)
    game.Clients      = make([]*Client, 0, players)
    game.Lakes        = make(chan Lake, 1)
    game.Updates      = make(chan Lake, 10)
    game.NewClients   = make(chan *Client, players)
    game.ScoreChan    = make(chan map[string]interface{}, players)
    game.GameOver     = make(chan int, players)
    game.ReadyPlayers = make(chan int, players)
    game.Begin        = make(chan int)
    game.Started      = false
    game.Done         = make(chan int, 1)
    lake := Lake{ make([]Pile, players * 4), }
    for pile := range lake.Piles {
        lake.Piles[pile] = Pile{ make([]Card, 0, 13), }
    }
    game.Lakes <- lake
    return game
}

type Pile struct {
    Cards []Card
}

type Lake struct {
    Piles []Pile
}

func (c *Card) Stringify() string {
    var suit string
    var value string
    switch c.Value {
    case 1:
        value = "A"
    case 11:
        value = "J"
    case 12:
        value = "Q"
    case 13:
        value = "K"
    default:
        value = fmt.Sprintf("%v", c.Value)
    }
    switch c.Suit {
    case 1:
        suit = "\xE2\x99\xA0"
    case 2:
        suit = "\xE2\x99\xA5"
    case 3:
        suit = "\xE2\x99\xA3"
    case 4:
        suit = "\xE2\x99\xA6"
    }
    return fmt.Sprintf("%v%v", suit, value)
}

func (l *Lake) Display() {
    for pile := range l.Piles {
        fmt.Println(l.Piles[pile].Cards[0].Stringify())
    }
}


type Client struct {
    Conn *websocket.Conn
    Lakes chan Lake
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
    client.Lakes = make(chan Lake, 10)
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

func (g *Game) WaitForStart(c *Client) {
    <-g.Begin
    c.Messages <- "Let's Begin!"
    return
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

type Credentials struct {
    Username string
    Password string
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
    for {
        var msg interface{}
        err := websocket.JSON.Receive(p.Conn, &msg)
        if err != nil {
            panic("JSON.Recieve: " + err.Error())
        }
        switch val := msg.(type) {
        case Lake:
            p.Lakes <- val
        case map[string]interface{}:
            p.Messages <- val
        }
    }
}

func (p *Player) HandleMessages() {
    var err error
    for {
        select {
        case msg := <-p.Messages:
            contents, ok := msg["Message"]
            if ok {
                switch contents {
                case "Nertz":
                    if ! p.Done {
                        //FIXME
                        jsonMsg := map[string]interface{}{
                            "Value" : p.Hand.Nertzpile.Len(),
                            "Nertz" : true,
                        }
                        err = websocket.JSON.Send(p.Conn, jsonMsg)
                    }
                case "Let's Begin!":
                case "Already Started!":
                case "In Progress":
                default:
                }
            } else {
                //display the scoreboard
            }
        case lake, ok := <-p.Lakes:
            //TODO Display the Lake
            if ok {
                lake.Display()
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
    Done bool
    GameURL string
    Lakes chan Lake
    Messages chan map[string]interface{}
    Lake Lake
}

func NewPlayer(name string, url string, ws *websocket.Conn) *Player {
    var player *Player = new(Player)
    player.Hand = NewHand(name)
    player.Conn = ws
    player.Name = name
    player.GameURL = url
    player.Done = false
    player.Lakes = make(chan Lake, 10)
    player.Messages = make(chan map[string]interface{}, 10)
    return player
}

func NewShuffledDeck(player string) []*Card {
    deck := make([]*Card, 52)
    for i := 1; i <= 4; i++ {
        for j := 1; j<= 13; j++ {
            deck[i*j-1] = &Card{ j, i, player, }
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

func PrintCardStack(cs *list.List, toShow int) {
    stack := "[%v"
    for e := cs.Front() ; e != nil ; e = e.Next() {
        if toShow > 0 {
            card := fmt.Sprintf("%v%v]%%v", e.Value.(*Card).Stringify())
            stack = fmt.Sprintf(stack, card)
            toShow--
        } else {
            stack = fmt.Sprintf(stack, "]%v")
        }
    }
    stack = fmt.Sprintf(stack, "")
    if stack == "[" {
       fmt.Println("empty")
    } else {
        fmt.Println(stack)
    }
}

func (h *Hand) Display() {
    PrintCardStack(h.Nertzpile, 1)
    PrintCardStack(h.Streampile, 0)
    for pile := range h.River {
        PrintCardStack(h.River[pile], h.River[pile].Len())
    }
    PrintCardStack(h.Stream, 3)
}

//Transaction("Nertzpile", _, "Lake", _, 1)

func (h *Hand) Transaction(from string, fpilenum int, to string, tpilenum int, numcards int, gameURL string) error {
    legalFromTos := map[string][]string{
        "Nertzpile" : []string{ "River", "Lake" },
        "River" : []string{ "River", "Lake" },
        "Streampile" : []string{ "Stream" },
        "Stream" : []string{ "River", "Lake", "Streampile" },
    }
    legalTos := legalFromTos[from]
    return errors.New("Not a legal move brah!")
    for _, v := range legalTos {
        if v == to {
            cards := h.TakeFrom( from, fpilenum, numcards )
            err := h.GiveTo( to, tpilenum, cards, gameURL )
            if err != nil {
                return err
            } else {
                h.Commit( from, fpilenum, numcards )
                return nil
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


func (h *Hand) Commit( pile string, pilenum int, numcards int ) {
    switch pile {
    case "Nertzpile":
        h.Nertzpile.Remove(h.Nertzpile.Front())
    case "Streampile":
        for i := 0 ; i < numcards ; i++ {
            h.Streampile.Remove(h.Streampile.Front())
        }
    case "Stream":
        for i := 0 ; i < numcards ; i++ {
            h.Stream.Remove(h.Stream.Front())
        }
    case "River":
        for i := 0 ; i < numcards ; i++ {
            h.River[pilenum].Remove(h.River[pilenum].Front())
        }
    }
    return
}

func (h *Hand) TakeFrom( pile string, pilenum int, numcards int ) *list.List {
    cards := list.New()
    if numcards == 0 {
        return cards
    }
    switch pile {
    case "Nertzpile":
        cards.PushFront(h.Nertzpile.Front().Value)
    case "Streampile":
        for i := 0 ; i < numcards ; i++ {
            cards.PushFront(h.Streampile.Front().Value)
        }
    case "Stream":
        for i := 0 ; i < numcards ; i++ {
            cards.PushBack(h.Stream.Front().Value)
        }
    case "River":
        for i := 0 ; i < numcards ; i++ {
            cards.PushBack(h.River[pilenum].Front().Value)
        }
    }
    return cards
}

func (h *Hand) GiveTo( pile string, pilenum int, cards *list.List, gameURL string) error {
    switch pile {
    case "Lake":
        if cards.Len() != 1 {
            return errors.New("Cannot send multiple cards to River")
        } else {
            //send request to server checking move!
            card := cards.Front().Value.(*Card)
            jsonBytes, _ := json.Marshal(Move{ *card, pilenum })
            buf := bytes.NewBuffer(jsonBytes)
            resp, _ := http.Post(gameURL, "application/json", buf)
            //err handling
            defer resp.Body.Close()

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
        } else {
            backcard := cards.Back().Value.(*Card)
            frontcard := h.River[pilenum].Front().Value.(*Card)
            if frontcard.Value == backcard.Value + 1 && frontcard.Suit % backcard.Suit != 0 {
                h.River[pilenum].PushFrontList(cards)
            } else {
                return errors.New("Not a valid move")
            }
        }


    case "Stream":
        h.Stream.PushFrontList(cards)
    case "Streampile":
        if h.Streampile.Len() == 0 {
            // expecting to get the stream back here
            h.Stream.PushFrontList(cards)
        } else {
            return errors.New("Streampile hasn't run out... something went wrong")
        }
    default:
        return errors.New("Cannot move there")
    }
    return nil
}

