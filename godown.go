package main

import "io"
import "bufio"
import "os"
import "fmt"
import "strconv"
import "flag"
import "path"
import "sync"
import "net/url"
import "net/http"
import "io/ioutil"

func main() {

    threads := flag.Int("threads", 3, "Number of threads")
    flag.Parse()

    if len(flag.Args()) == 0 {
        fmt.Println("No --url provided")
        os.Exit(1)
    }
    download_url := flag.Args()[0]

    resp, err := http.Head(download_url)
    if err != nil { panic(err) }
    defer resp.Body.Close()

    content_length, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
    offset := content_length / *threads
    remainder := content_length % (offset * *threads)
    start := 0

    parsed_url, _ := url.Parse(download_url)
    fileName := path.Base(parsed_url.Path)

    var wg sync.WaitGroup
    wg.Add(*threads)
    for i := 0; i < *threads; i++ {
        chunkName := fileName + ".part." + strconv.Itoa(i)
        go getChunk(download_url, start, offset, chunkName, &wg)
        start += offset
        if (i == *threads-2) {
            offset += remainder
        }
    }
    wg.Wait()
    outFile, err := os.Create(fileName)
    if err != nil { panic(err) }
    defer outFile.Close()
    for i := 0; i < *threads; i++ {
        chunkName := fileName + ".part." + strconv.Itoa(i)
        writeChunk(chunkName, outFile)
    }
}

func getChunk(uri string, start int, offset int, chunkName string, wg *sync.WaitGroup) {
    fmt.Println("Downloading range: ", start, "-", start+offset-1)

    client := &http.Client{}
    req, _ := http.NewRequest("GET", uri, nil)
    req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, start+offset-1))
    resp, err := client.Do(req)
    if err == io.EOF { return }
    if err != nil { panic(err) }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)

    outFile, err := os.Create(chunkName)
    if err != nil { panic(err) }
    defer outFile.Close()
    outFile.Write(body)
    wg.Done()
}

func writeChunk(chunkName string, outFile *os.File) {
    chunkFile, _  := os.Open(chunkName)
    defer chunkFile.Close()

    chunkReader := bufio.NewReader(chunkFile)
    chunkWriter := bufio.NewWriter(outFile)

    buf := make([]byte, 1024)
    for {
        n, err := chunkReader.Read(buf)
        if err != nil && err != io.EOF { panic(err) }
        if n == 0 { break }
        if _, err := chunkWriter.Write(buf[:n]); err != nil {
            panic(err)
        }
    }
    if err := chunkWriter.Flush(); err != nil { panic(err) }
    os.Remove(chunkName)
}
