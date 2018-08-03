package main;

import "net"
import "fmt"
import "os"
import "strconv"
import "strings"
import "path/filepath"
import "encoding/json"
import "errors"


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
    // Get a list of all files and directories in "./input"
    file_data, err:=list_dir_into_file_dir_data("input")
    if err != nil {
        fmt.Fprintln(os.Stderr, "Error (list_dir_as_json):", err)
        return
    }

    // Get the list of files and directories in json format
    json_file_dir_data, err:=json.Marshal(file_data)
    if err != nil {
        fmt.Fprintln(os.Stderr, "Error (json.Marshal):", err)
        return
    }

    // Get a padded string of len 20 of the length of the list of files and directories in json format
    json_len_string:=strconv.Itoa(len(json_file_dir_data))
    json_len_string+=strings.Repeat(" ", 20-len(json_len_string))

    // Create a tcp connection with server
    connection,err:=net.Dial("tcp", "localhost:1234")
    if err!=nil{
        fmt.Fprintln(os.Stderr, "Error (net.Dial):", err)
        return
    }
    defer connection.Close()

    // Send length of the json data to be sent
    n,err:=connection.Write([]byte(json_len_string))
    if err!=nil || n!=len(json_len_string){
        fmt.Fprintln(os.Stderr, "Error (connection.Write) (1):", err)
        return
    }

    // Actually send the json data
    n,err=connection.Write(json_file_dir_data)
    if err!=nil || n!=len(json_file_dir_data){
        fmt.Fprintln(os.Stderr, "Error (connection.Write) (2):", err)
        return
    }
}