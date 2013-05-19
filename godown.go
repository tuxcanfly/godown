package main

import "os"
import "fmt"
import "path"
import "strconv"
import "flag"
import "net/url"
import "net/http"
import "io/ioutil"

func download(uri string, chunks chan int, offset int, file *os.File) {
    for current := range chunks {

        fmt.Println("Downloading range: ", current, "-", current+offset)

        client := &http.Client{}
        req, _ := http.NewRequest("GET", uri, nil)
        req.Header.Set("Range: ", fmt.Sprintf("bytes=%d-%d", current, current+offset))
        resp, err := client.Do(req)
        if err != nil {
            panic(err)
        }
        defer resp.Body.Close()
        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            panic(err)
        }
        file.Write(body)
    }
}

func main() {

    download_url := flag.String("url", "", "URL to download")
    threads := flag.Int("threads", 5, "Number of threads")
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
    current := 0

    chunks := make(chan int)

    go download(*download_url, chunks, offset, file)

    for _ = range make([]int, *threads) {
        chunks <- current
        current += offset
    }
    fmt.Println("Download complete - saved to", filename)
}
