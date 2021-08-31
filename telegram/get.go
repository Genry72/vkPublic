package telegram

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"io/ioutil"
	"mime/multipart"
	"net/http"
)

//GetMsg открывает https. Серт получен https://certbot.eff.org/lets-encrypt/ubuntufocal-webproduct. Корневой сайт https://letsencrypt.org/ru/getting-started/
func GetMsg(msgChan chan map[int64]string, getErrChan chan error, tokenTelegram string) {
	// err = setWebHookTelegram(tokenTelegram, urlWebHookTelegram)
	// if err != nil {
	// 	getErrChan <- err
	// }
	http.HandleFunc("/"+tokenTelegram, func(w http.ResponseWriter, r *http.Request) {
		funcName := "GetMsg" //Имя фонкции для удобного поиска в случае ошибки
		fmt.Fprintf(w, "OK")
		ips := fmt.Sprintf("X-FORWARDED-FOR: %v, X-REAL-IP: %v, RemoteAddr: %v", r.Header.Get("X-REAL-IP"), r.Header.Get("X-FORWARDED-FOR"), r.RemoteAddr)
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			err = errors.New(ips + " " + funcName + " ioutil.ReadAll: " + err.Error())
			getErrChan <- err
			return
		}
		defer r.Body.Close()
		ms := GetmsgTelegramStruct{}
		err = json.Unmarshal(body, &ms)
		fmt.Println(string(body))
		if err != nil {
			err = errors.New(ips + " " + funcName + " Парсинг боди входящего хука: " + err.Error() + "Боди: " + string(body))
			getErrChan <- err
			return
		}
		m := make(map[int64]string) //Мапа для передачи сообщения (ключ chatID, значение сообщение - msg)
		msg := ms.Message.Text
		chatID := ms.Message.Chat.ID
		m[chatID] = msg
		if ms.Message.From.IsBot == false {
			msgChan <- m
		}

	})
	// return err
}

func setWebHookTelegram(tokenTelegram, urlWebHookTelegram string) (err error) {
	funcName := "setWebHookTelegram" //Имя фонкции для удобного поиска в случае ошибки
	url := "https://api.telegram.org/bot" + tokenTelegram + "/setWebhook"
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	err = writer.WriteField("url", urlWebHookTelegram)
	if err != nil {
		return err
	}
	// file, errFile2 := os.Open(publicCert)
	// defer file.Close()
	// part2,
	// 	errFile2 := writer.CreateFormFile("certificate", filepath.Base(publicCert))
	// _, errFile2 = io.Copy(part2, file)
	// if errFile2 != nil {
	// 	return err
	// }
	// err = writer.Close()
	// if err != nil {
	// 	return err
	// }
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, payload)

	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		err = errors.New(funcName + " Запрос на регистрацию webhook " + res.Status + " " + string(body))
		return err
	}
	t := setWebHookTelegramStruct{}
	err = json.Unmarshal(body, &t)
	if err != nil {
		err = errors.New(funcName + " Парсинг боди запроса на регистрацию webhook " + string(body))
		return err
	}
	if t.Ok == false || t.Result == false {
		err = errors.New(funcName + " Регистрация webhook не удалась " + string(body))
		return err
	}
	return err
}

//GetmsgTelegramStruct Структура входящей вебхуки
type GetmsgTelegramStruct struct {
	UpdateID int `json:"update_id"`
	Message  struct {
		MessageID int `json:"message_id"`
		From      struct {
			ID           int    `json:"id"`
			IsBot        bool   `json:"is_bot"`
			FirstName    string `json:"first_name"`
			LastName     string `json:"last_name"`
			Username     string `json:"username"`
			LanguageCode string `json:"language_code"`
		} `json:"from"`
		Chat struct {
			ID                          int64  `json:"id"`
			Title                       string `json:"title"`
			Type                        string `json:"type"`
			AllMembersAreAdministrators bool   `json:"all_members_are_administrators"`
		} `json:"chat"`
		Date int    `json:"date"`
		Text string `json:"text"`
	} `json:"message"`
}

type setWebHookTelegramStruct struct {
	Ok          bool   `json:"ok"`
	Result      bool   `json:"result"`
	Description string `json:"description"`
}
