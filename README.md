Share secrets using rsa public/private encryption.

Usage:

Alice has a secret message to send to Bob. Alice pings bob on slack and says hey bob I need to send you a secret.
Bob has a ssh key installed, but can't remember how to use gpg on the command line. Instead bob uses this sharesecret tool.

Bob opens a terminal and runs:

    ./build/linux/amd64/secretshare

It prints out the following message:

    run this command to encrypt data: ./build/linux/amd64/secretshare AAAAB3NzaC1yc2EAAAADAQABAAABAQCwuaFN5PT7fd67c/FdPtjxmzN4ZpoT3nq832zpizoOJAJBSTRTIUltjeh2PTAlml9nKodPODvZxTw88w695vdoB7mpLUfCKXKKtI/9DV8Ay7NJKoiQK9hwofgaPXBfSMfp5veMEw8iD3OKqkFdBbYryxwUhhGICktUuUaQHpQVVuHNjlbRoVBxnb8mp4apP2B627ARBCqyRXz00pY2u9zk96GLbpaNvE7ON50G0BG4qb9iy+09hWtMC8unncSw4AT+/Oihwe1/sVdvQunWj/27rnPHbrzCplkUO+59HXYhE/LUypbzSGGVyAOsBVeWcdg3v1oXmb9LaaXh8HDb2Od7

Then waits for bob to enter an encypted message via the prompt:

    enter encrypted data

Bob carefully leaves his terminal running (if he were to quit it, he would no longer be able to decode the data).
Bob sends the message to alice over slack:

    Hey Alice, you can send me a secret using secretshare by running it like:
    ./build/linux/amd64/secretshare AAAAB3NzaC1yc2EAAAADAQABAAABAQCwuaFN5PT7fd67c/FdPtjxmzN4ZpoT3nq832zpizoOJAJBSTRTIUltjeh2PTAlml9nKodPODvZxTw88w695vdoB7mpLUfCKXKKtI/9DV8Ay7NJKoiQK9hwofgaPXBfSMfp5veMEw8iD3OKqkFdBbYryxwUhhGICktUuUaQHpQVVuHNjlbRoVBxnb8mp4apP2B627ARBCqyRXz00pY2u9zk96GLbpaNvE7ON50G0BG4qb9iy+09hWtMC8unncSw4AT+/Oihwe1/sVdvQunWj/27rnPHbrzCplkUO+59HXYhE/LUypbzSGGVyAOsBVeWcdg3v1oXmb9LaaXh8HDb2Od7


Alice then Runs the above command and sees:

    enter data to encrypt

She enters the following data:

    Here's your new password. Dont share it with anyone. I hope nobody decrypts this message. anyway it's "querty". Have fun.

The program then outputs:

    --- Below is the encrypted data ---
  ptIDqdJTgu6SITIsWV/fM93RgkEeEn0TzAtZKy/sAo7I5SEwJ0UjrPCbZJNPX1skRRbUIUElNWepX37KG71zJfk41/bKg5q/PLq8bL5ZiGg4+PXld4zuPGZZShUJRWcL2+2/JQBi6qsvej1E2I1mf8EGkEd/rmnlFAoqxOgxCz7Y8Uu6RgP97PTvJQNFWyJEkNnAizb+RyXZyoQaTaZnSOr85DUdKSuOYEj6RLVLmpckTK6130NycMZlK5MjTWtzWlrmy45T1YXSxyrAPpnXjcWSxD6l0qbDCnflGmHaJ6CBtBnb8aYiCDFufzbGqW11ZGZaA/2GXvDjm5WFankLsA==

She sends this data to Bob, which copies and pastes the data into his terminal which was left running the sharesecret tool:

    enter encrypted data
  ptIDqdJTgu6SITIsWV/fM93RgkEeEn0TzAtZKy/sAo7I5SEwJ0UjrPCbZJNPX1skRRbUIUElNWepX37KG71zJfk41/bKg5q/PLq8bL5ZiGg4+PXld4zuPGZZShUJRWcL2+2/JQBi6qsvej1E2I1mf8EGkEd/rmnlFAoqxOgxCz7Y8Uu6RgP97PTvJQNFWyJEkNnAizb+RyXZyoQaTaZnSOr85DUdKSuOYEj6RLVLmpckTK6130NycMZlK5MjTWtzWlrmy45T1YXSxyrAPpnXjcWSxD6l0qbDCnflGmHaJ6CBtBnb8aYiCDFufzbGqW11ZGZaA/2GXvDjm5WFankLsA==

It then prints out the decrypted message:
  
    -- start of message --
    Here's your new password. Dont share it with anyone. I hope nobody decrypts this message. anyway it's "querty". Have fun.
  
    -- end of message --
    press ctrl-c to quit, or continue decoding more messages.

And there you have it!
