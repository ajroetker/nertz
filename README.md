nertz
=====

The rules for [nertz](http://en.wikipedia.org/wiki/Nertz "Link to Wikipedia Description of the Game") are simple.  

This repository is a Go implementation of nertz using websockets for client connections. All data is communicated with JSON over the socket is in JSON and the `/move` HTTP POST requests so there is no reason the client of the server couldn't be rewritten in another language.

current picture
===============

Server
------
- takes a `port` to start the service on
- accepts websocket connections at `"/"` until the game has started (at most 6 players)
- asks for client credentials by sending a JSON Request struct `{"Message" : "Credentials"}`
- client must provide Credentials as JSON which is parsed to (password does nothing at the moment)  

```go  
type Credentials struct {  
  Username string
  Password string  
}
```

- ?"Wait for user to ask to begin?"?
- Set `nertz.Game.Started := true`
- There is another handler listening for moves at `"/move"` which takes JSON that gets parsed to a move struct

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
- Determine if this was a quit or a game over
- Sends the user their score if it was a quit
- Wait until users have sent in their hands and update the tallied up scoreboard to reflect their nertz piles
- Send the users the scoreboard which will be a JSON encoded `map[string]int`
- close the connections



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
