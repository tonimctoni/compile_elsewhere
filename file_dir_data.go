package main;

type FileData struct{
    Path string `json: "path"`
    Size int `json: "size"`
}

type FileDirData struct{
    Files []FileData `json: "files"`
    Dirs []string `json: "dirs"`
}