// stage_read.go
package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
)

func readFiles(ctx context.Context, paths []string) <-chan LogLine {
	out := make(chan LogLine, 100)

	go func() {
		defer close(out)

		for _, path := range paths {
			select {
			case <-ctx.Done():
				return
			default:
			}

			var reader io.ReadCloser
			if path == "-" {
				reader = io.NopCloser(os.Stdin)
			} else {
				f, err := os.Open(path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "파일 열기 실패: %s: %v\n", path, err)
					continue
				}
				reader = f
			}

			scanner := bufio.NewScanner(reader)
			for scanner.Scan() {
				line := LogLine{Source: path, Line: scanner.Text()}
				select {
				case out <- line:
				case <-ctx.Done():
					reader.Close()
					return
				}
			}
			reader.Close()
		}
	}()

	return out
}
