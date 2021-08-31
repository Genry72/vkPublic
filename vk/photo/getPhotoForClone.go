package photo

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"bot.my/utils"
)

//GetPhotoList Возвращает колличетво и список урлов фоток на аватарке
func GetPhotoList(versinonAPIVk, token, ownerID, myID string, vkErrUserChan chan error) (count int, urls []string, err error) {
	log.Printf("Проверяем количество аватарок у пользоватлея %v", ownerID)
	err = utils.Antiban(myID, "GetPhotoList", 300)
	if err != nil {
		vkErrUserChan <- err
	}
	url := "https://api.vk.com/method/photos.get?v=" + versinonAPIVk + "&access_token=" + token + "&owner_id=" + ownerID + "&album_id=profile"
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return count, urls, err
	}
	res, err := client.Do(req)
	if err != nil {
		return count, urls, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return count, urls, err
	}
	if res.StatusCode != 200 {
		err = fmt.Errorf("Ошибка при получении списка фоток пользователя " + res.Status + " " + string(body))
		return count, urls, err
	}
	if strings.Contains(string(body), "failed") == true || strings.Contains(string(body), "error") {
		err = fmt.Errorf("Ошибка при получении списка фоток пользователяр: %v", string(body))
		return count, urls, err
	}
	m := getPhotoListStruct{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		err = fmt.Errorf("Ошибка парсинга ответа на запрос GetPhotoList: %v Боди: %v", err, string(body))
		return count, urls, err
	}
	var curl string                          //Переменная для хранения урла максимального изображения
	for _, photo := range m.Response.Items { //Идем по всем фоткам
		maxSizePhoto := 0
		for _, size := range photo.Sizes {
			if size.Height > maxSizePhoto {
				maxSizePhoto = size.Height
				curl = size.URL
			}
		}
		urls = append(urls, curl)

	}
	//Добавляем запись в историчну таблицу о выполнении запроса
	err = utils.InsertToDB(fmt.Sprintf("INSERT INTO %v VALUES('%v', '%v', '%v', %v);", "history", time.Now().UnixNano(), "GetInfo", myID, "NOW()"))
	if err != nil {
		vkErrUserChan <- err
	}
	log.Printf("У пользователя %v %v фоток на аватарке", ownerID, m.Response.Count)
	return m.Response.Count, urls, err
}

//DownloadPhoto загрузка фоток
func DownloadPhoto(filepath string, url string) (err error) {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		err = fmt.Errorf("Ошибка загрузки фото: %v", err)
		return err
	}
	defer resp.Body.Close()
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		err = fmt.Errorf("Ошибка создания файла для загрузки фото: %v", err)
		return err
	}
	defer out.Close()
	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	// log.Printf("Загружена фотка %v", filepath)
	return err
}

type getPhotoListStruct struct {
	Response struct {
		Count int `json:"count"`
		Items []struct {
			AlbumID int  `json:"album_id"`
			Date    int  `json:"date"`
			ID      int  `json:"id"`
			OwnerID int  `json:"owner_id"`
			HasTags bool `json:"has_tags"`
			PostID  int  `json:"post_id"`
			Sizes   []struct {
				Height int    `json:"height"`
				URL    string `json:"url"`
				Type   string `json:"type"`
				Width  int    `json:"width"`
			} `json:"sizes"`
			Text string `json:"text"`
		} `json:"items"`
	} `json:"response"`
}
