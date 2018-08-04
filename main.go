package main;

import "flag"


func main() {
    address:=flag.String("address", "localhost:7587", "Address of server if used as client. Address of client is used as server.")
    input:=flag.String("input", "input", "Path of files to send to the server. Only used by client.")
    output:=flag.String("output", "output", "Path to put files retrieved from server into. Only used by client.")
    server:=flag.Bool("server", false, "If set server is run. Otherwise, client is run.")
    flag.Parse()

    if *server{
        run_server(*address)
    } else {
        run_client(*address, *input, *output)
    }
}

// As root:
// apt-get install curl git build-essential software-properties-common
// wget https://dl.google.com/go/go1.10.3.linux-amd64.tar.gz
// tar zxfv go1.10.3.linux-amd64.tar.gz
// mv go /usr/local/
// export PATH=$PATH:/usr/local/go/bin
// curl -sL https://deb.nodesource.com/setup_10.x | sudo bash -
// apt-get install nodejs
// npm install -g elm --unsafe-perm=true --allow-root

// As user:
// curl https://sh.rustup.rs -sSf | sh
// rustup install nightly
// rustup default nightly

// On every start:
// export PATH=$PATH:/usr/local/go/bin