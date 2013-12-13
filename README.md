nertz
=====

The rules for [nertz](http://en.wikipedia.org/wiki/Nertz "Link to Wikipedia Description of the Game") are simple.  

This repository is a Go implementation of nertz using websockets for client connections. All data is communicated with JSON over the socket is in JSON and the `/move` HTTP POST requests so there is no reason the client of the server couldn't be rewritten in another language.

current picture
===============

#Server  
- `nertz-server` takes a `<port>` to start the service on  
- accepts websocket connections at `:<port>/` until the game has started (at most 6 players)  
    * asks for client credentials, sending JSON `{"Message" : "Credentials"}`  
    * client must provide Credentials as JSON (password does nothing at the moment)  

```go  
type Credentials struct {  
  Username string  
  Password string  
}
```

- Not sure the best way to start a game at the moment
- Set `nertz.Game.Started := true`
- another handler listening at `:<port>/move"` takes POST requests with JSON bodies that gets parsed to a move struct

```go  
type Card struct {  
  Value int
  Suit int
  Player string  
}

type Move struct {  
  Card *Card  
  Pile int  
}
```

- When a user is finished they should send a JSON encoding of their Hand (inclding if they just wish to quit)
- A Hand right now should be implemented using linked lists for the piles

```go  
type Hand struct {  
    Nertzpile *list.List  
    Streampile *list.List  
    River []*list.List  
    Stream *list.List  
}  
```

- Determine if this was a quit or a game over
    * Sends the user their score if it was a quit
    * Broadcasts a game over if nertz
- When the game is over
    * Users are expected to send in their hands
    * Calculate their scores and update the tallied up scoreboard
    * Send the users the scoreboard which will be a JSON encoded `map[string]int`
    * Close the websocket connections

to do
=====

There is a lot to do on the server side still, making it safe, printing error messages, closing connections and better engineering stuff.

__Server__
- [ ] Game Begin Interaction  
- [x] Game Over Interaction  
- [ ] Quiting  
- [ ] Handling Multiple Games  
- [ ] Handling variable numbers of clients  
- [ ] Database for Credentials/ High scores tracking

__Client__
- [ ] Client Nertz Code
- [ ] Terminal display code (a nice implementation might resemble a [progress bar](http://www.darkcoding.net/software/pretty-command-line-console-output-on-unix-in-python-and-go-lang/ "A nice example of a GoLang progress bar")  
- [ ] Credentials handling/Logging in  
- [ ] Browser client which contacts Server
- [ ] Browser display/Gui display  
