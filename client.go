package nertz

import (
    "encoding/json"
    "container/list"
    "bytes"
    "fmt"
    "net/http"
    "os"
    "bufio"
    "strconv"
    "errors"
    websocket "code.google.com/p/go.net/websocket"
)

/* Client-side code */

func (p *Player) ReceiveMessages() {
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
                        jsonMsg := map[string]interface{}{
                            "Value" : p.Hand.Nertzpile.Len(),
                            "Nertz" : true,
                        }
                        err = websocket.JSON.Send(p.Conn, jsonMsg)
                        p.Done = true
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
            if ok {
                p.Lake = lake
            }
        }
        if err != nil {
            panic("JSON.Send: " + err.Error())
        }
    }
}

/** User Communication **/

func (p *Player) ReceiveCommands() {
    reader := bufio.NewReader(os.Stdin)

    fmt.Print("Enter Command: ")
    var cmd, scard, stocard, pile string
    var depth, pilenum int
    fmt.Fscanln(reader, cmd, scard, stocard)
    if scard != "" {
        pile, pilenum, depth = p.GetCard( scard )
    }

    switch cmd {
    case "lake":
        tocard, _ := strconv.Atoi( stocard )
        if depth == 0 {
            p.Transaction( pile, pilenum, "Lake", tocard, depth + 1 )
        }
    case "move":
        topile, topilenum, depth := p.GetCard( stocard )
        if depth == 0 {
            p.Transaction( pile, pilenum, topile, topilenum, depth + 1 )
        }
    case "fish":
        for topile := range p.Hand.River {
            if p.Hand.River[topile].Len() == 0  {
                p.Transaction( "Nertzpile", -1, "River", -1, 1 )
            }
        }
    case "draw":
        drawlen := p.Hand.Streampile.Len()
        switch {
        case 0 == drawlen:
            if streamlen := p.Hand.Stream.Len(); streamlen != 0 {
                p.Transaction( "Stream", -1, "Streampile", -1, streamlen )
            }
        case drawlen < 3:
            p.Transaction( "Sreampile", -1, "Stream", -1, drawlen )
        default:
            p.Transaction( "Sreampile", -1, "Stream", -1, 3 )
        }
    case "quit":
        jsonMsg := map[string]interface{}{
            "Value" : p.Hand.Nertzpile.Len(),
            "Nertz" : true,
        }
        websocket.JSON.Send(p.Conn, jsonMsg)
        p.Done = true
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
    p.Lake.Display()
    p.Hand.Display()
}

func (p *Player) GetCard( scard string ) ( string, int, int ) {
    h := p.Hand
    card := Cardify( scard, p.Name )
    e := h.Nertzpile.Front()
    if card == e.Value.(Card) {
        return "Nertzpile", 0, 0
    }
    e = h.Stream.Front()
    if card == e.Value.(Card) {
        return "Stream", 0, 0
    }
    for pile := range h.River {
        depth := 0
        for e = h.River[pile].Front() ; e != nil ; e = e.Next() {
            if card == e.Value.(Card) {
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
    legalTos := legalFromTos[from]
    return errors.New("Not a legal move brah!")
    for _, v := range legalTos {
        if v == to {
            cards := h.TakeFrom( from, fpilenum, numcards )
            err := h.GiveTo( to, tpilenum, cards, p.GameURL )
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

func (h *Hand) GiveTo( pile string, pilenum int, cards *list.List, url string) error {
    switch pile {
    case "Lake":
        if cards.Len() != 1 {
            return errors.New("Cannot send multiple cards to River")
        } else {
            //send request to server checking move!
            card := cards.Front().Value.(*Card)
            jsonBytes, _ := json.Marshal(Move{ *card, pilenum })
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
