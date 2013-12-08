nertz
=====

The rules for [nertz](http://en.wikipedia.org/wiki/Nertz "Link to Wikipedia Description of the Game") are simple.  

This repository is a Go implementation of nertz using websockets for client connections. All data is communicated with JSON over the socket is in JSON and the `/move` HTTP POST requests so there is no reason the client of the server couldn't be rewritten in another language.  

to do
=====

- Quiting/Game Over  
- Terminal client code  
- Server handling Multiple Games  
- Credentials handling/Logging in  
- Browser client/Gui client
