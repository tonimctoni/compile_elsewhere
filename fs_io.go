package main;

import "path"
import "path/filepath"
import "os"
import "fmt"
import "strings"
import "errors"

type FileData struct{
    Path string `json: "path"`
    Size int `json: "size"`
}

type FileDirData struct{
    Files []FileData `json: "files"`
    Dirs []string `json: "dirs"`
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
        fmt.Fprintln(os.Stderr, "Error (filepath.Walk):", err)
        return FileDirData{}, err
    }

    return file_dir_data, nil
}