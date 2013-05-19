package main

import "os"
import "fmt"
import "path"
import "strconv"
import "flag"
import "net/url"
import "net/http"
import "io/ioutil"

func download(uri string, c chan int, offset int, out chan []byte) {
    for current := range c {

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
        out <- body
    }
}

func write(out chan []byte, f *os.File) {
    for b := range out {
        f.Write(b)
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

    f, err := os.Create(filename)
    if err != nil {
        panic(err)
    }
    out := make(chan []byte)

    resp, err := http.Head(*download_url)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    content_length, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
    offset := content_length / *threads
    current := 0

    c := make(chan int)

    go download(*download_url, c, offset, out)
    go write(out, f)

    for _ = range make([]int, *threads) {
        c <- current
        current += offset
    }
    fmt.Println("Download complete - saved to", filename)
}
