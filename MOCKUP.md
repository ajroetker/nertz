#Commands

```bash
help  
  Your commands are these:
  draw: reveals the next three cards in your stream
  move <thiscardname> <thatcardname>: moves this card under that card. Both cards are assumed to be in your hand.
    eg: move 4h 5s; move 9c Td; move Qc Kh;
  fish: fills an empty space in your river with the top card of your nertz pile
  lake <thiscardname> <pilenumber>: moves this card from your hand to the specified pile in the lake
    eg: move As 1; move 2s 1; move 3s 1; move Ad 2;

draw  

move Qh Ks  
     (works)  
move 4s 5h  
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
lake Ac 5
     (should probably just put Ac on 4)


quit
     Error: quitters never win
```

##Examples

```bash
lake:      1:  [ As[...[ Js ]  
           2:  [ Ah[...[ 6h ]  
           3:  [ Ad[ Kd[ Qd ]  
           4:  [-]  
(arenas have variable length)  

river:     1:  [ Jh[ Ts[ 9h[ 8c[ 7d[ 6s[ 5h ]  
           2:  [ 9s[ 8d[ 7h ]  
           3:  [ 4s[ 3h ]  
           4:  [ - ]  

stream:        [[[[[ Ad[ 2s[ Kc]  
         (or)  [...[ Ad[ 2s[ Ac]  


nertz pile:    [[[[[[[[[[[[[[[[[[[[[ Qh ]  
         (or)  [11 cards[ Qh ]  
         (or)  [[[[[[...[ Qh ]  
```
