package tamtam

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

//GetMsg принимает по подписке сообщения https://dev.tamtam.chat/#operation/subscribe. Откравает порт и парсит боди
func GetMsg(msgChan chan map[int64]string, getErrChan chan error, tokenTamTam, urlWebHookTamTam string) (err error) {
	funcName := "GetMsg" //Имя фонкции для удобного поиска в случае ошибки
	err = registrWebHookTamTam(tokenTamTam, urlWebHookTamTam)
	if err != nil {
		return err
	}
	http.HandleFunc("/"+tokenTamTam, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
		ips := fmt.Sprintf("X-FORWARDED-FOR: %v, X-REAL-IP: %v, RemoteAddr: %v", r.Header.Get("X-REAL-IP"), r.Header.Get("X-FORWARDED-FOR"), r.RemoteAddr)
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			err = errors.New(ips + " " + funcName + " ioutil.ReadAll " + err.Error())
			getErrChan <- err
			return
		}
		defer r.Body.Close()
		ms := GetmsgStruct{}
		err = json.Unmarshal(body, &ms)
		if err != nil {
			err = errors.New(ips + " " + funcName + " Парсинг боди входящего хука: " + err.Error() + " Боди: " + string(body))
			getErrChan <- err
			return
		}
		m := make(map[int64]string) //Мапа для передачи сообщения (ключ chatID, значение сообщение - msg)
		msg := ms.Message.Body.Text
		if msg == "" {
			msg = ms.Message.Body.Attachments[0].Payload.URL
		}
		chatID := ms.Message.Recipient.ChatID
		m[chatID] = msg
		if ms.Message.Sender.IsBot == false {
			msgChan <- m
		}

	})
	return err
}
func registrWebHookTamTam(tokenTamTam, urlWebHookTamTam string) (err error) {
	funcName := "registrWebHookTamTam" //Имя фонкции для удобного поиска в случае ошибки
	url := "https://botapi.tamtam.chat/subscriptions?access_token=" + tokenTamTam
	payload := strings.NewReader("{\n\"url\": \"" + urlWebHookTamTam + "\"\n}")
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Cookie", "web_ui_lang=ru")
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
	t := registerURLTamtamStruct{}
	err = json.Unmarshal(body, &t)
	if err != nil {
		err = errors.New(funcName + " Парсинг боди запроса на регистрацию webhook " + string(body))
		return err
	}
	if t.Success == false {
		err = errors.New(funcName + " Регистрация webhook не удалась " + string(body))
		return err
	}
	return err
}

//GetmsgStruct структура входящих сообщений по webhook
type GetmsgStruct struct {
	Timestamp int64 `json:"timestamp"`
	Message   struct {
		Sender struct {
			UserID           int64  `json:"user_id"`
			Name             string `json:"name"`
			IsBot            bool   `json:"is_bot"`
			LastActivityTime int    `json:"last_activity_time"`
		} `json:"sender"`
		Recipient struct {
			ChatID   int64  `json:"chat_id"`
			ChatType string `json:"chat_type"`
		} `json:"recipient"`
		Timestamp int64 `json:"timestamp"`
		Body      struct {
			Mid         string `json:"mid"`
			Seq         int64  `json:"seq"`
			Text        string `json:"text"`
			Attachments []struct {
				Payload struct {
					URL   string `json:"url"`
					Token string `json:"token"`
				} `json:"payload"`
				Title       string `json:"title"`
				Description string `json:"description"`
				ImageURL    string `json:"image_url"`
				Type        string `json:"type"`
			} `json:"attachments"`
		} `json:"body"`
	} `json:"message"`
	UpdateType string `json:"update_type"`
}

//registerURLTamtamStruct структура ответа на запрос регистрации webhook
type registerURLTamtamStruct struct {
	Success bool `json:"success"`
}
