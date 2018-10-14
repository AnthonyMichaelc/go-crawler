package main

import (
    "fmt"
    "golang.org/x/net/html"
    "net/http"
    "os"
    "strings"
)
// pulls the href
func getHref(t html.Token) (ok bool, href string) {
    // iterate over all tokens until found
    for _, a := range t.Attr {
        if a.Key == "href" {
            href = a.Val
            ok = true
        }
    }
    return
}

// get all urls found under this website
func crawl(url string, ch chan string, chFinished chan bool) {
    resp, err := http.Get(url)

    defer func() {// after this function, done
        chFinished <- true
    }()
    // Error checking if the crawl false
    if err != nil {
        fmt.Println("Error: Failed to crawl \"" + url + "\"")
        return
    }

    b :=resp.Body
    defer b.Close()// close this once the function returns

    z := html.NewTokenizer(b)

    for {
        tt := z.Next()
        switch {
        case tt == html.ErrorToken:
            // document ends here
            return
        case tt == html.StartTagToken:
            t := z.Token()

            // check if this is an anchor tag
            isAnchor := t.Data == "a"
            if !isAnchor {
                continue
            }

            // extract the link value, if there is one
            ok, url := getHref(t)
            if !ok {
                continue
            }

            // start the urls with the http attribute
            hasProto := strings.Index(url, "http") == 0
            if hasProto {
                ch <- url
            }
        }
    }
}

// app main function
func main() {
    foundUrls := make(map[string]bool)
    seedUrls := os.Args[1:]

    // ch = channels
    chUrls := make(chan string)
    chFinished := make(chan bool)

    // crawl process concurrency
    for _, url := range seedUrls {
        go crawl(url, chUrls, chFinished)
    }

    // subscribe to both channels
    for c := 0; c < len(seedUrls); {
        select {
        case url := <-chUrls:
            foundUrls[url] = true
        case <- chFinished:
        c++
        }
    }

    // now we can print the findings
    fmt.Println("\nFound", len(foundUrls), "unique urls:\n")

    // add a dash before each displayed url
    for url, _ := range foundUrls {
        fmt.Println("-" + url)
    }
    close(chUrls)
}
