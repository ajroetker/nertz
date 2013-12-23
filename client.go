package nertz

import (
    "encoding/json"
    "container/list"
    "bytes"
    "net/http"
    "errors"
    websocket "code.google.com/p/go.net/websocket"
)

/* Client-side code */

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


/**** Gameplay
 *
 * e.g. Transaction("Nertzpile", _, "Lake", _, 1)
 *
 ****/


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
