package main;

import "net"
import "fmt"
import "os"
import "strings"
import "strconv"
import "encoding/json"

func handle_incomming_connection(connection net.Conn){
    defer connection.Close()
    fmt.Printf("%s->%s:", connection.RemoteAddr(), connection.LocalAddr())

    // Read firs 32 incomming bytes, which should be a padded ascii string of the length of the json data to be gotten later
    json_len_string_bytes:=make([]byte, 32)
    n,err:=connection.Read(json_len_string_bytes)
    if err!=nil{
        fmt.Fprintln(os.Stderr, "Error (connection.Read) (1):", err)
        return
    }

    // Get the actual length out of the padded string
    json_len_string:=string(json_len_string_bytes)
    json_len_string=strings.TrimSpace(json_len_string)
    json_len, err:=strconv.Atoi(json_len_string)
    if err!=nil{
        fmt.Fprintln(os.Stderr, "Error (strconv.Atoi) (1):", err)
        return
    }

    // Make sure the length is within a sensible range
    if json_len<=0 || json_len>1024*1024{
        fmt.Fprintln(os.Stderr, "Error (json_len too long or too short):", json_len)
        return
    }

    // Get the actuall json data as a byte array
    json_filedirdata:=make([]byte, json_len)
    n,err=connection.Read(json_filedirdata)
    if err!=nil || n!=json_len{
        fmt.Fprintln(os.Stderr, "Error (connection.Read) (2):", err, "json_len:", json_len)
        return
    }

    // Get a go-structure out of the json data
    var file_dir_data FileDirData
    err=json.Unmarshal(json_filedirdata, &file_dir_data)
    if err!=nil{
        fmt.Fprintln(os.Stderr, "Error (json.Unmarshal):", err)
        return
    }

    fmt.Println(file_dir_data)
}

func main() {
    // Listen for connections (this being a server and all)
    listener,err:=net.Listen("tcp", ":1234")
    if err!=nil{
        fmt.Fprintln(os.Stderr, "Error (net.Listen):", err)
        return
    }
    defer listener.Close()
    fmt.Println("Listening on:", listener.Addr())

    // For each incomming connection
    for{
        connection, err:=listener.Accept()
        if err != nil {
            fmt.Fprintln(os.Stderr, "Error (listener.Accept):", err)
        }

        go handle_incomming_connection(connection)
    }
}