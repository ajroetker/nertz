nertz
=====

The rules for [nertz](http://en.wikipedia.org/wiki/Nertz "Link to Wikipedia Description of the Game") are simple.  

This repository is a Go implementation of nertz using websockets for client connections. All data is communicated with JSON over the socket is in JSON and the `/move` HTTP POST requests so there is no reason the client of the server couldn't be rewritten in another language.

###current picture

####Server  
- `nertz-server` takes a `<port>` to start the service on  
- accepts websocket connections at `:<port>/` until the game has started (at most 6 players)  
    * asks for client credentials, sending JSON `{ "Message" : "Credentials" }`  
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

- When a user is finished they should send JSON similar to `{"Value" : 4 , "Nertz" : true }`
    * "Value" here is determined by the size of their Nertz pile
- Determine if this was a quit or a game over according to "Nertz
    * Sends the user their score if it was a quit
    * Broadcasts a game over if nertz
- When the game is over
    * A `{ "Message" : "Nertz" }` is broadcast
    * Users are expected to send in their hands
    * Calculate their scores and update the tallied up scoreboard
    * Send the users the scoreboard which will be a JSON encoded `map[string]int`
    * Close the websocket connections

####Client  
- `nertz-client` takes a `<host>` and `<port>` to connect to  
- Wait for other clients to connect to the server  
- Once you're friends are connected type `ready` to block new clients from joining but wait for connected clients to `ready` as well
- a sample client specification in the terminal in shown in the accompaning README.md  
- the client can pickup linked lists of cards and move them to the other piles  
    * this implementation was chosen specifically to be extended easily to a GUI or browser  
- wait for a "Game over"  message from the sever and then count the nertz pile and send it back as a response  

###to do

There is a lot to do on the server side still, making it safe, printing error messages, closing connections and better engineering stuff.

__Server__
- [ ] Game Begin Interaction  
- [x] Game Over Interaction  
- [x] Quiting  
- [x] Accept clients and broadcast messages  
- [x] HTTP and Websocket Handlers  
- [ ] Handling Multiple Games  
- [ ] Handling variable numbers of clients  
- [ ] Database for Credentials/ High scores tracking

__Client__
- [ ] Display the Arena
- [x] Display the Hand
- [ ] Display the Messages
- [ ] Client Nertz Code
- [ ] Terminal display code (a nice implementation might resemble a [progress bar](http://www.darkcoding.net/software/pretty-command-line-console-output-on-unix-in-python-and-go-lang/ "A nice example of a GoLang progress bar")  
- [ ] Credentials handling/Logging in  
- [ ] Browser client which contacts Server
- [ ] Browser display/Gui display  
