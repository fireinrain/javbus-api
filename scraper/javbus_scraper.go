package scraper

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"regexp"
	"sort"
	"strconv"
	"strings"
	//_ "golang.org/x/image/webp"

	"github.com/PuerkitoBio/goquery"
	"github.com/dustin/go-humanize"
	"github.com/fireinrain/javbus-api/config"
	"github.com/fireinrain/javbus-api/consts"
	"github.com/fireinrain/javbus-api/model"
	"github.com/fireinrain/javbus-api/utils"
	"github.com/go-resty/resty/v2"
)

// -------------------------------------------------------------
// 正则表达式预编译 (提升性能)
// -------------------------------------------------------------
var (
	// gid 和 uc 的提取正则: var gid = 12345;
	gidRegex = regexp.MustCompile(`var gid = (\d+);`)
	ucRegex  = regexp.MustCompile(`var uc = (\d+);`)
	// 磁力链接 ID 提取: magnet:?xt=urn:btih:xxxx
	magnetIDRegex = regexp.MustCompile(`magnet:\?xt=urn:btih:(\w+)`)
	// 样品图文件名提取
	sampleIDRegex = regexp.MustCompile(`(\S+)\.(jpe?g|png|webp|gif)$`)
	// 标题中的页码提取
	titlePageRegex = regexp.MustCompile(`^(?:第\d+?頁 - )?(.+?) - `)
)

// StarInfoMap 映射表
var starInfoMap = map[string]string{
	"Birthday":   "生日: ",
	"Age":        "年齡: ",
	"Height":     "身高: ",
	"Bust":       "胸圍: ",
	"Waistline":  "腰圍: ",
	"Hipline":    "臀圍: ",
	"Birthplace": "出生地: ",
	"Hobby":      "愛好: ",
}

// 对应 Node.js 的 /^(?:第\d+?頁 - )?(.+?) - /
var titleRegex = regexp.MustCompile(`^(?:第\d+?頁 - )?(.+?) - `)

type JavbusScraper struct {
	SiteUrl string
	Client  *resty.Client
}

func NewJavbusScraper(cfg *config.Config) *JavbusScraper {
	client := NewRestyClient(cfg)
	return &JavbusScraper{
		SiteUrl: consts.JavBusURL,
		Client:  client,
	}
}

// -------------------------------------------------------------
// 解析器核心逻辑
// -------------------------------------------------------------

// requestText 辅助方法：发送请求并返回 GoQuery Document
func (s *JavbusScraper) requestDocument(url string, headers map[string]string) (*goquery.Document, error) {
	// 1. 使用 Resty 链式调用
	// .R() 创建请求对象
	// .SetHeaders() 直接支持 map[string]string，无需循环遍历
	// .Get() 发起 GET 请求
	resp, err := s.Client.R().
		SetHeaders(headers).
		Get(url)

	if err != nil {
		return nil, err
	}

	// 2. 检查状态码
	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("request failed with status code: %d", resp.StatusCode())
	}
	// 3. 转换 Resty Body 为 goquery Document
	// Resty 的 resp.Body() 返回 []byte，goquery 需要 io.Reader
	return goquery.NewDocumentFromReader(bytes.NewReader(resp.Body()))
}

func parseFilterInfo(doc *goquery.Document, filterType, filterValue string) *model.FilterInfo {
	title := doc.Find("title").Text()

	// 提取名称 (例如演员名、类别名)
	matches := titleRegex.FindStringSubmatch(title)
	name := ""
	if len(matches) > 1 {
		name = matches[1]
	}

	return &model.FilterInfo{
		Name:  name,
		Type:  filterType,
		Value: filterValue,
	}
}

// 对应 getMoviesByPage
// GetMoviesByPage 获取电影列表 (分页)
func (s *JavbusScraper) GetMoviesByPage(q *model.GetMoviesQuery) (*model.MoviesPage, error) {
	// 1. 处理页码 (int -> string)
	page := "1"
	pageInt, _ := strconv.Atoi(q.Page)
	if pageInt > 1 {
		page = strconv.Itoa(pageInt)
	}

	// 2. 构造 URL
	// 基础前缀
	prefix := consts.JavBusURL
	if q.Type != "" && q.Type != model.MovieTypeNormal {
		prefix = fmt.Sprintf("%s/%s", consts.JavBusURL, q.Type)
	}

	// 叠加 FilterType
	if q.FilterType != "" {
		prefix = fmt.Sprintf("%s/%s", prefix, q.FilterType)
	}

	// 拼接完整 URL
	var url string
	if page == "1" {
		if q.FilterType != "" {
			// /genre/id
			url = fmt.Sprintf("%s/%s", prefix, q.FilterValue)
		} else {
			// /
			url = prefix
		}
	} else {
		if q.FilterType != "" {
			// /genre/id/page
			url = fmt.Sprintf("%s/%s/%s", prefix, q.FilterValue, page)
		} else {
			// /page/2
			url = fmt.Sprintf("%s/page/%s", prefix, page)
		}
	}

	// 3. 构造 Headers (Cookie)
	headers := map[string]string{}
	// q.Magnet 来自 model 定义的枚举
	if q.Magnet == model.MagnetTypeExist {
		headers["Cookie"] = "existmag=mag"
	} else {
		headers["Cookie"] = "existmag=all"
	}

	// 4. 请求文档
	doc, err := s.requestDocument(url, headers)
	if err != nil {
		return nil, err
	}

	// 5. 解析列表
	moviesPage := parseMoviesPage(doc)

	// 6. 解析 Filter 信息 (如果存在筛选)
	// Node.js逻辑: filterType && filterValue ? parseFilterInfo(...) : undefined
	if q.FilterType != "" && q.FilterValue != "" {
		moviesPage.Filter = parseFilterInfo(doc, string(q.FilterType), q.FilterValue)
	}

	return moviesPage, nil
}

// 对应 getMoviesByKeywordAndPage
// GetMoviesByKeywordAndPage
func (s *JavbusScraper) GetMoviesByKeywordAndPage(keyword string, q *model.GetMoviesQuery) (*model.SearchMoviesPage, error) {
	// 1. 构造 URL
	prefix := fmt.Sprintf("%s/search", consts.JavBusURL)
	if q.Type != "" && q.Type != model.MovieTypeNormal {
		prefix = fmt.Sprintf("%s/%s/search", consts.JavBusURL, q.Type)
	}

	page := q.Page
	if page == "" {
		page = "1"
	}

	// Go 的 url.PathEscape 对应 encodeURIComponent，但保留了 '/'，通常够用，如果需要严格一致可以使用 query escape
	url := fmt.Sprintf("%s/%s/%s&type=1", prefix, strings.TrimSpace(keyword), page)

	// 2. Headers
	headers := map[string]string{}
	if q.Magnet == model.MagnetTypeExist {
		headers["Cookie"] = "existmag=mag"
	} else {
		headers["Cookie"] = "existmag=all"
	}

	doc, err := s.requestDocument(url, headers)
	if err != nil {
		// 搜索结果为空时 JavBus 可能会返回 404，这里需要在上层处理
		return nil, err
	}

	moviesPage := parseMoviesPage(doc)

	return &model.SearchMoviesPage{
		MoviesPage: *moviesPage,
		Keyword:    keyword,
	}, nil
}

// parseMoviesPage 通用页面解析逻辑
func parseMoviesPage(doc *goquery.Document) *model.MoviesPage {
	var movies []model.Movie

	doc.Find("#waterfall #waterfall .item").Each(func(i int, s *goquery.Selection) {
		imgTag := s.Find(".photo-frame img")
		rawImg := imgTag.AttrOr("src", "")
		title := imgTag.AttrOr("title", "")

		infoDate := s.Find(".photo-info date")
		id := infoDate.Eq(0).Text()
		date := infoDate.Eq(1).Text()

		var tags []string
		s.Find(".item-tag button").Each(func(i int, btn *goquery.Selection) {
			tags = append(tags, btn.Text())
		})

		// 格式化图片 URL (可能需要加上 host)
		img := utils.FormatImageURL(rawImg)

		if id != "" {
			movies = append(movies, model.Movie{
				ID:    id,
				Title: title,
				Img:   img,
				Date:  date,
				Tags:  tags,
			})
		}
	})

	// 分页信息解析
	activePageStr := doc.Find(".pagination .active a").Text()
	if activePageStr == "" {
		activePageStr = "1"
	}
	currentPage, _ := strconv.Atoi(activePageStr)

	var pages []int
	doc.Find(".pagination li a").Each(func(i int, s *goquery.Selection) {
		if p, err := strconv.Atoi(s.Text()); err == nil {
			pages = append(pages, p)
		}
	})

	hasNextPage := doc.Find(".pagination li #next").Length() > 0
	var nextPage int
	if hasNextPage {
		np := currentPage + 1
		nextPage = np
	}

	return &model.MoviesPage{
		Movies: movies,
		Pagination: model.Pagination{
			CurrentPage: currentPage,
			HasNextPage: hasNextPage,
			NextPage:    nextPage,
			Pages:       pages,
		},
	}
}

// 对应 getMovieDetail
// GetMovieDetail 获取详情

var ReqHeaders = map[string]string{
	"Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
	// "Accept-Encoding":          "gzip, deflate, br, zstd", // 建议注释掉，让 Go 自动处理 gzip。强行加 br 可能会导致乱码。
	"Cache-Control":               "no-cache",
	"Pragma":                      "no-cache",
	"Priority":                    "u=0, i",
	"Sec-Ch-Ua":                   `"Chromium";v="142", "Google Chrome";v="142", "Not_A Brand";v="99"`,
	"Sec-Ch-Ua-Arch":              `"x86"`,
	"Sec-Ch-Ua-Bitness":           `"64"`,
	"Sec-Ch-Ua-Full-Version":      `"142.0.7444.176"`,
	"Sec-Ch-Ua-Full-Version-List": `"Chromium";v="142.0.7444.176", "Google Chrome";v="142.0.7444.176", "Not_A Brand";v="99.0.0.0"`,
	"Sec-Ch-Ua-Mobile":            "?0",
	"Sec-Ch-Ua-Model":             `""`,
	"Sec-Ch-Ua-Platform":          `"macOS"`,
	"Sec-Ch-Ua-Platform-Version":  `"12.7.0"`,
	"Sec-Fetch-Dest":              "document",
	"Sec-Fetch-Mode":              "navigate",
	"Sec-Fetch-Site":              "none",
	"Sec-Fetch-User":              "?1",
	"Upgrade-Insecure-Requests":   "1",
	"User-Agent":                  "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36",
}

// 浅拷贝示例
func shallowCopyMap(original map[string]string) map[string]string {
	// 创建新map，大小为原map的长度
	newMap := make(map[string]string, len(original))

	// 遍历原map并复制键值对
	for key, value := range original {
		newMap[key] = value
	}
	return newMap
}

func (s *JavbusScraper) GetMovieDetail(id string) (*model.MovieDetail, error) {
	url := fmt.Sprintf("%s/%s", consts.JavBusURL, id)
	var cookieStr = ""
	headerMap := shallowCopyMap(ReqHeaders)
	headerMap["Cookie"] = cookieStr
	// 这里需要原始 HTML 字符串来做正则匹配 (gid/uc)，所以不能只用 goquery
	// 为了复用 requestDocument 的逻辑，我们可以稍作修改，或者这里单独发请求
	// 为了简单，我们先获取 Document，再获取 HTML 字符串
	doc, err := s.requestDocument(url, headerMap)
	if err != nil {
		return nil, err
	}

	html, _ := doc.Html()

	// 1. 标题与图片
	title := doc.Find(".container h3").Text()
	bigImg := doc.Find(".container .movie .bigImage img").AttrOr("src", "")
	imgURL := utils.FormatImageURL(bigImg)

	// 2. 图片尺寸探测 (Probe)
	var imageSize *model.ImageSize
	if imgURL != "" {
		// shharn/imagesize 库使用
		// 也可以自己实现只下载前几 KB
		// 为了不阻塞太久，可以在这里开个协程或者设置短超时
		width, height, _, err := getImageDimensions(s.Client, imgURL, url)
		if err == nil {
			imageSize = &model.ImageSize{
				Width:  width,
				Height: height,
			}
		}
		// 这里简化处理，暂不实现复杂的 Probe 逻辑，因为 Go 的 image 包需要下载 Body
		// 如果必须，可以使用 s.Client 发起 Range 请求
	}

	// 3. 基础信息解析
	infoNodes := doc.Find(".container .movie .info p")

	// Helper Funcs
	textInfo := func(header string, exclude string) string {
		var res string
		infoNodes.Each(func(i int, s *goquery.Selection) {
			if strings.Contains(s.Find(".header").Text(), header) {
				// 获取最后一个文本节点 (lastChild logic in goquery is tricky)
				// 通常取 Text() 后 Replace 掉 Header
				fullText := s.Text()
				val := strings.TrimSpace(strings.Split(fullText, ":")[1]) // 简单粗暴分割
				if exclude != "" {
					val = strings.ReplaceAll(val, exclude, "")
				}
				val = strings.TrimSpace(val)
				res = val
			}
		})
		return res
	}

	linkInfo := func(header string, prefix string) *model.Property {
		var prop *model.Property
		infoNodes.Each(func(i int, s *goquery.Selection) {
			if strings.Contains(s.Find(".header").Text(), header) {
				a := s.Find("a")
				if a.Length() > 0 {
					href := a.AttrOr("href", "")
					name := strings.TrimSpace(a.Text())

					// ID 提取逻辑
					isUncensored := strings.Contains(href, "uncensored")
					computedPrefix := prefix
					if isUncensored {
						computedPrefix = "uncensored/" + prefix
					}

					id := strings.Replace(href, consts.JavBusURL+"/"+computedPrefix+"/", "", 1)
					if id != "" && isUncensored {
						id = "uncensored/" + id
					}

					prop = &model.Property{ID: id, Name: name}
				}
			}
		})
		return prop
	}

	date := textInfo("發行日期", "")
	videoLengthStr := textInfo("長度", "分鐘")
	var videoLength int
	if videoLengthStr != "" {
		if v, err := strconv.Atoi(videoLengthStr); err == nil {
			videoLength = v
		}
	}

	director := linkInfo("導演", "director")
	producer := linkInfo("製作商", "studio")
	publisher := linkInfo("發行商", "label")
	series := linkInfo("系列", "series")

	// 批量提取 Genres 和 Stars
	var genres []model.Property
	var stars []model.Property

	infoNodes.Each(func(i int, s *goquery.Selection) {
		// Genres
		if s.Find(".genre").Length() > 0 && s.Find("span[onmouseover]").Length() == 0 {
			s.Find(".genre").Each(func(j int, g *goquery.Selection) {
				a := g.Find("label a")
				if a != nil && a.Length() > 0 {
					name := a.Text()
					href := a.AttrOr("href", "")
					id := strings.Split(href, "/genre/")[1] // 简化提取
					genres = append(genres, model.Property{ID: id, Name: name})
				}
			})
		}
		// Stars (hover trigger)
		if s.Find(".genre").Length() > 0 && s.Find("span[onmouseover]").Length() > 0 {
			s.Find(".genre").Each(func(j int, g *goquery.Selection) {
				a := g.Find("a")
				if a != nil && a.Length() > 0 {
					name := a.Text()
					href := a.AttrOr("href", "")
					id := strings.Split(href, "/star/")[1] // 简化提取
					stars = append(stars, model.Property{ID: id, Name: name})
				}
			})
		}
	})

	// 4. 正则提取 GID / UC
	var gidStr, ucStr string
	if match := gidRegex.FindStringSubmatch(html); len(match) > 1 {
		gidStr = match[1]
	}
	if match := ucRegex.FindStringSubmatch(html); len(match) > 1 {
		ucStr = match[1]
	}

	// 5. 样品图
	var samples []model.Sample
	doc.Find("#sample-waterfall .sample-box").Each(func(i int, s *goquery.Selection) {
		href := s.AttrOr("href", "")
		img := s.Find(".photo-frame img")
		thumb := img.AttrOr("src", "")
		alt := img.AttrOr("title", "")

		// ID 提取
		// URL: .../sample1.jpg
		parts := strings.Split(thumb, "/")
		filename := parts[len(parts)-1]
		idMatch := sampleIDRegex.FindStringSubmatch(filename)
		id := ""
		if len(idMatch) > 1 {
			id = idMatch[1]
		}

		samples = append(samples, model.Sample{
			Alt:       alt,
			ID:        id,
			Thumbnail: utils.FormatImageURL(thumb),
			Src:       utils.FormatImageURL(href),
		})
	})

	// 6. 相似影片
	var similar []model.SimilarMovie
	doc.Find("#related-waterfall a").Each(func(i int, s *goquery.Selection) {
		href := s.AttrOr("href", "")
		parts := strings.Split(href, "/")
		id := parts[len(parts)-1]
		title := s.AttrOr("title", "")
		img := s.Find("img").AttrOr("src", "")
		fImg := utils.FormatImageURL(img)

		similar = append(similar, model.SimilarMovie{
			ID:    id,
			Title: title,
			Img:   fImg,
		})
	})

	return &model.MovieDetail{
		ID:            id,
		Title:         title,
		Img:           imgURL,
		ImageSize:     imageSize,
		Date:          date,
		VideoLength:   videoLength,
		Director:      director,
		Producer:      producer,
		Publisher:     publisher,
		Series:        series,
		Genres:        genres,
		Stars:         stars,
		Samples:       samples,
		SimilarMovies: similar,
		GID:           gidStr,
		UC:            ucStr,
	}, nil
}

// GetImageDimensions 获取图片尺寸而不下载全图
func getImageDimensions(client *resty.Client, url string, pageUrl string) (int, int, string, error) {
	// 关键点 1: SetDoNotParseResponse(true)
	//TODO need fix
	// 告诉 Resty 不要自动读取和关闭 Body，把 Body 的控制权交给我们
	// 这样我们就可以像操作文件流一样操作网络流
	headers := shallowCopyMap(ReqHeaders)
	headers["Referer"] = pageUrl
	headers["Range"] = "bytes=0-512"
	headers["Cookie"] = ""
	headers["Accept"] = "image/webp,image/apng,image/*,*/*;q=0.8"
	resp, err := client.R().SetHeaders(headers).
		SetDoNotParseResponse(true).
		// 可选优化: 加上 Range 头，只请求前 32KB 数据。
		// 大多数图片头部都在前几 KB，但这取决于服务器是否支持 Range。
		// 如果不加这个头，Resty 会发起全量 Get，但我们靠下方的 Close() 提前掐断连接。
		// SetHeader("Range", "bytes=0-32768").
		Get(url)

	if err != nil {
		return 0, 0, "", err
	}
	defer resp.RawBody().Close()

	// 检查状态码 (206 Partial Content 是成功的标志)
	if resp.StatusCode() != 200 && resp.StatusCode() != 206 {
		return 0, 0, "", fmt.Errorf("http status: %d", resp.StatusCode())
	}
	// 4. 为了保险，先读取到内存 (32KB 很小，不会炸内存)
	// 直接传 resp.RawBody() 给 DecodeConfig 有时会因为网络包导致的 reader 行为差异而出错
	// 读成 []byte 最稳。
	//data, err := io.ReadAll(resp.RawBody())
	//if err != nil {
	//	return 0, 0, "", err
	//}
	//fmt.Println("----- 诊断报告 -----")
	//fmt.Printf("HTTP 状态码: %d\n", resp.StatusCode())
	//fmt.Printf("Content-Type: %s\n", resp.Header().Get("Content-Type"))
	//fmt.Printf("实际读取字节数: %d\n", len(data))
	//fmt.Printf("前 50 字节 (Hex): %x\n", data[:50])
	//fmt.Printf("前 50 字节 (String): %s\n", string(data[:50]))
	//fmt.Println("-------------------")

	// 5. 解析
	config, format, err := image.DecodeConfig(resp.RawBody())
	//fmt.Println("解析结果: ", err)
	//fmt.Printf("宽度: %d\n", config.Width)
	//fmt.Printf("高度: %d\n", config.Height)
	//fmt.Printf("格式: %s\n", format)
	if err != nil {
		// 如果这里还报错，说明不是 Go 支持的格式 (JPG/PNG/GIF/WebP)
		// 或者 32KB 依然不够 (极少见)
		return 0, 0, "", fmt.Errorf("解析错误: %v (请检查是否引入了 x/image/webp)", err)
	}

	return config.Width, config.Height, format, nil
}

// 对应 getStarInfo
// GetStarInfo 获取演员详细信息
// 对应 TS: export async function getStarInfo(starId: string, type?: MovieType)
func (s *JavbusScraper) GetStarInfo(starId string, movieType string) (*model.StarInfo, error) {
	// 1. 构造 URL 前缀
	prefix := consts.JavBusURL
	// 对应 !type || type === 'normal'
	if movieType != "" && movieType != "normal" {
		prefix = fmt.Sprintf("%s/%s", consts.JavBusURL, movieType)
	}
	url := fmt.Sprintf("%s/star/%s", prefix, starId)

	// 2. 发起请求 (Resty)
	resp, err := s.Client.R().Get(url)
	if err != nil {
		return nil, err
	}

	// 3. 加载 HTML
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body()))
	if err != nil {
		return nil, err
	}

	// 4. 解析并返回
	return parseStarInfo(doc, starId), nil
}

// parseStarInfo 解析演员详情 HTML
// 对应 TS: export function parseStarInfo(pageHTML: string, starId: string): StarInfo
func parseStarInfo(doc *goquery.Document, starId string) *model.StarInfo {
	// 1. 定位容器: #waterfall .item .avatar-box
	box := doc.Find("#waterfall .item .avatar-box")

	// 2. 解析头像 (Avatar)
	// 对应 TS: formatImageUrl(doc?.querySelector('.photo-frame img')?.getAttribute('src')) ?? null
	rawAvatar := box.Find(".photo-frame img").AttrOr("src", "")
	formattedAvatar := utils.FormatImageURL(rawAvatar)

	var avatarPtr string
	if formattedAvatar != "" {
		avatarPtr = formattedAvatar
	}

	// 3. 解析姓名 (Name)
	// 对应 TS: doc?.querySelector('.photo-info .pb10')?.textContent ?? ''
	name := strings.TrimSpace(box.Find(".photo-info .pb10").Text())

	// =================================================================
	// 4. 解析其他属性 (Rest)
	// 在 Go 中，与其用 Map 遍历反射，不如定义一个 Helper 函数闭包来提取值
	// =================================================================

	// 获取所有 P 标签，避免每次查找都重新遍历 DOM
	infoNodes := box.Find(".photo-info p")

	// Helper: 根据前缀查找文本 (返回 *string)
	// 对应 TS 的 infos?.find(p => p.textContent.includes(mapValue))...replace...
	getText := func(prefix string) string {
		var result string
		// EachWithBreak 允许我们在找到后立即停止遍历，性能更好
		infoNodes.EachWithBreak(func(i int, s *goquery.Selection) bool {
			text := s.Text()
			if strings.Contains(text, prefix) {
				// 移除前缀 (如 "生日: ") 并去空格
				val := strings.TrimSpace(strings.Replace(text, prefix, "", 1))
				if val != "" {
					result = val
				}
				return false // 返回 false 停止遍历
			}
			return true // 返回 true 继续遍历
		})
		return result
	}

	// Helper: 根据前缀查找并转为数字 (返回 *int)
	getInt := func(prefix string) int {
		strVal := getText(prefix)
		if strVal != "" {
			if intVal, err := strconv.Atoi(strVal); err == nil {
				return intVal
			}
		}
		return 0
	}

	// 5. 组装并返回结构体
	// 这里的硬编码字符串对应原 TS 代码中的 starInfoMap
	return &model.StarInfo{
		ID:         starId,
		Name:       name,
		Avatar:     avatarPtr,
		Birthday:   getText("生日: "),
		Age:        getInt("年齡: "), // model 中 Age 是 int
		Height:     getInt("身高: "), // model 中 Height 是 int
		Bust:       getText("胸圍: "),
		Waistline:  getText("腰圍: "),
		Hipline:    getText("臀圍: "),
		Birthplace: getText("出生地: "),
		Hobby:      getText("愛好: "),
	}
}

// 对应 getMovieMagnets
// GetMovieMagnets 获取磁力链接 (Ajax)
func (s *JavbusScraper) GetMovieMagnets(movieId, gid, uc, sortBy, sortOrder string) ([]model.Magnet, error) {
	// 1. 使用 Resty 发起请求
	// Resty 会自动处理 URL 参数编码，不需要手动 fmt.Sprintf 拼接参数
	resp, err := s.Client.R().
		SetQueryParams(map[string]string{
			"gid":  gid,
			"lang": "zh",
			"uc":   uc,
		}).
		SetHeader("Referer", fmt.Sprintf("%s/%s", consts.JavBusURL, movieId)).
		Get(consts.JavBusURL + "/ajax/uncledatoolsbyajax.php")

	if err != nil {
		return nil, err
	}

	// 2. 加载 HTML
	// Resty 的 Body() 返回 []byte，需要转换为 Reader 给 goquery 使用
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp.Body()))
	if err != nil {
		return nil, err
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

	// 4. 排序逻辑
	// 提取比较函数以保持代码整洁
	sort.Slice(magnets, func(i, j int) bool {
		a, b := magnets[i], magnets[j]

		// 如果指定了排序规则
		if sortBy != "" && sortOrder != "" {
			if sortBy == "size" {
				valA := int64(0)
				valB := int64(0)
				if a.NumberSize >= 0 {
					valA = a.NumberSize
				}
				if b.NumberSize >= 0 {
					valB = b.NumberSize
				}

				if sortOrder == "asc" {
					return valA < valB
				}
				return valA > valB
			}

			if sortBy == "date" {
				// 简单的字符串日期比较
				valA := ""
				valB := ""
				if a.ShareDate != "" {
					valA = a.ShareDate
				}
				if b.ShareDate != "" {
					valB = b.ShareDate
				}

				if sortOrder == "asc" {
					return valA < valB
				}
				return valA > valB
			}
		}

		// 默认排序：按大小降序
		valA := int64(0)
		valB := int64(0)
		if a.NumberSize >= 0 {
			valA = a.NumberSize
		}
		if b.NumberSize >= 0 {
			valB = b.NumberSize
		}
		return valA > valB
	})

	return magnets, nil
}

func (s *JavbusScraper) GetAccessJavbus() (*model.JavbusAccessStatus, error) {
	// 1. 使用 Resty 发起请求
	// Resty 会自动处理 URL 参数编码，不需要手动 fmt.Sprintf 拼接参数
	resp, err := s.Client.R().
		SetHeader("User-Agent", consts.UserAgent).
		Get(consts.JavBusURL)

	if err != nil {
		return &model.JavbusAccessStatus{
			Access:  false,
			Message: fmt.Sprintf("network error: %v", err),
		}, err
	}

	if resp != nil && resp.StatusCode() == 200 {
		return &model.JavbusAccessStatus{
			Access:  true,
			Message: "access javbus.com success!",
		}, nil
	}

	return &model.JavbusAccessStatus{
		Access:  false,
		Message: fmt.Sprintf("you may need use proxy for access javbus.com: %v", err),
	}, nil
}
