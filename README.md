# Share secrets using rsa public/private encryption.

## Installation

### Linux

    sudo /bin/sh -c 'wget https://github.com/alexcb/secretshare/releases/latest/download/secretshare-linux-amd64 -O /usr/local/bin/secretshare && chmod +x /usr/local/bin/secretshare'

### MacOS

    sudo /bin/sh -c 'wget https://github.com/alexcb/secretshare/releases/latest/download/secretshare-darwin-amd64 -O /usr/local/bin/secretshare && chmod +x /usr/local/bin/secretshare'

### Usage

Suppose you want to receive some data from Alice, first you must run `secretshare` locally:

    secretshare

This will print out something like:

    secretshare AAAAB3NzaC1yc2EAAAADAQABAAABAQCagdQYuPXgzyfg2k58CYTntkSo2rI9QtZWYnk45n6oZGW8wGUoqrmCeLCbCo7HG6JJD+jUuWx9ELmC2rRINQnhJdBdwbMlOx8v7oSr60xR8b0pYRL1gm6DseU0u/pClBfGcwV7tsxuRspk/c3/cpYOFcs2vN5mXo29qSpg1w4iE3snoAOQGajN1U4sT3rht5hjx188d9MxuBTnqd0yW7RQMSS8YqdszaAiwUQ2rtwMKDUp5CT+5LZ/QW/EMexRSEfgkjGGmz+NlBAhGDTrBByw2Dl+6pZESAtxmCaZ6vzG+CSqH5oQVXVl/A9YEQiuDVGdxcfF3jYmlYIfISeMtRTj < data_to_send

This output can safely be sent to alice. Alice will then run something like:

    echo 'Hi Bob, your new password is "querty".' | secretshare AAAAB3NzaC1yc2EAAAADAQABAAABAQCagdQYuPXgzyfg2k58CYTntkSo2rI9QtZWYnk45n6oZGW8wGUoqrmCeLCbCo7HG6JJD+jUuWx9ELmC2rRINQnhJdBdwbMlOx8v7oSr60xR8b0pYRL1gm6DseU0u/pClBfGcwV7tsxuRspk/c3/cpYOFcs2vN5mXo29qSpg1w4iE3snoAOQGajN1U4sT3rht5hjx188d9MxuBTnqd0yW7RQMSS8YqdszaAiwUQ2rtwMKDUp5CT+5LZ/QW/EMexRSEfgkjGGmz+NlBAhGDTrBByw2Dl+6pZESAtxmCaZ6vzG+CSqH5oQVXVl/A9YEQiuDVGdxcfF3jYmlYIfISeMtRTj

This will output something like

    gevonaUXKj7Wmct65A8yVJWTp9D/sd6YbOCi4BtrKGeUjdWs/fa0BbP0IQhdP2j4fS6n12zGmgkOXHLJyJnG9OPxZ+EaPcIfVBs5TfvNnC/8Dfu+V5ScRIYXBHVjRsDOBQCYzeOSwFvq1vUyuq20Wr7s3szbgFkDttxPsaXMKyxTcVEqkgSp09dhV7roqBmsRUDbAFpIWLIUb4ZAtCfU6rbWaAes9acSmMT3fvW/no1gsa3/Wobdpj3T7WVrQsj+upr2ANlFyA3Bt7IOKxmJhrrRYOBxAkk6NEnYmrHWR26KGRhz/VRPxAZWsB/qMoVAw5ukjnVple2+x8SMrIE9Gg==

Alice can then safely send this data back to you, and only you can decrypt it, but running this command:

    echo "gevonaUXKj7Wmct65A8yVJWTp9D/sd6YbOCi4BtrKGeUjdWs/fa0BbP0IQhdP2j4fS6n12zGmgkOXHLJyJnG9OPxZ+EaPcIfVBs5TfvNnC/8Dfu+V5ScRIYXBHVjRsDOBQCYzeOSwFvq1vUyuq20Wr7s3szbgFkDttxPsaXMKyxTcVEqkgSp09dhV7roqBmsRUDbAFpIWLIUb4ZAtCfU6rbWaAes9acSmMT3fvW/no1gsa3/Wobdpj3T7WVrQsj+upr2ANlFyA3Bt7IOKxmJhrrRYOBxAkk6NEnYmrHWR26KGRhz/VRPxAZWsB/qMoVAw5ukjnVple2+x8SMrIE9Gg==" | secretshare decrypt

Which will output the original message:

    Hi Bob, your new password is "querty".

## Under the hood

Under the hood `secretshare` will generate a private and public key stored under ~/.secretshare and ~/.secretshare.pub respectively. The public key is what is used by alice to encrypt data that can only be decrypted using the private key. If we `cat ~/.secretshare.pub`, we can see the same data is sent in the above usage example:

    ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCagdQYuPXgzyfg2k58CYTntkSo2rI9QtZWYnk45n6oZGW8wGUoqrmCeLCbCo7HG6JJD+jUuWx9ELmC2rRINQnhJdBdwbMlOx8v7oSr60xR8b0pYRL1gm6DseU0u/pClBfGcwV7tsxuRspk/c3/cpYOFcs2vN5mXo29qSpg1w4iE3snoAOQGajN1U4sT3rht5hjx188d9MxuBTnqd0yW7RQMSS8YqdszaAiwUQ2rtwMKDUp5CT+5LZ/QW/EMexRSEfgkjGGmz+NlBAhGDTrBByw2Dl+6pZESAtxmCaZ6vzG+CSqH5oQVXVl/A9YEQiuDVGdxcfF3jYmlYIfISeMtRTj


## Building from Source

secretshare makes use of the [Earthly](https://www.earthly.dev/) build system. To build from source,
first download the [earth](https://github.com/earthly/earthly) command, then run:

    earth +secretshare-all

which will produce binary files under `./build/<platform>/amd64/secretshare`

### Release

To release a new version, run the following command:

    RELEASE_TAG=v0.0.2 earth --build-arg RELEASE_TAG --secret GITHUB_TOKEN --push +release

### Preview

[![asciicast](preview.gif)](https://asciinema.org/a/369195?&speed=2)

