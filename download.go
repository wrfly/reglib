package reglib

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

func makeupRange(s, e int) string {
	return fmt.Sprintf("bytes=%d-%d", s, e)
}

const fixedSize = int(10 * mbSize)

func splitRanges(length int) []string {
	x := []string{}
	if length < fixedSize {
		return []string{makeupRange(0, length)}
	}

	step := length / 5
	if step < fixedSize {
		step = fixedSize
	}
	final := step + length%5
	for i := 0; i < length-final; i += step {
		x = append(x, makeupRange(i, i+step-1))
	}

	return append(x, makeupRange(length-final, length))
}

func (c *Client) parallelDownload(ctx context.Context, path, target string, length int) error {

	wg := new(sync.WaitGroup)
	errChan := make(chan error, 5)
	contentRanges := splitRanges(length)

	downloadStart := time.Now()
	for part, contentRange := range contentRanges {
		wg.Add(1)
		go func(part int, contentRange string) {
			defer wg.Done()

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
	wg.Wait()
	debug("download %s use %s", target, time.Now().Sub(downloadStart))

	start := time.Now()
	for part := len(contentRanges) - 1; part > 0; part-- {
		if part == 0 {
			break
		}

		curF, err := os.OpenFile(fmt.Sprintf("%s.part%d", target, part),
			os.O_RDONLY, 0644)
		if err != nil {
			return err
		}
		prvF, err := os.OpenFile(fmt.Sprintf("%s.part%d", target, part-1),
			os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		if _, err := io.Copy(prvF, curF); err != nil {
			return err
		}
		if err := curF.Close(); err != nil {
			return err
		}
		if err := prvF.Close(); err != nil {
			return err
		}
		if err := os.Remove(fmt.Sprintf("%s.part%d", target, part)); err != nil {
			return err
		}
	}
	debug("merge %s use %s", target, time.Now().Sub(start))

	return os.Rename(fmt.Sprintf("%s.part0", target), target)
}
