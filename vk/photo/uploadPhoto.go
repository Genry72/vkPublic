package photo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"bot.my/utils"
)

//UploadPhoto Загружает фото на аватарку
func UploadPhoto(cloneID, versinonAPIVk, token, filePath, myID string, vkErrUserChan chan error) (err error) {
	err = utils.Antiban(myID, "UploadPhoto", 300)
	if err != nil {
		vkErrUserChan <- err
	}
	uploadURL, err := firstUpload(versinonAPIVk, token)
	if err != nil {
		return err
	}
	server, photo, hash, err := secondUpload(uploadURL, filePath)
	if err != nil {
		return err
	}
	err = thirdUpload(versinonAPIVk, token, hash, photo, server)
	if err != nil {
		return err
	}
	//Добавляем запись в историчну таблицу о выполнении запроса по добавлению авы
	err = utils.InsertToDB(fmt.Sprintf("INSERT INTO %v (hist_id, hist_type, user_id, navi_date, clone_id) VALUES('%v', '%v', '%v', '%v', '%v');", "history", time.Now().UnixNano(), "Добавление аватарки", myID, "NOW()", cloneID))
	if err != nil {
		vkErrUserChan <- err
	}
	log.Printf("Для пользователя %v применена ава %v", myID, filePath)
	return err
}

//1. Получаем урл для последующей загрузки
func firstUpload(versinonAPIVk, token string) (uploadURL string, err error) {
	url := "https://api.vk.com/method/photos.getOwnerPhotoUploadServer?v=" + versinonAPIVk + "&access_token=" + token
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return uploadURL, err
	}
	res, err := client.Do(req)
	if err != nil {
		return uploadURL, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return uploadURL, err
	}
	if strings.Contains(string(body), "failed") == true || strings.Contains(string(body), "error") {
		err = fmt.Errorf("Ошибка получения url для последующей загрузки фото: %v", string(body))
		return uploadURL, err
	}
	m := firstUploadStruct{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		err = fmt.Errorf("Ошибка парсинга ответа на запрос firstUpload: %v Боди: %v", err, string(body))
		return uploadURL, err
	}
	return m.Response.UploadURL, err
}

//2. Загружаем фото на сервер
func secondUpload(uploadURL, filePath string) (server int, photo, hash string, err error) {
	url := uploadURL
	method := "POST"
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	file, err := os.Open(filePath)
	if err != nil {
		return server, photo, hash, err
	}
	defer file.Close()
	part1,
		err := writer.CreateFormFile("photo", filepath.Base(filePath))
	if err != nil {
		return server, photo, hash, err
	}
	_, err = io.Copy(part1, file)
	if err != nil {
		return server, photo, hash, err
	}
	err = writer.Close()
	if err != nil {
		return server, photo, hash, err
	}
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		return server, photo, hash, err
	}
	req.Header.Add("Content-Type", "multipart/form-data")
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		return server, photo, hash, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return server, photo, hash, err
	}
	if res.StatusCode != 200 {
		err = fmt.Errorf("Ошибка при загрузке фото на сервер " + res.Status + " " + string(body))
		return server, photo, hash, err
	}
	if strings.Contains(string(body), "failed") == true || strings.Contains(string(body), "error") {
		err = fmt.Errorf("Ошибка при загрузке фото на сервер: %v", string(body))
		return server, photo, hash, err
	}
	m := secondUploadStruct{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		err = fmt.Errorf("Ошибка парсинга ответа на запрос secondUpload: %v Боди: %v", err, string(body))
		return server, photo, hash, err
	}
	return m.Server, m.Photo, m.Hash, err
}

//3. Установка аватарки
func thirdUpload(versinonAPIVk, token, hash, photo string, server int) (err error) {
	// url := "https://api.vk.com/method/photos.saveOwnerPhoto?v=" + versinonAPIVk + "&access_token=" + token + "&server=" + strconv.FormatInt(int64(server), 10) + "&hash=" + hash + "&photo=" + photo
	url := "https://api.vk.com/method/photos.saveOwnerPhoto"
	payload := strings.NewReader("v=" + versinonAPIVk + "0&access_token=" + token + "&server=" + strconv.FormatInt(int64(server), 10) + "&hash=" + hash + "&photo=" + photo)
	method := "POST"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return err
	}
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
		err = fmt.Errorf("Ошибка при установке аватарки " + res.Status + " " + string(body))
		return err
	}
	if strings.Contains(string(body), "failed") == true || strings.Contains(string(body), "error") {
		err = fmt.Errorf("Ошибка при установке аватарки: %v", string(body))
		return err
	}
	return err
}

type firstUploadStruct struct {
	Response struct {
		UploadURL string `json:"upload_url"`
	} `json:"response"`
}

type secondUploadStruct struct {
	Server      int    `json:"server"`
	Photo       string `json:"photo"`
	Mid         int    `json:"mid"`
	Hash        string `json:"hash"`
	MessageCode int    `json:"message_code"`
	ProfileAid  int    `json:"profile_aid"`
}
