arena:     1:  [ As[...[ Js ]
           2:  [ Ah[...[ 6h ]
           3:  [ Ad[ Kd[ Qd ]
           4:  [-]
(arenas have variable length)

river:     1:  [ Jh[ Ts[ 9h[ 8c[ 7d[ 6s[ 5h ]
           2:  [ 9s[ 8d[ 7h ]
           3:  [ 4s[ 3h ]
           4:  [ Ks ]

stream:        [[[[[ Ad[ 2s[ Ac]
	 (or)  [...[ Ad[ 2s[ Ac]


nertz pile:    [[[[[[[[[[[[[[[[[[[[[ Qh ]
         (or)  [11 cards[ Qh ]
         (or)  [[[[[[...[ Qh ]





~~~~-+-+-+\*\*COMMANDS*/*/+-+-+-~~~~

move Qh Ks
     (works)
move 4s 5h
     (works)
move 4s Ks
     Error: not a legal move
move 2s 3h
     Error: 2s is not an available card


arena 7h 2
     (works)
arena 7h 1
     Error: not a legal move
arena Ad 4
     Error: Ad is not an available card
arena Ac 4
     (works)
arena Ac 5
     (should probably just put Ac on 4)

draw

quit
     Error: quitters never win
