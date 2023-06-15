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

// 汎用ヘルパ (v, err) を受け取ってerrがあれば返す。呼び出し側で二値戻りを使いやすくする。
func ok[T any](v T, err error) (T, error) {
    if err != nil {
        var zero T
        return zero, err
    }
    return v, nil
}

func main() {
    username := "H-Hp" // 自分のGitHubユーザー名

    followers, err := collectAllPages(username, "followers")
    if err != nil {
        log.Fatalf("collect followers: %v", err)
    }
    following, err := collectAllPages(username, "following")
    if err != nil {
        log.Fatalf("collect following: %v", err)
    }

    // セットにして差分を取る (following - followers)
    followersSet := sliceToSet(followers)
    unreciprocated := []string{}
    for _, u := range following {
        if _, ok := followersSet[u]; !ok {
            unreciprocated = append(unreciprocated, u)
        }
    }

    // 結果表示
    fmt.Printf("followers: %d, following: %d, unreciprocated: %d\n", len(followers), len(following), len(unreciprocated))
    for _, u := range unreciprocated {
        fmt.Println(u)
    }
}

// collectAllPagesはpage=1から順にfetchして、ページ内ユーザーを列挙して返す
// tabは"followers"か"following"
func collectAllPages(username, tab string) ([]string, error) {
    var result []string
    page := 1
    for {
        // 1ページを取得してユーザ名リストとhasNextを受け取る
        users, hasNext, err := fetchOnePage(username, tab, page)
        if err != nil {
            return nil, err
        }
        result = append(result, users...)

        //GitHubは50人ずつ表示（PCレイアウトでは）。50人なら次ページがある可能性が高い。
        if !hasNext {
            break
        }
        page++
        // 負荷低減のため短いスリープ
        time.Sleep(500 * time.Millisecond)
    }
    // 重複を排除して返す（念のため）
    return uniqueStrings(result), nil
}

// fetchOnePageは指定ページのHTMLをパースしてユーザー名のスライスと、次ページの存在（hasNext）を返す
/*func fetchOnePage(username, tab string, page int) ([]string, bool, error) {
    u := fmt.Sprintf("https://github.com/%s?page=%d&tab=%s", url.PathEscape(username), page, url.PathEscape(tab))
    doc, err := fetchHTML(u)
    if err != nil {
        return nil, false, err
    }

    // 親ノードを取得: //*[@id='user-profile-frame']/div
    parentNodes := htmlquery.Find(doc, "//*[@id='user-profile-frame']/div")
    var nodesToIterate []*html.Node

    if len(parentNodes) == 0 {
        // フォールバック: 以前のパス（div/div）
        nodesToIterate = htmlquery.Find(doc, "//*[@id='user-profile-frame']/div/div")
    } else {
        // 親ノード内にある各child divを集める。期待する子はparent/div (つまりparentの子div)
        // parentNodesはparent要素群（ページレイアウトによって複数になる可能性あり）
        // ここでは親ノード群の中の直接の子divを集める
        for _, p := range parentNodes {
            // 親の直下の子div要素群を取って追加
            children := htmlquery.Find(p, "./div")
            if len(children) > 0 {
                nodesToIterate = append(nodesToIterate, children...)
            } else {
                // 親ノード自体がユーザー要素の場合
                nodesToIterate = append(nodesToIterate, p)
            }
        }
    }

    usernames := []string{}

    // 親ノードごとに相対XPathで子要素を取得
    for _, n := range nodesToIterate {
        nameNode := htmlquery.FindOne(n, ".//div[2]/a/span[1]")
        if nameNode == nil {
            nameNode = htmlquery.FindOne(n, ".//a[contains(@href,'/')]/span[1]")
        }
        if nameNode == nil {
            // 最後の手段リンクテキスト全体を拾う
            nameNode = htmlquery.FindOne(n, ".//a[contains(@href,'/')]")
        }
        if nameNode == nil {
            continue
        }
        name := strings.TrimSpace(htmlquery.InnerText(nameNode))
        if name == "" {
            continue
        }
        usernames = append(usernames, name)
    }

    // もし上で何も取れなかった場合、絶対XPathでインデックス指定して再試行（デバッグ用）
    if len(usernames) == 0 && len(nodesToIterate) > 0 {
        // nodesToIterateの長さを使って絶対パスを組み立てる
        for i := 0; i < len(nodesToIterate); i++ {
            // 絶対XPath例: //*[@id='user-profile-frame']/div/div[<i+1>]/div[2]/a/span[1]
            abs := "//*[@id='user-profile-frame']/div/div[" + strconv.Itoa(i+1) + "]/div[2]/a/span[1]"
            nameNode := htmlquery.FindOne(doc, abs)
            if nameNode == nil {
                // 別の絶対パターン（親直下のdivの場合）
                abs2 := "//*[@id='user-profile-frame']/div[" + strconv.Itoa(i+1) + "]/div[2]/a/span[1]"
                nameNode = htmlquery.FindOne(doc, abs2)
            }
            if nameNode == nil {
                continue
            }
            name := strings.TrimSpace(htmlquery.InnerText(nameNode))
            if name == "" {
                continue
            }
            usernames = append(usernames, name)
        }
    }

    // ページ内ユーザ数で次ページの存在を推定。通常50人表示。
    hasNext := len(usernames) >= 50
    return usernames, hasNext, nil
}*/
// fetchOnePageは指定ページのHTMLをパースしてユーザー名のスライスと、次ページの存在（hasNext）を返す
func fetchOnePage(username, tab string, page int) ([]string, bool, error) {
    u := fmt.Sprintf("https://github.com/%s?page=%d&tab=%s", url.PathEscape(username), page, url.PathEscape(tab))
    doc, err := fetchHTML(u)
    if err != nil {
        return nil, false, err
    }

    // まずはユーザー要素を直接狙う（複数パターン）
    nodes := htmlquery.Find(doc, "//*[@id='user-profile-frame']//div[contains(@class,'d-table') or contains(@class,'user-list-item') or contains(@class,'follow-list-item') or contains(@class,'d-flex')]")

    // フォールバック　　親ノードから子 div を集める
    if len(nodes) == 0 {
        parents := htmlquery.Find(doc, "//*[@id='user-profile-frame']/div")
        for _, p := range parents {
            if p == nil {
                continue
            }
            children := htmlquery.Find(p, "./div")
            if len(children) > 0 {
                for _, c := range children {
                    if c != nil {
                        nodes = append(nodes, c)
                    }
                }
            } else {
                nodes = append(nodes, p)
            }
        }
    }

    usernames := []string{}
    for _, n := range nodes {
        if n == nil {
            continue
        }
        // 安全な選択・関数呼び出しを含まないXPathで<a>要素を取得し、Go側でTrimする
        aNode := htmlquery.FindOne(n, ".//div[2]/a")
        if aNode == nil {
            aNode = htmlquery.FindOne(n, ".//a[contains(@href,'/')]")
        }
        if aNode == nil {
            // 最低限リンクを探すパターン
            aNode = htmlquery.FindOne(n, ".//a")
        }
        if aNode == nil {
            continue
        }
        // テキストはGo側で正規化する（normalize-space をXPathで呼ばない）
        txt := strings.TrimSpace(htmlquery.InnerText(aNode))
        if txt == "" {
            continue
        }
        usernames = append(usernames, txt)
    }

    // 最終手段　　まだ空なら絶対パスでインデックス指定（デバッグ向け）
    if len(usernames) == 0 {
        limit := 50
        for i := 0; i < limit; i++ {
            abs := "//*[@id='user-profile-frame']/div/div[" + strconv.Itoa(i+1) + "]/div[2]/a"
            nn := htmlquery.FindOne(doc, abs)
            if nn == nil {
                abs2 := "//*[@id='user-profile-frame']/div[" + strconv.Itoa(i+1) + "]/div[2]/a"
                nn = htmlquery.FindOne(doc, abs2)
            }
            if nn == nil {
                continue
            }
            txt := strings.TrimSpace(htmlquery.InnerText(nn))
            if txt != "" {
                usernames = append(usernames, txt)
            }
        }
    }

    hasNext := len(usernames) >= 50
    return usernames, hasNext, nil
}


//fetchHTMLはURLを取得してHTMLノードを返す
func fetchHTML(urlStr string) (*html.Node, error) {
    client := &http.Client{Timeout: 15 * time.Second}
    req, err := http.NewRequest("GET", urlStr, nil)
    if err != nil {
        return nil, err
    }
    //User-Agentを付ける。必要ならCookie/トークンを追加して認証済みリクエストにする。
    req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Go-http-client)")
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // ステータスチェック
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("http status %d for %s", resp.StatusCode, urlStr)
    }

    doc, err := htmlquery.Parse(resp.Body)
    if err != nil {
        return nil, err
    }
    return doc, nil
}

// ユーティリティ・スライスをセットに変換
func sliceToSet(s []string) map[string]struct{} {
    m := make(map[string]struct{}, len(s))
    for _, v := range s {
        m[v] = struct{}{}
    }
    return m
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
