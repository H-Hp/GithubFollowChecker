package main

import (
    "fmt"
    "log"
    "net/http"
    "net/url"
    "strconv"
    "strings"
    "time"

    "github.com/antchfx/htmlquery"
    "golang.org/x/net/html"
)


/*
Githubで相互フォロー以外の自分だけが相手をフォローしているアカウントを発見する

	//1ページに50人まで表示される
	//2ページ目以降はURLが変わる： https://github.com/H-Hp?page=2&tab=followers

処理の流れ 
mainメソッドからcollectAllPagesを呼び出し、ループさせ、フォロワーとフォロー中のユーザー名の配列を作成していく
	1ページ目のフォロワーの一覧ページのHTMLを取得しパース
	xPathでユーザー一覧のdivを取得
	ループさせて、ユーザーを配列に入れていく
	配列が50人溜まったら、次のページも存在するかもなのでページNoをインクリメントし処理を繰り返す
フォロワーとフォロー中のユーザー名の配列を作成後に、2つを比較し一致しないユーザーを一覧表示

*/
func main() {
    username := "H-Hp"
    fmt.Printf("GitHub相互フォローチェッカー\nユーザー: %s\n", username)
    
   
}

// fetchHTMLはURLを取得してHTMLノードを返す
func fetchHTML(urlStr string) (*html.Node, error) {
    client := &http.Client{Timeout: 15 * time.Second}
    req, err := http.NewRequest("GET", urlStr, nil)
    if err != nil {
        return nil, err
    }
    req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Go-http-client)")
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("http status %d for %s", resp.StatusCode, urlStr)
    }

    doc, err := htmlquery.Parse(resp.Body)
    if err != nil {
        return nil, err
    }
    return doc, nil
}

// ユーティリティ・重複削除
func uniqueStrings(in []string) []string {
    m := make(map[string]struct{}, len(in))
    out := []string{}
    for _, v := range in {
        if _, ok := m[v]; ok {
            continue
        }
        m[v] = struct{}{}
        out = append(out, v)
    }
    return out
}