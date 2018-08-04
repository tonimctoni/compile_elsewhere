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