package groups

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"bot.my/utils"
	"bot.my/vk/users"
)

//GetMsg получение сообщений из группы vk https://vk.com/dev/bots_docs
func GetMsg(vkMsgChan chan string, getErrChan chan error, keyAccesVk, passAPIVk, versinonAPIVk string) {
	//Выводим в чат количество подписчиков для каждой группы
	go func() {
		for {
			dt := time.Now()
			b := fmt.Sprintf(dt.Format("15"))
			if b == "19" {
				groupsMap, err := utils.GetGroupsFromDB() //мапа с группами
				if err != nil {
					getErrChan <- err
				}
				for groupID, value := range groupsMap {
					count, err := GetMembers(groupID, versinonAPIVk, value[0])
					if err != nil {
						getErrChan <- err
					}
					vkMsgChan <- fmt.Sprintf("В группе %v %v подписчиков", value[4], count)
				}
				time.Sleep(2 * time.Hour)
			}
			time.Sleep(30 * time.Minute)
		}
	}()
	// count, err := GetMembers(groupID, versinonAPIVk, token)
	http.HandleFunc("/callback/"+keyAccesVk, func(w http.ResponseWriter, r *http.Request) {
		funcName := "GetMsg" //Имя фонкции для удобного поиска в случае ошибки
		fmt.Fprintf(w, "ок")
		ips := fmt.Sprintf("X-FORWARDED-FOR: %v, X-REAL-IP: %v, RemoteAddr: %v", r.Header.Get("X-REAL-IP"), r.Header.Get("X-FORWARDED-FOR"), r.RemoteAddr)
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			err = errors.New(ips + " " + funcName + " ioutil.ReadAll: " + err.Error())
			getErrChan <- err
			return
		}
		defer r.Body.Close()
		vh := vkWebhoockStruct{}
		err = json.Unmarshal(body, &vh)
		if err != nil {
			err = errors.New(ips + " " + funcName + " Парсинг боди входящего хука: " + err.Error() + "Боди: " + string(body))
			getErrChan <- err
			return
		}
		// log.Printf("Боди по группе: %v", string(body))
		var mapMsgType = map[string]string{ //Мапа для расшифровки всех типов
			"message_new":            "Входящее сообщение",
			"message_reply":          "новое исходящее сообщение",
			"message_edit":           "редактирование сообщения",
			"message_allow":          "подписка на сообщения от сообщества",
			"message_deny":           "новый запрет сообщений от сообщества",
			"message_typing_state":   "статус набора текста",
			"message_event":          "действие с сообщением. Используется для работы с Callback-кнопками",
			"photo_new":              "добавление фотографии",
			"photo_comment_new":      "добавление комментария к фотографии",
			"photo_comment_edit":     "редактирование комментария к фотографии",
			"photo_comment_restore":  "восстановление комментария к фотографии",
			"photo_comment_delete":   "удаление комментария к фотографии",
			"audio_new":              "добавление аудио",
			"video_new":              "добавление видео",
			"video_comment_new":      "комментарий к видео",
			"video_comment_edit":     "редактирование комментария к видео",
			"video_comment_restore":  "восстановление комментария к видео",
			"video_comment_delete":   "удаление комментария к видео",
			"wall_post_new":          "запись на стене",
			"wall_repost":            "репост записи из сообщества",
			"wall_reply_new":         "добавление комментария на стене",
			"wall_reply_edit":        "редактирование комментария на стене",
			"wall_reply_restore":     "восстановление комментария на стене",
			"wall_reply_delete":      "удаление комментария на стене",
			"like_add":               "Событие о новой отметке \"Мне нравится\"",
			"like_remove":            "Событие о снятии отметки \"Мне нравится\"",
			"board_post_new":         "создание комментария в обсуждении",
			"board_post_edit":        "редактирование комментария",
			"board_post_restore":     "восстановление комментария",
			"board_post_delete":      "удаление комментария в обсуждении",
			"market_comment_new":     "новый комментарий к товару",
			"market_comment_edit":    "редактирование комментария к товару",
			"market_comment_restore": "восстановление комментария к товару",
			"market_comment_delete":  "удаление комментария к товару",
			"market_order_new":       "новый заказ",
			"market_order_edit":      "редактирование заказа",
			"group_leave":            "удаление участника из сообщества",
			"group_join":             "добавление участника или заявки на вступление в сообщество",
			"user_block":             "добавление пользователя в чёрный список",
			"user_unblock":           "удаление пользователя из чёрного списка",
			"poll_vote_new":          "добавление голоса в публичном опросе",
			"group_officers_edit":    "редактирование списка руководителей",
			"group_change_settings":  "изменение настроек сообщества",
			"group_change_photo":     "изменение главного фото",
			"vkpay_transaction":      "платёж через VK Pay",
			"app_payload":            "Событие в VK Mini Apps",
		}
		msgType := vh.Type                                               //Тип уведомления
		userID := strconv.FormatInt(int64(vh.Object.Message.FromID), 10) //ИД пользователя
		msg := vh.Object.Message.Text                                    //Сообщение пользователя
		if msg == "" {                                                   //Структура разная в разнах для разных сообщений (личка или нет)
			msg = vh.Object.Text
		}
		if msg != "" { //Если текст передается то отправим его в чат
			msg = "Текст: " + msg
		}
		// log.Printf("userID_0: %v", userID)
		if userID == "0" {
			userID = strconv.FormatInt(int64(vh.Object.FromID), 10)
		}
		// log.Printf("userID_1: %v", userID)
		groupID := strconv.FormatInt(int64(vh.GroupID), 10) //ИД группы
		if mapMsgType[msgType] == "" {                      //Проверяем, если поле еще не добавлено, то выводим ошибку
			err = fmt.Errorf("Не хватает типа уведомления %v в мапе mapMsgType", msgType)
			getErrChan <- err
			return
		}
		if vh.Object.FromID < 0 { //Если отрицательное значение то отправлено от имени группы, скипаем
			// log.Printf("Запись от имени группы %v", userID)
			return
		}
		if msgType == "group_join" || msgType == "group_leave" { //Скипаем уведомления о вступлении и покидании группы
			return
		}
		if msgType == "like_add" || msgType == "like_remove" { //Скипаем добавление и удаление лайков
			return
		}
		//Получаем имя пользвателя и группы
		if userID == "0" {
			err = fmt.Errorf("userID нулевый. Боди %v", string(body))
			getErrChan <- err
			return
		}
		firstName, lastName, _, _, err := users.GetInfo(userID, keyAccesVk, versinonAPIVk, "", false, getErrChan)
		if err != nil {
			err = fmt.Errorf("Не удалось получить имя пользвателя по id %v: %v. Функция %v", userID, err, funcName)
			getErrChan <- err
			return
		}
		groupName, _, _, _, err := users.GetInfo(groupID, keyAccesVk, versinonAPIVk, "", true, getErrChan)
		if err != nil {
			err = fmt.Errorf("Не удалось получить имя группы по id %v: %v", userID, err)
			getErrChan <- err
			return
		}
		//Отправляем в канал сообщение
		vkMsgChan <- fmt.Sprintf("Пользователь %v %v в группе %v произвел %v %v", firstName, lastName, groupName, mapMsgType[msgType], msg)
	})
}

//vkWebhoockStruct Структура получения сообщений из вебхука
type vkWebhoockStruct struct {
	Type   string `json:"type"`
	Object struct {
		Message struct {
			Date                  int           `json:"date"`
			FromID                int           `json:"from_id"`
			ID                    int           `json:"id"`
			Out                   int           `json:"out"`
			PeerID                int           `json:"peer_id"`
			Text                  string        `json:"text"`
			ConversationMessageID int           `json:"conversation_message_id"`
			FwdMessages           []interface{} `json:"fwd_messages"`
			Important             bool          `json:"important"`
			RandomID              int           `json:"random_id"`
			Attachments           []interface{} `json:"attachments"`
			IsHidden              bool          `json:"is_hidden"`
		} `json:"message"`
		ClientInfo struct {
			ButtonActions  []string `json:"button_actions"`
			Keyboard       bool     `json:"keyboard"`
			InlineKeyboard bool     `json:"inline_keyboard"`
			Carousel       bool     `json:"carousel"`
			LangID         int      `json:"lang_id"`
		} `json:"client_info"`
		ID           int           `json:"id"`
		FromID       int           `json:"from_id"`
		PostID       int           `json:"post_id"`
		OwnerID      int           `json:"owner_id"`
		ParentsStack []interface{} `json:"parents_stack"`
		Date         int           `json:"date"`
		Text         string        `json:"text"`
		Thread       struct {
			Count int `json:"count"`
		} `json:"thread"`
		PostOwnerID int `json:"post_owner_id"`
	} `json:"object"`
	GroupID int    `json:"group_id"`
	EventID string `json:"event_id"`
	Secret  string `json:"secret"`
}
