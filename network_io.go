package main;

import "net"
import "os"
import "strings"
import "fmt"
import "strconv"
import "encoding/json"
import "io/ioutil"
import "path"

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

func read_from_connection_into_files(connection net.Conn, root_dir string, file_data []FileData, verbose bool) error{
    // Read all needed files from connection and save them in their paths within root_dir
    for _,file_data:=range file_data{
        if verbose{
            fmt.Println("Receiving:", file_data.Path)
        }
        file_content, err:=read_bytes_from_connection(connection, file_data.Size)
        if err!=nil{
            return err
        }

        err=ioutil.WriteFile(path.Join(root_dir, file_data.Path), file_content, 0766)
        if err!=nil{
            fmt.Fprintln(os.Stderr, "Error (ioutil.WriteFile):", err)
            return err
        }
    }

    return nil
}

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

func write_files_into_connection(connection net.Conn, root_dir string, file_data []FileData, verbose bool) error{
    // Write files in file_data to connection
    for _, file_data:=range file_data{
        if verbose{
            fmt.Println("Sending:", file_data.Path)
        }
        file_content, err:=ioutil.ReadFile(path.Join(root_dir, file_data.Path))
        if err!=nil{
            fmt.Fprintln(os.Stderr, "Error (ioutil.ReadFile):", err)
            return err
        }

        err=write_bytes_to_connection(connection, file_content)
        if err!=nil{
            return err
        }
    }

    return nil
}