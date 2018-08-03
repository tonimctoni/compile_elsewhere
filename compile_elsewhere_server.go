package main;

import "net"
import "fmt"
import "os"
import "strings"
import "strconv"
import "encoding/json"
import "errors"
import "sync/atomic"
import "path"

func get_file_dir_data_from_connection(connection net.Conn) (FileDirData, error){
    // Read firs 32 incomming bytes, which should be a padded ascii string of the length of the json data to be gotten later
    json_len_string_bytes:=make([]byte, 32)
    _,err:=connection.Read(json_len_string_bytes)
    if err!=nil{
        fmt.Fprintln(os.Stderr, "Error (connection.Read) (1):", err)
        return FileDirData{}, err
    }

    // Get the actual length out of the padded string
    json_len_string:=string(json_len_string_bytes)
    json_len_string=strings.TrimSpace(json_len_string)
    json_len, err:=strconv.Atoi(json_len_string)
    if err!=nil{
        fmt.Fprintln(os.Stderr, "Error (strconv.Atoi) (1):", err)
        return FileDirData{}, err
    }

    // Make sure the length is within a sensible range
    if json_len<=0 || json_len>1024*1024{
        fmt.Fprintln(os.Stderr, "Error (json_len too long or too short):", json_len)
        return FileDirData{}, errors.New("json_len out of sensible range")
    }

    // Get the actuall json data as a byte array
    json_filedirdata:=make([]byte, json_len)
    _,err=connection.Read(json_filedirdata)
    if err!=nil{
        fmt.Fprintln(os.Stderr, "Error (connection.Read) (2):", err, "json_len:", json_len)
        return FileDirData{}, err
    }

    // Get a go-structure out of the json data
    var file_dir_data FileDirData
    err=json.Unmarshal(json_filedirdata, &file_dir_data)
    if err!=nil{
        fmt.Fprintln(os.Stderr, "Error (json.Unmarshal):", err)
        return FileDirData{}, err
    }

    return file_dir_data, nil
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
    file_dir_data, err:=get_file_dir_data_from_connection(connection)
    if err!=nil{
        return
    }

    root_dir:="work_dir_"+strconv.Itoa(int(atomic.AddInt64(work_dir_counter, 1)))
    err=create_directories(root_dir, file_dir_data.Dirs)
    if err!=nil{
        return
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