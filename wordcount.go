package main

import (
    "bufio"
    "flag"
    "fmt"
    "io"
    "os"
    "sort"
    "strings"
    "unicode"
)

type wordFreq struct {
    Word  string
    Count int
}

// 단어 정규화: 소문자화 + 좌우 구두점 제거
func normalize(word string) string {
    w := strings.ToLower(word)
    return strings.TrimFunc(w, func(r rune) bool {
        return !unicode.IsLetter(r) && !unicode.IsDigit(r)
    })
}

func countWords(r io.Reader, minLen int) map[string]int {
    counts := make(map[string]int)
    scanner := bufio.NewScanner(r)
    scanner.Split(bufio.ScanWords)
    for scanner.Scan() {
        w := normalize(scanner.Text())
        if len([]rune(w)) < minLen || w == "" {
            continue
        }
        counts[w]++
    }
    return counts
}

func topN(counts map[string]int, n int) []wordFreq {
    list := make([]wordFreq, 0, len(counts))
    for w, c := range counts {
        list = append(list, wordFreq{w, c})
    }
    sort.Slice(list, func(i, j int) bool {
        if list[i].Count != list[j].Count {
            return list[i].Count > list[j].Count
        }
        return list[i].Word < list[j].Word // 동률은 사전순
    })
    if len(list) > n {
        list = list[:n]
    }
    return list
}

func openInput(args []string) (io.ReadCloser, error) {
    if len(args) > 0 {
        return os.Open(args[0])
    }
    return io.NopCloser(os.Stdin), nil
}

func main() {
    var (
        topCount = flag.Int("n", 10, "출력할 상위 단어 개수")
        minLen   = flag.Int("min", 1, "단어의 최소 길이")
    )
    flag.Usage = func() {
        fmt.Fprintf(flag.CommandLine.Output(),
            "사용법: %s [옵션] [파일경로]\n옵션:\n", os.Args[0])
        flag.PrintDefaults()
    }
    flag.Parse()

    in, err := openInput(flag.Args())
    if err != nil {
        fmt.Fprintln(os.Stderr, "입력 열기 실패:", err)
        os.Exit(1)
    }
    defer in.Close()

    counts := countWords(in, *minLen)
    top := topN(counts, *topCount)

    fmt.Printf("=== Top %d (min=%d) ===\n", *topCount, *minLen)
    for i, wf := range top {
        fmt.Printf("%2d. %-20s %d\n", i+1, wf.Word, wf.Count)
    }
}