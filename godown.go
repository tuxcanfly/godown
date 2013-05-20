package main

import "io"
import "os"
import "fmt"
import "path"
import "strconv"
import "flag"
import "net/url"
import "net/http"
import "io/ioutil"

func download(uri string, start int, offset int, chunks chan []byte) {
    fmt.Println("Downloading range: ", start, "-", start+offset)

    client := &http.Client{}
    req, _ := http.NewRequest("GET", uri, nil)
    req.Header.Set("Range: ", fmt.Sprintf("bytes=%d-%d", start, start+offset))
    resp, err := client.Do(req)
    if err == io.EOF {
        return
    }
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    chunks <- body
}

func main() {

    download_url := flag.String("url", "", "URL to download")
    threads := flag.Int("threads", 3, "Number of threads")
    flag.Parse()

    if len(*download_url) == 0 {
        fmt.Println("No --url provided")
        os.Exit(1)
    }

    parsed_url, _ := url.Parse(*download_url)
    filename := path.Base(parsed_url.Path)

    file, err := os.Create(filename)
    if err != nil {
        panic(err)
    }

    resp, err := http.Head(*download_url)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    content_length, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
    offset := content_length / *threads
    start := 0

    chunks := make(chan []byte)

    for _ = range make([]int, *threads) {
        go download(*download_url, start, offset, chunks)
        start += offset
    }
    file.Write(<-chunks)
    fmt.Println("Download complete - saved to", filename)
}
