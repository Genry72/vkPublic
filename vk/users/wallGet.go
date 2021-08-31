package users

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"bot.my/utils"
)

//WallGet Возвращает список записей со стены пользователя или сообщества.
func WallGet(versinonAPIVk, token, id, myID string, vkErrUserChan chan error) (map[string][]string, error) {
	log.Printf("Запрашиваем список записей со стены пользователя %v из-под пользователя %v", id, myID)
	err := utils.Antiban(myID, "WallGet", 300)
	if err != nil {
		vkErrUserChan <- err
	}
	mapaForLikes := make(map[string][]string) // Мапа объектов для последующих лайков. Ключь id объекта, значение список: тип объекта и количество лайков
	url := "https://api.vk.com/method/wall.get?count=100&v=" + versinonAPIVk + "&access_token=" + token + "&filter=owner&owner_id=" + id
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return mapaForLikes, err
	}
	res, err := client.Do(req)
	if err != nil {
		return mapaForLikes, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return mapaForLikes, err
	}
	if res.StatusCode != 200 {
		err = fmt.Errorf("Ошибка получения списка записей со стены пользователя или сообщества " + res.Status + " " + string(body))
		return mapaForLikes, err
	}
	if strings.Contains(string(body), "failed") == true || strings.Contains(string(body), "\"error\"") {
		err = fmt.Errorf("Ошибка получения списка записей со стены пользователя или сообщества: %v", string(body))
		return mapaForLikes, err
	}
	m := wallGetStruct{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		err = fmt.Errorf("Ошибка парсинга ответа на запрос WallGet: %v Боди: %v", err, string(body))
		return mapaForLikes, err
	}
	for _, item := range m.Response.Items {
		if len(item.Attachments) == 0 { //Пропускаем, если это репост, нам нужны сообщения от самого пользоватлея
			continue
		}
		if item.FromID < 0 { //Скипаем если новость от группы
			continue
		}
		if item.Attachments[0].Photo.ID != 0 {
			objectSlice := []string{item.Attachments[0].Type, fmt.Sprintf("%v", item.Likes.Count)} //Списсок для результирующей мапы
			mapaForLikes[fmt.Sprintf("%v", item.Attachments[0].Photo.ID)] = objectSlice
		}
		if item.Attachments[0].Video.ID != 0 {
			objectSlice := []string{item.Attachments[0].Type, fmt.Sprintf("%v", item.Likes.Count)} //Списсок для результирующей мапы
			mapaForLikes[fmt.Sprintf("%v", item.Attachments[0].Video.ID)] = objectSlice
		}
		if item.Attachments[0].Doc.ID != 0 {
			objectSlice := []string{item.Attachments[0].Type, fmt.Sprintf("%v", item.Likes.Count)} //Списсок для результирующей мапы
			mapaForLikes[fmt.Sprintf("%v", item.Attachments[0].Doc.ID)] = objectSlice
		}
	}
	//Добавляем запись в историчну таблицу о выполнении запроса
	err = utils.InsertToDB(fmt.Sprintf("INSERT INTO %v VALUES('%v', '%v', '%v', %v);", "history", time.Now().UnixNano(), "WallGet", myID, "NOW()"))
	if err != nil {
		vkErrUserChan <- err
	}
	log.Printf("Запросили список записей со стены пользователя %v из-под пользователя %v", id, myID)
	return mapaForLikes, err
}

type wallGetStruct struct {
	Response struct {
		Count int `json:"count"`
		Items []struct {
			ID          int    `json:"id"`
			FromID      int    `json:"from_id"`
			OwnerID     int    `json:"owner_id"`
			Date        int    `json:"date"`
			Type        string `json:"type"`
			PostType    string `json:"post_type"`
			Text        string `json:"text"`
			Attachments []struct {
				Type  string `json:"type"`
				Photo struct {
					ID int `json:"id"`
				} `json:"photo"`
				Video struct {
					ID int `json:"id"`
				} `json:"Video"`
				Doc struct {
					ID int `json:"id"`
				} `json:"doc"`
			} `json:"attachments"`
			PostSource struct {
				Type string `json:"type"`
			} `json:"post_source"`
			Comments struct {
				Count         int  `json:"count"`
				CanPost       int  `json:"can_post"`
				GroupsCanPost bool `json:"groups_can_post"`
			} `json:"comments"`
			Likes struct {
				Count      int `json:"count"`
				UserLikes  int `json:"user_likes"`
				CanLike    int `json:"can_like"`
				CanPublish int `json:"can_publish"`
			} `json:"likes"`
			Reposts struct {
				Count        int `json:"count"`
				UserReposted int `json:"user_reposted"`
			} `json:"reposts"`
			Views struct {
				Count int `json:"count"`
			} `json:"views"`
			IsFavorite bool `json:"is_favorite"`
		} `json:"items"`
		NextFrom string `json:"next_from"`
	} `json:"response"`
}
