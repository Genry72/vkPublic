package groups

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
)

//UploadFotoToWall Загружает фото в ленту группы
func UploadFotoToWall(versinonAPIVk, token, filePath, groupID string) (err error) {
	//1. Получаем урл для последующей загрузки
	uploadURL, err := firstUploadToWall(versinonAPIVk, token, groupID)
	if err != nil {
		return err
	}
	// 2. Загружаем фото на сервер
	server, photo, hash, err := secondUploadToWall(uploadURL, filePath)
	if err != nil {
		return err
	}
	// 3. Получение данных по загруженной фотке
	ownerID, photoID, err := thirdUploadToWall(versinonAPIVk, token, hash, photo, groupID, server)
	if err != nil {
		return err
	}
	// wallPost отправляем сообщение с вложением (без текста). От лица группы. Пользователь должен быть админом.
	err = wallPost(token, groupID, ownerID, photoID, versinonAPIVk)
	if err != nil {
		return err
	}
	return err
}

//1. Получаем урл для последующей загрузки
func firstUploadToWall(versinonAPIVk, token, grouID string) (uploadURL string, err error) {
	url := "https://api.vk.com/method/photos.getWallUploadServer?v=" + versinonAPIVk + "&access_token=" + token + "&group_id=" + grouID
	// fmt.Printf("1. %v\n", url)
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
func secondUploadToWall(uploadURL, filePath string) (server int, photo, hash string, err error) {
	url := uploadURL
	// fmt.Printf("2. %v\n", url)
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

//3. Получение данных по загруженной фотке
func thirdUploadToWall(versinonAPIVk, token, hash, photo, grouID string, server int) (ownerID, photoID string, err error) {
	// url := "https://api.vk.com/method/photos.saveOwnerPhoto?v=" + versinonAPIVk + "&access_token=" + token + "&server=" + strconv.FormatInt(int64(server), 10) + "&hash=" + hash + "&photo=" + photo
	url := "https://api.vk.com/method/photos.saveWallPhoto"
	// fmt.Printf("3. %v\n", url)
	payload := strings.NewReader("v=" + versinonAPIVk + "0&access_token=" + token + "&server=" + strconv.FormatInt(int64(server), 10) + "&hash=" + hash + "&photo=" + photo + "&group_id=" + grouID)
	method := "POST"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return ownerID, photoID, err
	}
	res, err := client.Do(req)
	if err != nil {
		return ownerID, photoID, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return ownerID, photoID, err
	}
	if res.StatusCode != 200 {
		err = fmt.Errorf("Ошибка при установке аватарки %v %v Боди: %v ", res.Status, string(body), payload)
		return ownerID, photoID, err
	}
	if strings.Contains(string(body), "failed") == true || strings.Contains(string(body), "error") {
		err = fmt.Errorf("Ошибка при установке аватарки %v", string(body))
		return ownerID, photoID, err
	}
	m := thirdUploadStruct{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		err = fmt.Errorf("Ошибка парсинга ответа на запрос secondUpload: %v Боди: %v", err, string(body))
		return ownerID, photoID, err
	}
	return fmt.Sprintf("%v", m.Response[0].OwnerID), fmt.Sprintf("%v", m.Response[0].ID), err
}

//wallPost отправляем сообщение с вложением (без текста). От лица группы. Пользователь должен быть админом.
func wallPost(token, grouID, ownerID, photoID, versinonAPIVk string) (err error) {
	url := "https://api.vk.com/method/wall.post?v=" + versinonAPIVk + "&access_token=" + token + "&owner_id=-" + grouID + "&attachments=photo" + ownerID + "_" + photoID + "&scope=wall&from_group=1"
	// fmt.Printf("4. %v\n", url)
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	_, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	log.Printf("Загружена фотка в ленту группы %v", grouID)
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

type thirdUploadStruct struct {
	Response []struct {
		AlbumID   int    `json:"album_id"`
		Date      int    `json:"date"`
		ID        int    `json:"id"`
		OwnerID   int    `json:"owner_id"`
		HasTags   bool   `json:"has_tags"`
		AccessKey string `json:"access_key"`
		Sizes     []struct {
			Height int    `json:"height"`
			URL    string `json:"url"`
			Type   string `json:"type"`
			Width  int    `json:"width"`
		} `json:"sizes"`
		Text string `json:"text"`
	} `json:"response"`
}
