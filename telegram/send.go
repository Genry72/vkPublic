package telegram

import (
	"errors"
	"encoding/json"
	"strconv"
	"bytes"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

//SendMsg отправляет сообщение в телегу
func SendMsg(tokenTelegram, msg string, chatID int64) (err error) {
	url := "https://api.telegram.org/bot" + tokenTelegram + "/sendMessage"
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("chat_id", strconv.FormatInt(chatID, 10))
	_ = writer.WriteField("text", msg)
	err = writer.Close()
	if err != nil {
		return err
	}

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
	t := sendMsgTelegaStruct{}
	err = json.Unmarshal(body, &t)
	if err != nil {
		err = errors.New("Парсинг боди ответа на отправку сообщения " + string(body))
		return err
	}
	if t.Ok == false {
		err = errors.New(string(body))
		return err
	}
	return err
}
//sendMsgTelegaStruct структура ответного боди на отправку сообщения
type sendMsgTelegaStruct struct {
	Ok     bool `json:"ok"`
	Result struct {
		MessageID int `json:"message_id"`
		From      struct {
			ID        int    `json:"id"`
			IsBot     bool   `json:"is_bot"`
			FirstName string `json:"first_name"`
			Username  string `json:"username"`
		} `json:"from"`
		Chat struct {
			ID        int    `json:"id"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Username  string `json:"username"`
			Type      string `json:"type"`
		} `json:"chat"`
		Date int    `json:"date"`
		Text string `json:"text"`
	} `json:"result"`
}
