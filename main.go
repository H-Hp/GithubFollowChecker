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
		
		// テスト・1ページ目のフォロワーを取得
		users, hasNext, err := fetchOnePage(username, "followers", 1)
		if err != nil {
				log.Fatal(err)
		}
		fmt.Printf("取得: %d人, 次ページ: %v\n", len(users), hasNext)
		for _, u := range users {
				fmt.Println(u)
		}
}
// fetchOnePageは指定ページのHTMLをパースしてユーザー名のスライスと、次ページの存在を返す
func fetchOnePage(username, tab string, page int) ([]string, bool, error) {
		u := fmt.Sprintf("https://github.com/%s?page=%d&tab=%s", url.PathEscape(username), page, url.PathEscape(tab))
		doc, err := fetchHTML(u)
		if err != nil {
				return nil, false, err
		}
		// ユーザー要素を取得
		nodes := htmlquery.Find(doc, "//*[@id='user-profile-frame']//div[contains(@class,'d-table') or contains(@class,'user-list-item') or contains(@class,'follow-list-item') or contains(@class,'d-flex')]")
		// フォールバック
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
				aNode := htmlquery.FindOne(n, ".//div[2]/a")
				if aNode == nil {
						aNode = htmlquery.FindOne(n, ".//a[contains(@href,'/')]")
				}
				if aNode == nil {
						aNode = htmlquery.FindOne(n, ".//a")
				}
				if aNode == nil {
						continue
				}
				txt := strings.TrimSpace(htmlquery.InnerText(aNode))
				if txt == "" {
						continue
				}
				usernames = append(usernames, txt)
		}
		// 最終手段
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
// fetchHTML は前のコミットから継続
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