package main;

import "net"
import "fmt"
import "os"
// import "strings"
import "strconv"
// import "encoding/json"
import "sync/atomic"
// import "path"
// import "io/ioutil"
import "os/exec"

func handle_incomming_connection(connection net.Conn, work_dir_counter *int64){
    defer connection.Close()
    fmt.Printf("%s->%s:\n", connection.RemoteAddr(), connection.LocalAddr())

    // Get list of files to retrieve and the directories they are in
    var file_dir_data FileDirData
    err:=read_struct_as_json_from_connection(connection, &file_dir_data)
    if err!=nil{
        return
    }

    // Create a new directory for the current user and in it put all needed subdirectories
    root_dir:="work_dir_"+strconv.Itoa(int(atomic.AddInt64(work_dir_counter, 1)))
    err=create_directories(root_dir, file_dir_data.Dirs)
    if err!=nil{
        return
    }

    // When done with the connection, remove the directory created specifically for it
    defer os.RemoveAll(root_dir)

    // Read contents of files from connection and write them to actual files
    err=read_from_connection_into_files(connection, root_dir, file_dir_data.Files)
    if err!=nil{
        return
    }

    // Compile the received source code
    cmd:=exec.Command("make")
    cmd.Dir=root_dir
    err=cmd.Run()
    if err!=nil{
        fmt.Fprintln(os.Stderr, "Error (exec.Command(...).Run()):", err)
        return
    }

    // // Get a list of all files and directories pertinent to the current user
    // file_dir_data, err=list_dir_into_file_dir_data(root_dir)
    // if err != nil {
    //     return
    // }

    // // Write it to connection
    // err=write_struct_as_json_to_connection(connection, file_dir_data)
    // if err!=nil{
    //     return
    // }

    // fmt.Println(file_dir_data)
}

func main() {
    work_dir_counter:=int64(0)

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

        go handle_incomming_connection(connection, &work_dir_counter)
    }
}