##Commands

```bash
help  
  Your commands are these:
  `ready`: for when you want to play!
  `draw`: reveals the next three cards in your stream
  `move <from> <to>`: moves the from card onto the to card.
    e.g.: move 4h 5s | move 9c Td | move Qc Kh
  `fish`: moves top card of your nertz pile to an empty space in the river
  `lake <card> <pile>`: moves card from hand to the specified pile in the lake
    e.g.: lake As 1 | lake 2s 1 | lake 3s 1 | lake Ad 2

ready  
    (tell the server you are ready to play)

draw  
    (self-explanatory)

move Qh Ks  
     (works)  
move 4s Ks  
     Error: not a legal move  
move 2s 3h  
     Error: 2s is not an available card  

fish  
     (fills an empty space in the river with
      the top of the nertz pile)
fish  
     Error: no empty spaces on the river

lake 7h 2
     (works)
lake 7h 1
     Error: not a legal move
lake Ad 4
     Error: Ad is not an available card
lake Ac 4
     (works)

quit
     Error: quitters never win
```

##Examples

```bash
Lake: [ J s ] [ 6 h ] [ Q d ] [ ] ...  

River: [ J h [ T s [ 9 h [ 8 c [ 7 d [ 6 s [ 5 h ]  
       [ 9 s [ 8 d [ 7 h ]  
       [ 4 s [ 3 h ]  
       [ ]  

Stream: [[[[[ A d [ 2 s [ K c ]  
Streampile: [ ]]]]]]]]]]]]]]]  

Nertzpile:    [[[[[[[[ Q h ]  
```
