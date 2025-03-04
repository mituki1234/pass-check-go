package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// 62進数の文字リスト
var woList = []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")

var (
	maxLen   = 8        // 最大長
	numWorks = 8        // 並列処理のスレッド数
	batch    = 50000000 // 進捗表示の間隔
)

// グローバルな変数
var (
	target   string
	found    bool
	result   string
	progress string
	mu       sync.Mutex // 進捗の更新を保護するためのロック
)

func main() {
	http.HandleFunc("/start", handleStart)
	http.HandleFunc("/progress", handleProgress)
	http.HandleFunc("/", handleOptions) // CORS対応

	fmt.Println("HTTPSサーバーが https://localhost:8080 で起動しました...")
	err := http.ListenAndServeTLS(":8080", "/etc/ssl/certs/origin.pem", "/etc/ssl/private/origin.key", nil)
	if err != nil {
		fmt.Println("サーバーエラー:", err)
	}
}

// CORS ヘッダーを設定（すべてのオリジンからのリクエストを許可）
func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*") // すべてのオリジンを許可
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS") // 許可するメソッド
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type") // 許可するヘッダー
	w.Header().Set("Access-Control-Allow-Credentials", "true") // 認証情報の送信を許可
}

// OPTIONSリクエストを処理
func handleOptions(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
}

// パスワード探索を開始
func handleStart(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	target = r.FormValue("target")
	reset()
	go passCheckParallel(target)

	fmt.Fprintln(w, "パスワード探索を開始しました。")
}

// 進捗を取得
func handleProgress(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)

	mu.Lock()
	data := struct {
		Progress string `json:"progress"`
		Result   string `json:"result"`
	}{
		Progress: progress,
		Result:   result,
	}
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// 並列処理で総当たり探索
func passCheckParallel(target string) {
	var wg sync.WaitGroup
	resultChan := make(chan string, 1)

	for i := 0; i < numWorks; i++ {
		wg.Add(1)
		go worker(target, i, numWorks, &wg, resultChan)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for res := range resultChan {
		setResult(res)
	}
}

// 各スレッドが担当する範囲を決めて探索
func worker(target string, startIndex, step int, wg *sync.WaitGroup, resultChan chan string) {
	defer wg.Done()

	knowPassNum := []int{startIndex}
	passBytes := make([]byte, 1)
	count := 0

	for {
		if found {
			return
		}
		count++

		for i, num := range knowPassNum {
			passBytes[i] = woList[num]
		}

		if string(passBytes) == target {
			resultChan <- string(passBytes)
			return
		}

		if count%batch == 0 {
			updateProgress(fmt.Sprintf("現在のパスワード: %s", string(passBytes)))
		}

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

		if carry {
			if len(knowPassNum) >= maxLen {
				return
			}
			knowPassNum = append([]int{startIndex}, knowPassNum...)
			passBytes = append([]byte{woList[startIndex]}, passBytes...)
		}
	}
}

// 進捗を更新
func updateProgress(msg string) {
	mu.Lock()
	progress = msg
	mu.Unlock()
}

// 結果をセット
func setResult(res string) {
	mu.Lock()
	result = res
	found = true
	progress = "パスワードが見つかりました！"
	mu.Unlock()
}

// 結果をリセット
func reset() {
	mu.Lock()
	found = false
	result = ""
	progress = "探索中..."
	mu.Unlock()
}
