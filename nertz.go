package nertz

import (
    "fmt"
    "container/list"
    "math/rand"
    "strings"
    "strconv"
    "time"
    websocket "code.google.com/p/go.net/websocket"
)

/* Where we define the structures for Nertz and their constructors */

/* Common */

/** Gameplay **/

type Card struct{
    Value int
    Suit int
    Player string
}

type Move struct{
    Card Card
    Pile int
}


/*** Display ***/

func (c Card) Stringify() string {
    var suit string
    var value string
    switch c.Value {
    case 1:
        value = "A"
    case 10:
        value = "T"
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
    return fmt.Sprintf("%v %v", value, suit)
}

func Cardify(s string, player string) Card {
    cardinfo := strings.Split(s, "")
    if len(cardinfo) != 2 {
        return Card{}
    } else {
        svalue := cardinfo[0]
        ssuit := cardinfo[1]
        var suit int
        var value int
        switch svalue {
        case "A":
            value = 1
        case "T":
            value = 10
        case "J":
            value = 11
        case "Q":
            value = 12
        case "K":
            value = 13
        default:
            value, _ = strconv.Atoi(svalue)
        }
        switch ssuit {
        case "s":
            suit = 1
        case "h":
            suit = 2
        case "c":
            suit = 3
        case "d":
            suit = 4
        }
        return Card{ value, suit, player, }
    }
}

/** Communication **/

type Credentials struct {
    Username string
    Password string
}

/* Server */

/** Gameplay **/

type Pile struct {
    Cards []Card
}

type Lake struct {
    Piles []Pile
}

/** Communication **/

type Client struct {
    Conn *websocket.Conn
    Lakes chan Lake
    Messages chan string
    Name string
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

/*** Constructors ***/

func NewLake(players int) Lake {
    lake := Lake{ make([]Pile, players * 4), }
    for pile := range lake.Piles {
        lake.Piles[pile] = Pile{ make([]Card, 0, 13), }
    }
    return lake
}

func NewGame() *Game {
    var game *Game    = new(Game)
    // Max of 6 players
    game.Clients      = make([]*Client, 0, 6)
    game.Lakes        = make(chan Lake, 1)
    game.Updates      = make(chan Lake, 10)
    game.NewClients   = make(chan *Client, 6)
    game.ScoreChan    = make(chan map[string]interface{}, 6)
    game.GameOver     = make(chan int, 6)
    game.ReadyPlayers = make(chan int, 1)
    game.ReadyPlayers <- 0
    game.Begin        = make(chan int, 1)
    game.Started      = false
    game.Done         = make(chan int, 1)
    game.Done <- 0
    return game
}

func (game *Game) Init(players int) {
    lake := NewLake(players)
    game.Lakes <- lake
    game.Updates <- lake
}

func DisplayScoreboard(scores map[string]interface{}) {
    for key, val := range scores {
        fmt.Printf("%v = %v\n", key, val)
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

/*** Display ***/

func (l *Lake) Display() {
    var scard string
    scard ="Lake: %v"
    for pile := range l.Piles {
        var toprint string
        if pileLen := len(l.Piles[pile].Cards) ; pileLen != 0 {
            toprint = fmt.Sprintf(" %v ", l.Piles[pile].Cards[ pileLen - 1 ].Stringify())
        } else {
            toprint = "     "
        }
        var spile string
        if pile < 10 {
            spile = fmt.Sprintf(" %v", pile)
        } else {
            spile = fmt.Sprintf("%v", pile)
        }
        if (pile + 1) % 4 == 0 {
            toprint = fmt.Sprintf("%v: [%v]\n      %%v", spile, toprint )
        } else {
            toprint = fmt.Sprintf("%v: [%v] %%v", spile, toprint )
        }
        scard = fmt.Sprintf(scard, toprint)
    }
    scard = fmt.Sprintf(scard, "")
    scard = strings.TrimSpace(scard)
    fmt.Println(scard)
}

/* Client */

/** Gameplay **/

type Hand struct {
    Nertzpile *list.List
    Streampile *list.List
    River []*list.List
    Stream *list.List
}

/*** Constructors ***/

func NewShuffledDeck(player string) []Card {
    deck := make( []Card , 52)
    for i := 0; i < 4; i++ {
        for j := 0; j < 13; j++ {
            deck[ i * 13 + j ] = Card{ j + 1 , i + 1 , player, }
        }
    }
    for i := 51; i > 0; i-- {
        r := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
        j   := r.Intn(i+1)
        tmp := deck[i]
        deck[i] = deck[j]
        deck[j] = tmp
    }
    return deck
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
        hand.Streampile.PushFront(cards[i])
    }

    hand.Stream = list.New()

    return hand
}


/** Communication **/

type Player struct {
    Name string
    Password string
    Hand *Hand
    Conn *websocket.Conn
    Ready bool
    Started bool
    Done bool
    URL string
    Lakes chan Lake
    Messages chan map[string]interface{}
    Lake Lake
}

/*** Constructors ***/

func NewPlayer(name string, pw string, url string, ws *websocket.Conn) *Player {
    var player *Player = new(Player)
    player.Hand = NewHand(name)
    player.Conn = ws
    player.Name = name
    player.Password = pw
    player.URL = url
    player.Lake = Lake{}
    player.Ready = false
    player.Started = false
    player.Done = false
    player.Lakes = make(chan Lake, 10)
    player.Messages = make(chan map[string]interface{}, 10)
    return player
}

/*** Display ***/

func PrintCardStack(cs *list.List, toShow int) {
    stack := "[ %v"
    for e := cs.Front() ; e != nil ; e = e.Next() {
        if toShow > 0 {
            card := fmt.Sprintf("%v ]%%v", e.Value.(Card).Stringify())
            stack = fmt.Sprintf(stack, card)
            toShow--
        } else {
            stack = fmt.Sprintf(stack, "]%v")
        }
    }
    stack = fmt.Sprintf(stack, "")
    if stack == "[ " {
       fmt.Println("")
    } else {
        fmt.Println(stack)
    }
}

func (h *Hand) Display() {
    fmt.Print("Nertzpile: ")
    PrintCardStack(h.Nertzpile, 1)
    fmt.Print("\n")
    for pile := range h.River {
        if pile == 0 {
            fmt.Print("River: ")
        } else {
            fmt.Print("       ")
        }
        PrintCardStack(h.River[pile], h.River[pile].Len())
    }
    fmt.Print("\n")
    fmt.Print("Streampile: ")
    PrintCardStack(h.Streampile, 0)
    fmt.Print("\n")
    fmt.Print("Stream: ")
    PrintCardStack(h.Stream, 3 )
    fmt.Print("\n")
}

func (h *Hand) IsNertz() bool {
    return h.Nertzpile.Len() == 0
}
