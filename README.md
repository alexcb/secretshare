Share secrets using rsa public/private encryption.

Usage:

Alice has a secret message to send to Bob. Alice pings bob on slack and says hey bob I need to send you a secret.
Bob has a ssh key installed, but can't remember how to use gpg on the command line. Instead bob uses this sharesecret tool.

Bob opens a terminal and runs:

    ./build/linux/amd64/secretshare

It prints out the following message:

    To decrypt data, run: ./build/linux/amd64/secretshare decrypt < file_to_decrypt
    To encrypt data, run: ./build/linux/amd64/secretshare <encryption_key> < data_to_encrypt
    
    For example if someone wanted to send you data, they would run:
    ./build/linux/amd64/secretshare AAAAB3NzaC1yc2EAAAADAQABAAABAQCkSAXQqyy+99uJGaYy6dBVNITdTrYeNewigGhx6/SrPppJX7KLPo6qSI8vP/ej8VDiFJGB4FjbiCLarkn1X1e1F4GW7CkjylUmD1X7njl6EeuZSzqvzsWoyO3Pgwa94d/mkQNvfvGyC9FopJh0pdVbLcPuyX75Tc6SmD8jq9PifoyC3nX2qeUOSZMgjbADpsIGABENaaDs1gTeRp2KwYHG2UwxnAUNKoANFIUK1McAL37xSJJ32pY4vEtlYxzhu2Rji7fUvQB4gqWhKuoOOoP1aP4zcOSPORMyZyPOPLT3SiVnW4GI10j0p73Y/aoYeg0eRUvhKB8WDRwOXIldgWrv

Bob replies to Alice and tells her she can send him data using:

    ./build/linux/amd64/secretshare AAAAB3NzaC1yc2EAAAADAQABAAABAQCkSAXQqyy+99uJGaYy6dBVNITdTrYeNewigGhx6/SrPppJX7KLPo6qSI8vP/ej8VDiFJGB4FjbiCLarkn1X1e1F4GW7CkjylUmD1X7njl6EeuZSzqvzsWoyO3Pgwa94d/mkQNvfvGyC9FopJh0pdVbLcPuyX75Tc6SmD8jq9PifoyC3nX2qeUOSZMgjbADpsIGABENaaDs1gTeRp2KwYHG2UwxnAUNKoANFIUK1McAL37xSJJ32pY4vEtlYxzhu2Rji7fUvQB4gqWhKuoOOoP1aP4zcOSPORMyZyPOPLT3SiVnW4GI10j0p73Y/aoYeg0eRUvhKB8WDRwOXIldgWrv

Alice then runs the command:

    echo 'Here's your new password. Dont share it with anyone. I hope nobody decrypts this message. anyway it's "querty". Have fun.' | ./build/linux/amd64/secretshare AAAAB3NzaC1yc2EAAAADAQABAAABAQCkSAXQqyy+99uJGaYy6dBVNITdTrYeNewigGhx6/SrPppJX7KLPo6qSI8vP/ej8VDiFJGB4FjbiCLarkn1X1e1F4GW7CkjylUmD1X7njl6EeuZSzqvzsWoyO3Pgwa94d/mkQNvfvGyC9FopJh0pdVbLcPuyX75Tc6SmD8jq9PifoyC3nX2qeUOSZMgjbADpsIGABENaaDs1gTeRp2KwYHG2UwxnAUNKoANFIUK1McAL37xSJJ32pY4vEtlYxzhu2Rji7fUvQB4gqWhKuoOOoP1aP4zcOSPORMyZyPOPLT3SiVnW4GI10j0p73Y/aoYeg0eRUvhKB8WDRwOXIldgWrv

Which outputs:

    DpNw53fz1c2TUVwXTO/K2w5CYPvARTVzjzQb/9Rdt8+lW7tfFy8t9BWqdwN2vq7lwMjkqmKxJIKDhc8OwMh+TkMrmC4Q0Qr146td9E9vu+XV5LyBD7OlYXqFo0edE9QmyrRCcG9teV3XuLpmO7XgnenYnyOTySepjHX1rZgff6VVGn1dWtHLKk32H6D39q+HkY+8k+cTUgXwWe3rJTcjipXJAsDmH9/DoPJcCwH/Rc/9mz+zJW9YEiASuxl37e639erSFY3JwTTTOfnBzOoS4pWJPQtpCNdx77RXlSbNPyx+CgBMLpx5/QgqyLcBe0mWf2pA17eYHN3Z6gDt3Wsavg==

She sends this encrypted data to Bob, who runs the following command to decrypt it:


    echo "DpNw53fz1c2TUVwXTO/K2w5CYPvARTVzjzQb/9Rdt8+lW7tfFy8t9BWqdwN2vq7lwMjkqmKxJIKDhc8OwMh+TkMrmC4Q0Qr146td9E9vu+XV5LyBD7OlYXqFo0edE9QmyrRCcG9teV3XuLpmO7XgnenYnyOTySepjHX1rZgff6VVGn1dWtHLKk32H6D39q+HkY+8k+cTUgXwWe3rJTcjipXJAsDmH9/DoPJcCwH/Rc/9mz+zJW9YEiASuxl37e639erSFY3JwTTTOfnBzOoS4pWJPQtpCNdx77RXlSbNPyx+CgBMLpx5/QgqyLcBe0mWf2pA17eYHN3Z6gDt3Wsavg==" | ./build/linux/amd64/secretshare decrypt

Which outputs:

    Here's your new password. Dont share it with anyone. I hope nobody decrypts this message. anyway it's "querty". Have fun.
