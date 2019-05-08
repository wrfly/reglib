package reglib

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

func mkupRange(s, e int) string {
	return fmt.Sprintf("bytes=%d-%d", s, e)
}

const fixedSize = int(100 * mbSize)

func splitRanges(length int) []string {
	x := []string{}
	if length < fixedSize {
		return []string{mkupRange(0, length)}
	}

	step := length / 10
	if step < fixedSize {
		step = fixedSize
	}
	final := step + length%5
	for i := 0; i < length-final; i += step {
		x = append(x, mkupRange(i, i+step-1))
	}

	return append(x, mkupRange(length-final, length))
}

func (c *Client) parallelDownload(ctx context.Context,
	wg *sync.WaitGroup, path, target string, length int) error {

	wg.Add(1)

	subWG := new(sync.WaitGroup)
	errChan := make(chan error, 5)
	contentRanges := splitRanges(length)

	for part, contentRange := range contentRanges {
		subWG.Add(1)
		go func(part int, contentRange string) {
			defer subWG.Done()

			req, err := http.NewRequest("GET",
				fmt.Sprintf("%s%s", c.baseURL, path), nil)
			if err != nil {
				errChan <- err
				return
			}
			req.Header.Set("Range", contentRange)
			resp, err := c.client.Do(req)
			if err != nil {
				errChan <- err
				return
			}
			defer resp.Body.Close()
			f, err := os.Create(fmt.Sprintf("%s.part%d", target, part))
			if err != nil {
				errChan <- err
				return
			}
			defer f.Close()
			if _, err := io.Copy(f, resp.Body); err != nil {
				errChan <- err
			}
		}(part, contentRange)
	}

	go func() {
		defer wg.Done()

		subWG.Wait()
		for part := len(contentRanges) - 1; part >= 0; part-- {
			if part-1 < 0 {
				break
			}

			curF, err := os.OpenFile(fmt.Sprintf("%s.part%d", target, part),
				os.O_RDONLY, 0644)
			if err != nil {
				errChan <- err
				return
			}

			prvF, err := os.OpenFile(fmt.Sprintf("%s.part%d", target, part-1),
				os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				errChan <- err
				return
			}

			if _, err := io.Copy(prvF, curF); err != nil {
				errChan <- err
			}

			curF.Close()
			prvF.Close()

			os.Remove(fmt.Sprintf("%s.part%d", target, part))
		}
		os.Rename(
			fmt.Sprintf("%s.part0", target),
			target,
		)

		close(errChan)
	}()

	return <-errChan
}
