package main;

import "net"
import "fmt"
import "os"
// import "strconv"
// import "strings"
// import "path/filepath"
// import "encoding/json"
// import "errors"
// import "io/ioutil"
// import "path"


func main() {
    input_dir:="input"
    output_dir:="output"

    // Create a tcp connection with server
    connection,err:=net.Dial("tcp", "localhost:1234")
    if err!=nil{
        fmt.Fprintln(os.Stderr, "Error (net.Dial):", err)
        return
    }
    defer connection.Close()

    // Get a list of all files and directories in "./input"
    file_dir_data, err:=list_dir_into_file_dir_data(input_dir)
    if err != nil {
        return
    }

    // Write the list of relevant files and directories to connection
    err=write_struct_as_json_to_connection(connection, file_dir_data)
    if err!=nil{
        return
    }

    // Write the actual files to connection
    err=write_files_into_connection(connection, input_dir, file_dir_data.Files)
    if err!=nil{
        return
    }

    // Receive the list of files from server (after compilation was done)
    file_dir_data=FileDirData{} //Just in case
    err=read_struct_as_json_from_connection(connection, &file_dir_data)
    if err!=nil{
        return
    }

    // Create directories to write new (compiled) files to
    err=create_directories(output_dir, file_dir_data.Dirs)
    if err!=nil{
        return
    }

    // Read contents of files from connection and write them to actual files
    err=read_from_connection_into_files(connection, output_dir, file_dir_data.Files)
    if err!=nil{
        return
    }
}