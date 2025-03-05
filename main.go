package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
)

var woList = []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")

const (
	maxLen   = 8        // 最大長
	numWorks = 8        // 並列処理のスレッド数
	batch    = 50000000 // 進捗表示の間隔
)

var (
	target   []byte
	found    int32
	result   atomic.Value
	progress atomic.Value
)

func main() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/start", handleStart)
	http.HandleFunc("/progress", handleProgress)
	fmt.Println("サーバーが http://localhost:8080 で起動しました...")
	http.ListenAndServe(":8080", nil)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func handleStart(w http.ResponseWriter, r *http.Request) {
	target = []byte(r.FormValue("target"))
	reset()
	go passCheckParallel()
	fmt.Fprintln(w, "パスワード探索を開始しました。")
}

func passCheckParallel() {
	var wg sync.WaitGroup
	resultChan := make(chan []byte, 1)

	for i := 0; i < numWorks; i++ {
		wg.Add(1)
		go worker(i, numWorks, &wg, resultChan)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for res := range resultChan {
		setResult(res)
	}
}

func worker(startIndex, step int, wg *sync.WaitGroup, resultChan chan []byte) {
	defer wg.Done()

	// 最初の桁を startIndex から開始
	knowPassNum := []int{startIndex}
	passBytes := make([]byte, 1)
	count := 0

	for {
		if atomic.LoadInt32(&found) == 1 {
			return
		}
		count++

		// 現在のパスワードを設定
		for i, num := range knowPassNum {
			passBytes[i] = woList[num]
		}

		// 目標パスワードと比較
		if bytes.Equal(passBytes, target) {
			resultChan <- append([]byte(nil), passBytes...)
			return
		}

		// 進捗表示
		if count%batch == 0 {
			updateProgress(fmt.Sprintf("現在のパスワード: %s", passBytes))
		}

		// 次のパスワードへ（桁の繰り上がり処理）
		carry := true
		for i := len(knowPassNum) - 1; i >= 0 && carry; i-- {
			if i == 0 {
				knowPassNum[i] += step
			} else {
				knowPassNum[i]++
			}

			if knowPassNum[i] >= len(woList) {
				knowPassNum[i] = 0
			} else {
				carry = false
			}
		}

		// すべての桁で繰り上がった場合、新しい桁を追加
		if carry {
			if len(knowPassNum) >= maxLen {
				return
			}
			knowPassNum = append([]int{startIndex}, knowPassNum...)
			passBytes = append([]byte{woList[startIndex]}, passBytes...)
		}
	}
}

func updateProgress(msg string) {
	progress.Store(msg)
}

func setResult(res []byte) {
	atomic.StoreInt32(&found, 1)
	result.Store(string(res))
	progress.Store("パスワードが見つかりました！")
}

func handleProgress(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Progress string `json:"progress"`
		Result   string `json:"result"`
	}{
		Progress: progress.Load().(string),
		Result:   result.Load().(string),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func reset() {
	atomic.StoreInt32(&found, 0)
	result.Store("")
	progress.Store("探索中...")
}
