package main;

import "net"
import "fmt"
import "os"
import "strings"
import "strconv"
import "encoding/json"
// import "errors"
import "sync/atomic"
import "path"
import "io/ioutil"

// Reads 20 bytes from connection, expecting the padded ascii representation of an int. Returns the int.
func read_int_from_connection(connection net.Conn) (int, error){
    // Read 20 bytes from connection
    padded_ascii_int_bytes:=make([]byte, 20)
    _,err:=connection.Read(padded_ascii_int_bytes)
    if err!=nil{
        fmt.Fprintln(os.Stderr, "Error (connection.Read) (1):", err)
        return 0, err
    }

    // Get the actual int out of what was read
    padded_ascii_int_string:=string(padded_ascii_int_bytes)
    ascii_int_string:=strings.TrimSpace(padded_ascii_int_string)
    int_to_return, err:=strconv.Atoi(ascii_int_string)
    if err!=nil{
        fmt.Fprintln(os.Stderr, "Error (strconv.Atoi) (1):", err)
        return 0, err
    }

    return int_to_return, nil
}

// Reads bytes from connection. If the size is larger than 1024, it is read in chunks of 1024
func read_bytes_from_connection(connection net.Conn, size int) ([]byte, error){
    buffer:=make([]byte, size)
    index:=int(0)

    // While more than 1024 bytes are left to read: read them and add them to buffer
    for size>1024{
        n,err:=connection.Read(buffer[index:index+1024])
        if err!=nil{
            fmt.Fprintln(os.Stderr, "Error (connection.Read) (2):", err)
            return nil, err
        }
        index+=n
        size-=n
    }

    // If there is anything left to read (that is, size was not a multiple of 1024) read it and add it to buffer
    // Also, if size was a multiple of 1024 exactly 1024 bytes are sent here
    if size!=0{
        n,err:=connection.Read(buffer[index:])
        if err!=nil{
            fmt.Fprintln(os.Stderr, "Error (connection.Read) (3):", err)
            return nil, err
        }
        index+=n
        size-=n
    }

    if size!=0{
        panic("read_file_from_connection implemented incorrectly")
    }

    return buffer, nil
}

func read_struct_as_json_from_connection(connection net.Conn, struct_to_receive interface{}) error{
    // Get length of the json file to be read from connection
    json_struct_len, err:=read_int_from_connection(connection)
    if err!=nil{
        return err
    }

    // Get the actuall json data as a byte array
    json_struct, err:=read_bytes_from_connection(connection, json_struct_len)
    if err!=nil{
        return err
    }

    // Get struct from json data
    err=json.Unmarshal(json_struct, struct_to_receive)
    if err!=nil{
        fmt.Fprintln(os.Stderr, "Error (json.Unmarshal):", err)
        return err
    }

    return nil
}

// Create all directories in "paths" within "root" directory
func create_directories(root string, paths []string) error{
    for _,a_path:=range paths{
        a_path=path.Join(root, a_path)
        err:=os.MkdirAll(a_path, 0766)
        if err!=nil{
            fmt.Fprintln(os.Stderr, "Error (os.MkdirAll):", err)
            return err
        }
    }
    return nil
}

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

    // Read all needed files from connection and save them in their paths within root_dir
    for _,file_data:=range file_dir_data.Files{
        file_content, err:=read_bytes_from_connection(connection, file_data.Size)
        if err!=nil{
            return
        }

        err=ioutil.WriteFile(path.Join(root_dir, file_data.Path), file_content, 0766)
        if err!=nil{
            fmt.Fprintln(os.Stderr, "Error (ioutil.WriteFile):", err)
            return
        }
    }

    fmt.Println(file_dir_data)
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