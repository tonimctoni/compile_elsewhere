package main;

import "net"
import "fmt"
import "os"
import "strconv"
import "sync/atomic"
import "os/exec"
import "bytes"


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
    err=read_from_connection_into_files(connection, root_dir, file_dir_data.Files, false)
    if err!=nil{
        return
    }

    // // Make a set of all existing files, to only send new ones later
    // retrieved_file_path_set:=make(map[string]bool)
    // for _,file_data:=range file_dir_data.Files{
    //     retrieved_file_path_set[file_data.Path]=true
    // }

    // Compile the received source code
    cmd:=exec.Command("make")
    cmd_stdout:=new(bytes.Buffer)
    cmd_stderr:=new(bytes.Buffer)
    cmd.Stdout=cmd_stdout
    cmd.Stderr=cmd_stderr
    cmd.Dir=root_dir
    cmd_err:=cmd.Run()

    // Write stdout/stderr of make command to connection
    err=write_int_to_connection(connection, cmd_stdout.Len())
    if err!=nil{
        return
    }
    err=write_bytes_to_connection(connection, cmd_stdout.Bytes())
    if err!=nil{
        return
    }
    err=write_int_to_connection(connection, cmd_stderr.Len())
    if err!=nil{
        return
    }
    err=write_bytes_to_connection(connection, cmd_stderr.Bytes())
    if err!=nil{
        return
    }

    // If there was an error in the compilation process: abort
    if cmd_err!=nil{
        fmt.Fprintln(os.Stderr, "Error (exec.Command(...).Run()):", err)
        return
    }

    // Get a list of all files and directories pertinent to the current user
    file_dir_data, err=list_dir_into_file_dir_data(root_dir)
    if err != nil {
        return
    }

    // // Remove old files from list, such that only ones created when compiling are sent to client
    // new_file_data:=make([]FileData, 0, len(file_dir_data.Files))
    // for _,file_data:=range file_dir_data.Files{
    //     if retrieved_file_path_set[file_data.Path]==false{
    //         new_file_data=append(new_file_data, file_data)
    //     }
    // }
    // file_dir_data.Files=new_file_data

    // Write it to connection
    err=write_struct_as_json_to_connection(connection, file_dir_data)
    if err!=nil{
        return
    }

    // Write the actual files to connection
    err=write_files_into_connection(connection, root_dir, file_dir_data.Files, false)
    if err!=nil{
        return
    }
}

func run_server(address string) {
    // address:=":1234"
    work_dir_counter:=int64(0)

    // Listen for connections (this being a server and all)
    listener,err:=net.Listen("tcp", address)
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