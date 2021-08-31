package groups

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

//ParseFishki парсим сайт фишек для получения картинок. Принимает номер страницы. Возвращает слайс с урлами картинок
func ParseFishki(number string) (urlList []string, err error) {
	log.Printf("Парсим фишки, номер страницы %v", number)
	if number == "0" {
		number = ""
	}
	res, err := http.Get("https://fishki.net/demotivation/" + number + "/")
	if err != nil {
		return urlList, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		err = fmt.Errorf("status code error: %v Номер страницы %v", res.Status, number)
		return urlList, err
	}
	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return urlList, err
	}
	// Find the review items
	// doc.Find(".drag_list .drag_element .slider--wide slider__cut-hover clearfix .slide__item .post-img picture .picture-holder").Each(func(i int, s *goquery.Selection) {
	doc.Find(".drag_element .picture-holder").Each(func(i int, s *goquery.Selection) {
		title := s.Find("link")
		// dsd := title.Text()
		link, _ := title.Attr("href")
		// fmt.Printf("# %v ## %v ### %v\n", i, dsd, link)
		urlList = append(urlList, link)
	})
	return urlList, err
}
