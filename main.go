// Command manytabs is a chromedp example showing the use of multiple tabs
package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

func main() {
	dir, err := ioutil.TempDir("", "chromedp-manytabs")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Using Temp: %s for chrome's user-data-dir", dir)
	defer os.RemoveAll(dir)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		// chromedp.Flag("disable-gpu", false),
		// chromedp.DisableGPU,
		chromedp.UserDataDir(dir),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer allocCancel()

	// create context for main (first) window/tab
	mainWindowCtx, mainCancel := chromedp.NewContext(allocCtx)
	defer mainCancel() // this will also close all worker tabs

	// ensure the first tab is created
	if err := chromedp.Run(mainWindowCtx); err != nil {
		log.Fatal(err)
	}

	lister(mainWindowCtx)

	var wg sync.WaitGroup

	numWorkers := 3
	for w := 0; w < numWorkers; w++ {
		tabCtx, _ := chromedp.NewContext(mainWindowCtx)
		// ensure the worker tab is created
		if err := chromedp.Run(tabCtx); err != nil {
			log.Fatal(err)
		}
		wg.Add(1)
		go worker(tabCtx, &wg, w, 10, 1000)
	}
	wg.Wait()
	log.Printf("Done waiting for %d workers", numWorkers)
}

func lister(ctx context.Context) {
	url := `https://via.placeholder.com/320x200/f00/fff?text=Main+Window`
	// Navigate to an image page
	if err := chromedp.Run(ctx, showPage(url, `img`)); err != nil {
		log.Fatal(err)
	}
}

func worker(ctx context.Context, wg *sync.WaitGroup, w, n, maxDelay int) {
	// Navigate to an image page
	for i := 0; i < n; i++ {
		if err := chromedp.Run(ctx, showPage(workerImageURL(w, i), `img`)); err != nil {
			log.Fatal(err)
		}
		delay := rand.Intn(maxDelay)
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
	wg.Done()
}

func workerImageURL(w, i int) string {
	// https://via.placeholder.com/320x200/00f/fff?text=Worker:1+Image:42
	return fmt.Sprintf("https://via.placeholder.com/320x200/00f/fff?text=Worker:%d+Image:%d", w, i)
}

// showPage takes a screenshot of a specific element.
func showPage(urlstr, sel string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.WaitVisible(sel, chromedp.ByQuery),
	}
}
