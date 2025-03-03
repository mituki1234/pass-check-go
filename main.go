package main

import (
	"fmt"
	"strings"
)

func main() {
    var target string = "aaaaaa"
    fmt.Scan(&target)
    var answer = passCheck(target)
    fmt.Println("パスワード:", answer)
}

// 先頭に要素を追加する関数（スライスを拡張）
func prependInt(slice []int, value int) []int {
    return append([]int{value}, slice...)
}

func prependStr(slice []string, value string) []string {
    return append([]string{value}, slice...)
}

// 文字列を 62 進数のようにカウントアップする関数
func passCheck(target string) string {
    // 数字 '0' ～ '9' + 大文字 'A' ～ 'Z' + 小文字 'a' ～ 'z' のリスト（62進数）
    woList := []string{}
    for ch := '0'; ch <= '9'; ch++ {
        woList = append(woList, string(ch))
    }
    for ch := 'A'; ch <= 'Z'; ch++ {
        woList = append(woList, string(ch))
    }
    for ch := 'a'; ch <= 'z'; ch++ {
        woList = append(woList, string(ch))
    }

    var knowPassNum = []int{0} // 数字のように増加させる
    var result string
    var count int

    for {
        count++;
        // 文字列を組み立てる
        knowPass := make([]string, len(knowPassNum))
        for i := range knowPassNum {
            knowPass[i] = woList[knowPassNum[i]]
        }

        // 文字列を結合してパスワードにする
        result = strings.Join(knowPass, "")

        // 目標のパスワードと一致した場合
        if result == target {
            return result
        }

        // 62 進数のように繰り上がりを考慮しながら増加
        knowPassNum[len(knowPassNum)-1]++

        for i := len(knowPassNum) - 1; i >= 0; i-- {
            if knowPassNum[i] == len(woList) {
                knowPassNum[i] = 0
                if i == 0 {
                    // 一番左の桁が繰り上がった場合、新しい桁を追加
                    knowPassNum = prependInt(knowPassNum, 0)
                } else {
                    knowPassNum[i-1]++
                }
            }
        }

        if count % 5000000 == 0 {
            count = 0
            fmt.Println("計算中",result) // 試行中のパスワードを表示
        }

        // 最大長 8 を超えたら強制終了
        if len(knowPassNum) > 8 {
            break
        }
    }

    return ""
}
