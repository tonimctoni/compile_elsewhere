package main;

import "net"
import "fmt"
import "os"
import "strconv"
import "strings"
import "path/filepath"
import "encoding/json"
import "errors"
import "io/ioutil"
import "path"

// Sends and integer through the connection. It is sent in a padded ascii format of 20 bytes
func write_int_to_connection(connection net.Conn, int_to_send int) error{
    // Get a padded string of len 20 of the int to send
    int_to_send_string:=strconv.Itoa(int_to_send)
    int_to_send_padded_string:=int_to_send_string+strings.Repeat(" ", 20-len(int_to_send_string))

    // Send it
    _,err:=connection.Write([]byte(int_to_send_padded_string))
    if err!=nil{
        fmt.Fprintln(os.Stderr, "Error (connection.Write) (1):", err)
        return err
    }

    return nil
}

// Sends bytes through connection. If there are more than 1024, they are sent in chunks of 1024
func write_bytes_to_connection(connection net.Conn, bytes_to_send []byte) error{
    // While more than 1024 bytes are left to send: send chunks of 1024 bytes
    for len(bytes_to_send)>1024{
        n,err:=connection.Write(bytes_to_send[:1024])
        if err!=nil{
            fmt.Fprintln(os.Stderr, "Error (connection.Write) (2):", err)
            return err
        }
        bytes_to_send=bytes_to_send[n:]
    }

    // Send the at most 1024 bytes that are left to send
    if len(bytes_to_send)!=0{
        n,err:=connection.Write(bytes_to_send[:])
        if err!=nil{
            fmt.Fprintln(os.Stderr, "Error (connection.Write) (3):", err)
            return err
        }
        bytes_to_send=bytes_to_send[n:]
    }

    if len(bytes_to_send)!=0{
        panic("write_bytes_to_connection implemented incorrectly")
    }

    return nil
}

func write_struct_as_json_to_connection(connection net.Conn, struct_to_send interface{}) error{
    // Get the struct as json data
    struct_to_send_as_json, err:=json.Marshal(struct_to_send)
    if err != nil {
        fmt.Fprintln(os.Stderr, "Error (json.Marshal):", err)
        return err
    }

    // Send length of the json data to be sent
    err=write_int_to_connection(connection, len(struct_to_send_as_json))
    if err!=nil{
        return err
    }

    // Send the actual json data
    err=write_bytes_to_connection(connection, struct_to_send_as_json)
    if err!=nil{
        return err
    }

    return nil
}


func list_dir_into_file_dir_data(root string) (FileDirData, error){
    file_dir_data:=FileDirData{}

    // For each file and directory in "root":
    err:=filepath.Walk(root, func(path string, info os.FileInfo, err error) error{
        // Make sure it is not the root itself
        if path==root{
            return nil
        }

        // Make sure root is included in path
        if !strings.HasPrefix(path, root+"/"){
            fmt.Fprintln(os.Stderr, "Error (filepath.Walk) (HasPrefix):", "path does not have 'root'/ as prefix")
            return errors.New("Error (filepath.Walk) (HasPrefix): path does not have 'root'/ as prefix")
        }

        // Remove root from path
        path=path[len(root)+1:]

        // Make sure there were no errors reaching this file/directory listing its contents
        if err != nil {
            fmt.Fprintln(os.Stderr, "Error (filepath.Walk) (WalkFunc):", err)
            return nil
        }

        // If it is a dir, add its path to a slice to be returnet
        // If it is a file, add its path and size to a slice to be returned
        if info.IsDir(){
            file_dir_data.Dirs=append(file_dir_data.Dirs, path)
        } else{
            file_dir_data.Files=append(file_dir_data.Files, FileData{path, int(info.Size())})
        }

        return nil
    })

    // Make sure files and directories were listed correctly
    if err != nil {
        return FileDirData{}, err
    }

    return file_dir_data, nil
}


func main() {
    input_dir:="input"
    // Get a list of all files and directories in "./input"
    file_dir_data, err:=list_dir_into_file_dir_data(input_dir)
    if err != nil {
        fmt.Fprintln(os.Stderr, "Error (list_dir_as_json):", err)
        return
    }

    // Create a tcp connection with server
    connection,err:=net.Dial("tcp", "localhost:1234")
    if err!=nil{
        fmt.Fprintln(os.Stderr, "Error (net.Dial):", err)
        return
    }
    defer connection.Close()

    // Write the list of relevant files and directories to connection
    err=write_struct_as_json_to_connection(connection, file_dir_data)
    if err!=nil{
        return
    }

    // Write the actual files to connection
    for _, file_data:=range file_dir_data.Files{
        file_content, err:=ioutil.ReadFile(path.Join(input_dir, file_data.Path))
        if err!=nil{
            fmt.Fprintln(os.Stderr, "Error (ioutil.ReadFile):", err)
            return
        }

        err=write_bytes_to_connection(connection, file_content)
        if err!=nil{
            return
        }
    }
}