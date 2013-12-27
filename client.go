package nertz

import (
    "encoding/json"
    "container/list"
    "bytes"
    "fmt"
    "strings"
    "net/http"
    "time"
    "os"
    "bufio"
    "strconv"
    "errors"
    websocket "code.google.com/p/go.net/websocket"
)

/* Client-side code */

func (p *Player) ReceiveMessages() {
    done := false
    for ! done {
        var msg interface{}
        err := websocket.JSON.Receive(p.Conn, &msg)
        if err != nil {
            panic("JSON.Recieve: " + err.Error())
        }

        switch val := msg.(type) {
        case map[string]interface{}:
            if _, ok := val["Piles"]; ok {
                var lake Lake
                byteLake, _ := json.Marshal(val)
                json.Unmarshal(byteLake, &lake)
                p.Lake = lake
            } else {
                //if this was a quit or gameover message stop looping
                _, ok := val["Value"]
                _, ook := val[p.Name]
                if ok || ook {
                    done = true
                }
                p.Messages <- val
            }
        }
    }
}

func (p *Player) EndGame(isQuitting bool) {
    if ! p.Done {
        jsonMsg := map[string]interface{}{
            "Value" : p.Hand.Nertzpile.Len(),
            "Nertz" : !isQuitting,
        }
        websocket.JSON.Send(p.Conn, jsonMsg)
        p.Done = true
    }
}

func (p *Player) HandleMessages() {
    waiting := 0
    PrintSeparator("=")
    for {
        select {
        case msg := <-p.Messages:
            contents, ok := msg["Message"]
            if ok {
                switch contents.(string) {
                case "Nertz":
                    fmt.Printf("\n! Looks like the game is over !\n")
                    p.EndGame(false)
                case "Let's Begin!":
                    fmt.Printf("\r\n-----------------------\n! Let the games begin !\n-----------------------\n")
                    p.Started = true
                case "Already Started!":
                    fmt.Printf("\r\n! The game already started !\n")
                case "Waiting on the other players...":
                    if ! p.Started {
                        fmt.Printf("%v\n", contents)
                    }
                case "In Progress":
                    fmt.Printf("\n! The game already started !\n")
                case "Credentials":
                    err := websocket.JSON.Send(p.Conn, Credentials{ p.Name, p.Password, })
                    if err != nil {
                        panic("JSON.Send: " + err.Error())
                    }
                default:
                    fmt.Printf("\n! %v !\n", contents)
                }
            } else {
                //display the scoreboard
                if val, ok := msg["Value"]; ok {
                    fmt.Printf( "You quit with a score of %v.\nThanks for playing!", val )
                } else {
                    if _, ok := msg[p.Name]; ok {
                        DisplayScoreboard(msg)
                    } else {
                        fmt.Println(msg)
                    }
                }
            }
        default:
            if ! p.Done && ! p.Ready {
                fmt.Print("\nLet the server know when you're ready....\n")
                p.ReceiveCommands()
            } else {
                if ! p.Started && ! p.Done {
                    s := "."
                    for i := 0 ; i < waiting ; i++ {
                        s = s + "."
                    }
                    for i := 0 ; i < 60 - waiting ; i++ {
                        s = s + " "
                    }
                    s = s + "\r"
                    fmt.Print(s)
                    waiting = ( waiting + 1 ) % 60
                    time.Sleep(100 * time.Millisecond)
                    waiting++
                } else {
                    if ! p.Done {
                        fmt.Print("\r")
                        PrintSeparator("=")
                        p.RenderBoard()
                        p.ReceiveCommands()
                    } else {
                        fmt.Println("\nGoodbye!\n")
                        return
                    }
                }
            }
        }
    }
}

/** User Communication **/

func PrintSeparator(char string) {
    for i := 0 ; i < 80 ; i++ {
        fmt.Print(char)
    }
    fmt.Print("\n")
}


func (p *Player) ReceiveCommands() {
    reader := bufio.NewReader(os.Stdin)

    fmt.Print("Enter Command: ")
    var scard, stocard, pile string
    var depth, pilenum int
    cmd, _ := reader.ReadString('\n')
    parts := strings.Split(strings.TrimSpace(cmd), " ")
    cmd     = parts[0]

    var err error
    switch cmd {
    case "ready":
        resp, _ := http.Get(p.URL + "/ready")
        defer resp.Body.Close()

        data := make( map[string]interface{} )
        dec := json.NewDecoder(resp.Body)
        dec.Decode(&data)
        p.Messages <- data
        p.Ready = true

    case "lake":
        if len(parts) == 3 && p.Started {
            scard   = parts[1]
            stocard = parts[2]
            pile, pilenum, depth = p.GetCard( scard )
            tocard, _ := strconv.Atoi( stocard )
            if tocard < len(p.Lake.Piles) {
                if depth == 0 {
                    p.Transaction( pile, pilenum, "Lake", tocard, depth + 1 )
                }
            }
        } else {
            fmt.Fprintf(os.Stderr,"usage: %v <card> <pile>\n", cmd)
        }
    case "move":
        if len(parts) == 3 && p.Started {
            scard   = parts[1]
            stocard = parts[2]
            pile, pilenum, depth = p.GetCard( scard )
            topile, topilenum, _ := p.GetCard( stocard )
            err = p.Transaction( pile, pilenum, topile, topilenum, depth + 1 )
            if err != nil {
                fmt.Println(err)
            }
        } else {
            fmt.Fprintf(os.Stderr,"usage: %v <from> <to>\n", cmd)
        }
    case "fish":
        for topile := range p.Hand.River {
            if p.Hand.River[topile].Len() == 0  {
                p.Transaction( "Nertzpile", -1, "River", topile, 1 )
            }
        }
    case "draw":
        drawlen := p.Hand.Streampile.Len()
        switch {
        case 0 == drawlen:
            if streamlen := p.Hand.Stream.Len(); streamlen != 0 {
                err = p.Transaction( "Stream", -1, "Streampile", -1, streamlen )
            }
        case drawlen < 3:
            err = p.Transaction( "Streampile", -1, "Stream", -1, drawlen )
        default:
            err = p.Transaction( "Streampile", -1, "Stream", -1, 3 )
        }
        if err != nil {
            fmt.Println(err)
        }
    case "nertz":
        if p.Hand.Nertzpile.Len() == 0 && p.Started {
            p.EndGame(false)
        }
    case "quit":
        p.EndGame(true)
        time.Sleep(100 * time.Millisecond)
    case "help":
        fmt.Print("Your commands are these:\n  draw: reveals the next three cards in your stream\n  move <thiscardname> <thatcardname>: moves this card under that card. Both cards are assumed to be in your hand.\n    eg: move 4h 5s || move 9c Td || move Qc Kh\n  fish: fills an empty space in your river with the top card of your nertz pile\n  lake <thiscardname> <pilenumber>: moves this card from your hand to the specified pile in the lake\n    eg: lake As 1 ||  lake 2s 1 || lake 3s 1 || lake Ad 2\n")
    default:
        fmt.Print("Not a valid command. Use `help` for more details.\n")

    }
}

/**** Gameplay
 *
 * e.g. Transaction("Nertzpile", _, "Lake", _, 1)
 *
 ****/

func (p *Player) RenderBoard() {
    fmt.Print("\n")
    p.Lake.Display()
    fmt.Print("\n")
    p.Hand.Display()
}

func (p *Player) GetCard( scard string ) ( string, int, int ) {
    h := p.Hand
    card := Cardify( scard, p.Name )
    e := h.Nertzpile.Front()
    if card == e.Value.(Card) {
        println("Found in the nertz-pile...")
        return "Nertzpile", 0, 0
    }
    if h.Stream.Len() > 0 {
        e = h.Stream.Front()
        if card == e.Value.(Card) {
            println("Found in the stream...")
            return "Stream", 0, 0
        }
    }

    for pile := range h.River {
        depth := 0
        for e = h.River[pile].Front() ; e != nil ; e = e.Next() {
            if card == e.Value.(Card) {
                println("Found in the river...")
                return "River", pile, depth
            }
            depth++
        }
    }
    return "", -1, -1
}

func (p *Player) Transaction(from string, fpilenum int, to string, tpilenum int, numcards int) error {
    h := p.Hand
    legalFromTos := map[string][]string{
        "Nertzpile" : []string{ "River", "Lake" },
        "River" : []string{ "River", "Lake" },
        "Streampile" : []string{ "Stream" },
        "Stream" : []string{ "River", "Lake", "Streampile" },
    }
    legalTos, ok := legalFromTos[from]
    if !ok {
        return errors.New("Not a legal move brah!")
    }
    for _, v := range legalTos {
        if v == to {
            cards := h.TakeFrom( from, fpilenum, numcards )
            err := h.GiveTo( to, tpilenum, cards, p.URL + "/move" )
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
    var e *list.Element
    switch pile {
    case "Nertzpile":
        e = h.Nertzpile.Front()
    case "Streampile":
        e = h.Streampile.Front()
    case "Stream":
        e = h.Stream.Front()
    case "River":
        e = h.River[pilenum].Front()
    }
    for i := 0 ; i < numcards ; i++ {
        if pile != "River" {
            cards.PushFront(e.Value)
        } else {
            cards.PushBack(e.Value)
        }
        e = e.Next()
    }
    return cards
}

func (h *Hand) GiveTo( pile string, pilenum int, cards *list.List, url string) error {
    switch pile {
    case "Lake":
        if cards.Len() != 1 {
            return errors.New("Cannot send multiple cards to River")
        } else {
            //send request to server checking move!
            card := cards.Front().Value.(Card)
            jsonBytes, _ := json.Marshal(Move{ card, pilenum })
            buf := bytes.NewBuffer(jsonBytes)
            resp, _ := http.Post(url, "application/json", buf)
            //err handling
            defer resp.Body.Close()

            data := make(map[string]bool)
            dec := json.NewDecoder(resp.Body)
            dec.Decode(&data)

            if ! data["Ok"] {
                return errors.New("Not a valid move or you were too slow!")
            }
            // Wait for the new lake
            time.Sleep(100 * time.Millisecond)
        }
    case "River":
        if h.River[pilenum].Len() == 0 {
            h.River[pilenum].PushFrontList(cards)
        } else {
            backcard := cards.Back().Value.(Card)
            frontcard := h.River[pilenum].Front().Value.(Card)
            if frontcard.Value == backcard.Value + 1 && ( frontcard.Suit + backcard.Suit ) % 2 != 0 {
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
            h.Streampile.PushFrontList(cards)
        } else {
            return errors.New("Streampile hasn't run out... something went wrong")
        }
    default:
        return errors.New("Cannot move there")
    }
    return nil
}
