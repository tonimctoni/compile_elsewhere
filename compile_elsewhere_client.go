package main;

import "net"
import "fmt"
import "os"


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

    // Get the stdout and stderr output of the compilation process
    cmd_stdout, cmd_stderr, err:=func() ([]byte, []byte, error){
        size, err:=read_int_from_connection(connection)
        if err!=nil{
            return nil, nil, err
        }

        cmd_stdout, err:=read_bytes_from_connection(connection, size)
        if err!=nil{
            return nil, nil, err
        }

        size, err=read_int_from_connection(connection)
        if err!=nil{
            return nil, nil, err
        }

        cmd_stderr, err:=read_bytes_from_connection(connection, size)
        if err!=nil{
            return nil, nil, err
        }

        return cmd_stdout, cmd_stderr, nil
    }()
    if err!=nil{
        return
    }
    fmt.Printf("Stdout:\n%s\n\n\nStderr:\n%s\n\n\n", cmd_stdout, cmd_stderr)

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

    fmt.Println("Successful exit")
}