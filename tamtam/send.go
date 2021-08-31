package tamtam

import (
	"bytes"
	// "crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//SendToChat Функция отправки в tamtam. Принимает Имя чата, либо chatid. В случае если chatName равен "", то получит chatid по имени чата
func SendToChat(accessToken, text string, filePhoto, chatName string, chatID int64) (err error) {
	var tokenPhoto string
	var maxlenText = 3500 //Максимальный размер сообщения для отправки

	// //Задаем прокси
	// proxyStr := "http://vlg-proxy.megafon.ru:3128"
	// proxyURL, err := url.Parse(proxyStr)
	// if err != nil {
	// 	return err
	// }
	// basicAuth := "Basic " + logpassAdLong
	// hdr := http.Header{}
	// hdr.Add("Proxy-Authorization", basicAuth)
	// transport := &http.Transport{
	// 	Proxy:              http.ProxyURL(proxyURL),
	// 	ProxyConnectHeader: hdr,
	// 	TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
	// }

	client := &http.Client{
		// Transport: transport,
		Timeout: time.Second * 10,
	}
	//Если указали имя чата то получаем его id
	if chatName != "" {
		chatID, err = getChatID(accessToken, chatName, client)
		if err != nil {
			return err
		}
	}
	//Если передали путь до файла до получаем его id и url
	if filePhoto != "" {
		urle, phptoID, err := first(accessToken, client)
		if err != nil {
			return err
		}
		//Получаем токен фотки
		tokenPhoto, err = second(urle, filePhoto, phptoID, client)
		if err != nil {
			return err
		}
	}
	if len(text) >= maxlenText { //Первичная проверка на длину сообщения
		textSlice := strings.Split(text, " ") //Бьем пробелами для получения слайса
		var rezultMSG string                  //Обрезанная строка для отправки в чат
		for _, mes := range textSlice {
			if len(rezultMSG)+len(mes+" ") < maxlenText { //Если размер меньше максимального то добавляем из слайса в стринг
				rezultMSG = rezultMSG + " " + mes
			} else { //Если в совокупности получается размер сообщения выше, то отправляем что есть
				err = third(rezultMSG, tokenPhoto, accessToken, chatID, client)
				if err != nil {
					return err
				}
				rezultMSG = mes
			}
		}
		//Отправляем сообщение
		err = third(rezultMSG, tokenPhoto, accessToken, chatID, client)
		if err != nil {
			return err
		}
	} else {
		err = third(text, tokenPhoto, accessToken, chatID, client)
		if err != nil {
			return err
		}
	}
	return err
}

//Получаем список доступных чатов и достаем chatID
func getChatID(accessToken, chatName string, client *http.Client) (chatID int64, err error) {
	funcName := "getChatID" //Имя фонкции для удобного поиска в случае ошибки
	urle := "https://botapi.tamtam.chat/chats?access_token=" + accessToken
	req, err := http.NewRequest("GET", urle, nil)
	if err != nil {
		err = errors.New(funcName + " http.NewRequest: " + err.Error())
		return chatID, err
	}
	res, err := client.Do(req)
	if err != nil {
		err = errors.New(funcName + " client.Do: " + err.Error())
		return chatID, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		err = errors.New(funcName + " ioutil.ReadAll: " + err.Error())
		return chatID, err
	}
	if res.StatusCode != 200 {
		err = errors.New(funcName + " Запрос на получение chatID " + res.Status + " " + string(body))
		return chatID, err
	}
	t := GetChatstruct{}
	err = json.Unmarshal(body, &t)
	if err != nil {
		err = errors.New(funcName + " Парсинг боди запроса на получение списка доступных чатов " + string(body))
		return chatID, err
	}
	for i := 0; i < len(t.Chats); i++ {
		if t.Chats[i].Title == chatName {
			chatID = t.Chats[i].ChatID
		}
		if t.Chats[i].DialogWithUser.Name == chatName {
			chatID = t.Chats[i].ChatID
		}
	}
	return chatID, err
}

//Получение урла для загрузки изображения, возвращает url и id фотки
func first(token string, client *http.Client) (urll, photoIds string, err error) {
	funcName := "first" //Имя фонкции для удобного поиска в случае ошибки
	urle := "https://botapi.tamtam.chat/uploads?access_token=" + token + "&type=image"
	req, err := http.NewRequest("POST", urle, nil)
	if err != nil {
		err = errors.New(funcName + " http.NewRequest: " + err.Error())
		return urll, photoIds, err
	}
	res, err := client.Do(req)
	if err != nil {
		err = errors.New(funcName + " client.Do " + err.Error())
		return urll, photoIds, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		err = errors.New(funcName + " ioutil.ReadAll " + err.Error())
		return urll, photoIds, err
	}
	if res.StatusCode != 200 {
		err = errors.New(funcName + " Получении url для загрузки изображения " + res.Status + " " + string(body))
		return urll, photoIds, err
	}
	t := firstTamTamStruct{}
	err = json.Unmarshal(body, &t)
	if err != nil {
		err = errors.New(funcName + " Парсинг боди получения урла " + string(body))
		return urll, photoIds, err
	}
	//Парсим урл чтобы достать id фото
	u, err := url.Parse(t.URL)
	if err != nil {
		err = errors.New(funcName + " Парсинг урла " + t.URL)
		return urll, photoIds, err
	}

	m, err := url.ParseQuery(u.RawQuery)
	err = errors.New(funcName + " ParseQuery " + err.Error())
	return t.URL, m["photoIds"][0], err
}

//На вход принимает урл и путь к файлу, возвращает токен
func second(uploadURL, filePath, photoid string, client *http.Client) (token string, err error) {
	funcName := "second" //Имя фонкции для удобного поиска в случае ошибки
	urle := uploadURL
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	file, err := os.Open(filePath)
	if err != nil {
		err = errors.New(funcName + " Открытие файла " + filePath + ":" + err.Error())
		return token, err
	}
	defer file.Close()
	part1,
		err := writer.CreateFormFile("data", filepath.Base(filePath))
	if err != nil {
		err = errors.New(funcName + " CreateFormFile " + filePath + ":" + err.Error())
		return token, err
	}
	_, err = io.Copy(part1, file)
	if err != nil {
		err = errors.New(funcName + " io.Copy :" + err.Error())
		return token, err
	}
	err = writer.Close()
	if err != nil {
		err = errors.New(funcName + " writer.Close :" + err.Error())
		return token, err
	}

	req, err := http.NewRequest("POST", urle, payload)
	if err != nil {
		err = errors.New(funcName + " http.NewRequest :" + err.Error())
		return token, err
	}
	req.Header.Add("Content-Type", "multipart/form-data")
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		err = errors.New(funcName + " client.Do :" + err.Error())
		return token, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		err = errors.New(funcName + " ioutil.ReadAll :" + err.Error())
		return token, err
	}
	if res.StatusCode != 200 {
		err = errors.New(funcName + "Отправка запроса на получение токена изображения " + res.Status + " " + string(body))
		return token, err
	}
	t := secondTamTamStruct{}
	err = json.Unmarshal(body, &t)
	if err != nil {
		err = errors.New(funcName + "Парсинг боди запроса на получение токена изображения " + string(body))
		return token, err
	}
	// fmt.Println("token="+t.Token)
	return t.Photos[photoid].Token, err
}

//Отправляем вложение в чат с передачей токена изображения
func third(text, token, accessToken string, chatID int64, client *http.Client) (err error) {
	funcName := "third" //Имя фонкции для удобного поиска в случае ошибки
	urle := "https://botapi.tamtam.chat/messages?access_token=" + accessToken + "&chat_id=" + strconv.FormatInt(chatID, 10)
	method := "POST"
	var m MessageWithBodyStruct
	//Если отправляем сообщение с картинкой, то меняем боди
	if token != "" {
		//Избавляемся от index out of range https://www.linux.org.ru/forum/development/14862594
		m = MessageWithBodyStruct{
			Attachments: []Attachments{
				Attachments{
					Type: "image",
				},
			},
		}
		m.Attachments[0].Payload.Token = token
	}
	m.Text = text
	//Собираем json
	messageBody, err := json.Marshal(m)
	if err != nil {
		err = errors.New(funcName + "Сборка json " + string(m.Text))
		return err
	}
	payload := strings.NewReader(string(messageBody))
	req, err := http.NewRequest(method, urle, payload)
	if err != nil {
		err = errors.New(funcName + "http.NewRequest " + string(m.Text))
		return err
	}
	req.Header.Add("Content-Type", "multipart/form-data")

	res, err := client.Do(req)
	if err != nil {
		err = errors.New(funcName + "client.Do " + string(m.Text))
		return err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		err = errors.New(funcName + "ioutil.ReadAll " + string(m.Text))
		return err
	}
	if res.StatusCode != 200 {
		err = errors.New("Ошибка при отправке сообщения " + res.Status + " " + string(messageBody) + ": " + string(body))
		return err
	}
	return err
}

type firstTamTamStruct struct {
	URL string `json:"url"`
}

type secondTamTamStruct struct {
	URL    string                `json:"url,omitempty"`    // Any external image URL you want to attach
	Token  string                `json:"token,omitempty"`  // Token of any existing attachment
	Photos map[string]PhotoToken `json:"photos,omitempty"` // Tokens were obtained after uploading images
}

//PhotoToken Вложенная мапа в структуру secondTamTamStruct
type PhotoToken struct {
	Token string `json:"token"` // Encoded information of uploaded image
}

//MessageWithBodyStruct Структура для формирования json при отправке в чат с боди
type MessageWithBodyStruct struct {
	Text        string        `json:"text"`
	Attachments []Attachments `json:"attachments"`
}

//Payload ..
type Payload struct {
	Token string `json:"token"`
}

//Attachments ...
type Attachments struct {
	Type    string  `json:"type"`
	Payload Payload `json:"payload"`
}

//GetChatstruct структура получения всех доступных чатов
type GetChatstruct struct {
	Chats []struct {
		ChatID            int64  `json:"chat_id"`
		Type              string `json:"type"`
		Status            string `json:"status"`
		Title             string `json:"title,omitempty"`
		LastEventTime     int64  `json:"last_event_time"`
		ParticipantsCount int    `json:"participants_count"`
		IsPublic          bool   `json:"is_public"`
		OwnerID           int64  `json:"owner_id,omitempty"`
		MessagesCount     int    `json:"messages_count,omitempty"`
		DialogWithUser    struct {
			UserID int64  `json:"user_id"`
			Name   string `json:"name"`
			IsBot  bool   `json:"is_bot"`
		} `json:"dialog_with_user,omitempty"`
	} `json:"chats"`
}
