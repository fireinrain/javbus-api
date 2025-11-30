package scraper

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/dustin/go-humanize"
	"github.com/fireinrain/javbus-api/model"
)

func TestJavbusScraper_GetMovieMagnets(t *testing.T) {
	file, err2 := os.ReadFile("../magnets.html")
	if err2 != nil {
		t.Error(err2)
	}
	str := string(file)
	sprintf := fmt.Sprintf("<table>%s</table>", str)

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(sprintf)))
	if err != nil {
		t.Error(err)
	}

	var magnets []model.Magnet

	// 3. 解析 DOM
	doc.Find("tr").Each(func(i int, tr *goquery.Selection) {
		firstAnchor := tr.Find("td a").First()
		link := firstAnchor.AttrOr("href", "")

		// 提取 Hash ID
		idMatch := magnetIDRegex.FindStringSubmatch(link)
		id := ""
		if len(idMatch) > 1 {
			id = idMatch[1]
		}

		if id == "" {
			return
		}

		// 处理 Tags (高清/字幕) 并清洗标题
		isHD := false
		hasSub := false

		// 获取原始完整标题
		fullTitle := firstAnchor.Text()

		tr.Find("td a a").Each(func(_ int, tag *goquery.Selection) {
			t := tag.Text()
			if strings.Contains(t, "高清") {
				isHD = true
			}
			if strings.Contains(t, "字幕") {
				hasSub = true
			}
			// 技巧：把 tag 的文本从完整标题中移除，得到纯净标题
			fullTitle = strings.ReplaceAll(fullTitle, t, "")
		})

		cleanTitle := strings.TrimSpace(fullTitle)

		// 提取其他字段
		sizeStr := strings.TrimSpace(tr.Find("td").Eq(1).Find("a").Text())
		dateStr := strings.TrimSpace(tr.Find("td").Eq(2).Find("a").Text())

		// 解析文件大小
		// humanize.ParseBytes 返回 uint64，需要转 int64
		numSize, _ := humanize.ParseBytes(sizeStr)
		nsInt := int64(numSize)

		magnets = append(magnets, model.Magnet{
			ID:          id,
			Link:        link,
			IsHD:        isHD,
			HasSubtitle: hasSub,
			Title:       cleanTitle,
			Size:        sizeStr, // 假设 Model 是 *string
			NumberSize:  nsInt,   // 假设 Model 是 *int64
			ShareDate:   dateStr, // 假设 Model 是 *string
		})
	})
	fmt.Println(magnets)
}
