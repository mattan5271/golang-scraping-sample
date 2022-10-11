package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sclevine/agouti"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func main() {
	// Web ドライバーを立ち上げる
	driver := agouti.ChromeDriver()
	err := driver.Start()
	if err != nil {
		fmt.Println(err)
	}
	defer driver.Stop()

	// Chrome を立ち上げる
	page, err := driver.NewPage(agouti.Browser("chrome"))
	if err != nil {
		fmt.Println(err)
	}

	// 価格.com へアクセスする
	err = page.Navigate("https://kakaku.com/")
	if err != nil {
		fmt.Println(err)
	}

	// 検索フォームを取得してキーワードを入力する
	form := page.FindByID("query")
	form.Fill("乾燥機付き洗濯機 ドラム")
	time.Sleep(3 * time.Second)

	// 検索ボタンを取得してクリック
	err = page.FindByID("main_search_button").Click()
	if err != nil {
		fmt.Println(err)
	}
	url, err := page.URL()
	if err != nil {
		fmt.Println(err)
	}
	time.Sleep(3 * time.Second)

	// 検索結果画面の DOM を取得
	res, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	if res.StatusCode != 200 {
		fmt.Println(err)
	}
	defer res.Body.Close()

	// Body を読み込む
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println(err)
	}

	// 検索結果画面から商品情報を取得して出力する
	doc.Find(".p-result_item_row").Each(func(i int, s *goquery.Selection) {
		url, _ := s.Find("div.p-result_item_cell-1 > div.p-item > p > a").Attr("href")
		maker := s.Find("div.p-result_item_cell-1 > div.p-item > div.p-item_detail > p.p-item_maker").Text()
		name := s.Find("div.p-result_item_cell-1 > div.p-item > div.p-item_detail > p.p-item_name").Text()
		price := s.Find("div.p-result_item_cell-2 > div > p.p-item_price > span").Text()

		itemMaker, itemName, itemPrice := encodingScrapeDetails(maker, name, price)
		fmt.Printf("%d メーカー: %s / 商品名: %s / 金額: %s / URL: %s\n", i, itemMaker, itemName, itemPrice, url)
	})
}

// 文字化け対策
func encodingScrapeDetails(maker string, name string, price string) (string, string, string) {
	m := transform.NewReader(strings.NewReader(maker), japanese.ShiftJIS.NewDecoder())
	n := transform.NewReader(strings.NewReader(name), japanese.ShiftJIS.NewDecoder())
	p := transform.NewReader(strings.NewReader(price), japanese.ShiftJIS.NewDecoder())

	itemMaker, _ := io.ReadAll(m)
	itemName, _ := io.ReadAll(n)
	itemPrice, _ := io.ReadAll(p)

	return string(itemMaker), string(itemName), string(itemPrice)
}
