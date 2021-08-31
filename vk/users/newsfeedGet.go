package users

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"bot.my/utils"
)

// NewsfeedGet получаем посты из ленты. Возврашает сипсок объектов для репоста
func NewsfeedGet(versinonAPIVk, token, myID string, vkErrUserChan chan error) (postsIDs []string, err error) {
	log.Printf("Получаем список постов из ленты пользоватлея %v", myID)
	err = utils.Antiban(myID, "NewsfeedGet", 300)
	if err != nil {
		vkErrUserChan <- err
	}
	//Проверяем что за сегодня было меньше 10 репостов
	countRepostSTR, err := utils.GetOneRowDB(fmt.Sprintf("select count(*) from history t where t.user_id = '%v' and t.navi_date > 'today' and t.hist_type = 'Репост'", myID))
	if err != nil {
		return postsIDs, err
	}
	countRepost, err := strconv.Atoi(countRepostSTR)
	if err != nil {
		return postsIDs, err
	}
	if countRepost > 10 {
		log.Println("Количество репостов больше 10")
		time.Sleep(30 * time.Minute)
		return postsIDs, err
	}
	url := fmt.Sprintf("https://api.vk.com/method/newsfeed.get?count=100&v=" + versinonAPIVk + "&access_token=" + token)
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return postsIDs, err
	}
	res, err := client.Do(req)
	if err != nil {
		return postsIDs, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return postsIDs, err
	}
	if res.StatusCode != 200 {
		err = fmt.Errorf("Ошибка получения списка постов для группы " + res.Status + " " + string(body))
		return postsIDs, err
	}
	if strings.Contains(string(body), "failed") == true || strings.Contains(string(body), "error_code") {
		err = fmt.Errorf("Ошибка получения списка постов для группы: %v", string(body))
		return postsIDs, err
	}
	m := newsfeedGetStruct{}
	err = json.Unmarshal(body, &m)
	if err != nil {
		err = fmt.Errorf("Ошибка парсинга ответа на запрос NewsfeedGet: %v Боди: %v", err, string(body))
		return postsIDs, err
	}
	for _, post := range m.Response.Items {
		if len(post.Attachments) == 0 {
			continue
		}
		if post.Attachments[0].Photo.ID != 0 {
			if post.SourceID > 0 { //Скипаем если новость не от группы
				continue
			}
			postsIDs = append(postsIDs, post.Attachments[0].Type+fmt.Sprintf("%v_%v", post.SourceID, post.Attachments[0].Photo.ID))
		}
		if post.Attachments[0].Video.ID != 0 {
			if post.SourceID > 0 {
				continue
			}
			postsIDs = append(postsIDs, post.Attachments[0].Type+fmt.Sprintf("%v_%v", post.SourceID, post.Attachments[0].Video.ID))
		}
		if post.Attachments[0].Doc.ID != 0 {
			if post.SourceID > 0 {
				continue
			}
			postsIDs = append(postsIDs, post.Attachments[0].Type+fmt.Sprintf("%v_%v", post.SourceID, post.Attachments[0].Doc.ID))
		}
	}
	//Добавляем запись в историчну таблицу о выполнении запроса
	err = utils.InsertToDB(fmt.Sprintf("INSERT INTO %v VALUES('%v', '%v', '%v', %v);", "history", time.Now().UnixNano(), "NewsfeedGet", myID, "NOW()"))
	if err != nil {
		vkErrUserChan <- err
	}
	log.Printf("Получили список постов из ленты пользователя %v", myID)
	return postsIDs, err
}

type newsfeedGetStruct struct {
	Response struct {
		Items []struct {
			SourceID         int    `json:"source_id"`
			Date             int    `json:"date"`
			CanDoubtCategory bool   `json:"can_doubt_category,omitempty"`
			CanSetCategory   bool   `json:"can_set_category,omitempty"`
			PostType         string `json:"post_type,omitempty"`
			Text             string `json:"text,omitempty"`
			MarkedAsAds      int    `json:"marked_as_ads,omitempty"`
			Attachments      []struct {
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
			} `json:"post_source,omitempty"`
			Comments struct {
				Count   int `json:"count"`
				CanPost int `json:"can_post"`
			} `json:"comments,omitempty"`
			Likes struct {
				Count      int `json:"count"`
				UserLikes  int `json:"user_likes"`
				CanLike    int `json:"can_like"`
				CanPublish int `json:"can_publish"`
			} `json:"likes,omitempty"`
			Reposts struct {
				Count        int `json:"count"`
				UserReposted int `json:"user_reposted"`
			} `json:"reposts,omitempty"`
			Views struct {
				Count int `json:"count"`
			} `json:"views,omitempty"`
			IsFavorite       bool   `json:"is_favorite,omitempty"`
			PostID           int    `json:"post_id,omitempty"`
			Type             string `json:"type"`
			PushSubscription struct {
				IsSubscribed bool `json:"is_subscribed"`
			} `json:"push_subscription"`
			TrackCode string `json:"track_code,omitempty"`
			Photos    struct {
				Count int `json:"count"`
				Items []struct {
					AlbumID   int    `json:"album_id"`
					Date      int    `json:"date"`
					ID        int    `json:"id"`
					OwnerID   int    `json:"owner_id"`
					HasTags   bool   `json:"has_tags"`
					AccessKey string `json:"access_key"`
					PostID    int    `json:"post_id"`
					Sizes     []struct {
						Height int    `json:"height"`
						URL    string `json:"url"`
						Type   string `json:"type"`
						Width  int    `json:"width"`
					} `json:"sizes"`
					Text   string `json:"text"`
					UserID int    `json:"user_id"`
					Likes  struct {
						UserLikes int `json:"user_likes"`
						Count     int `json:"count"`
					} `json:"likes"`
					Reposts struct {
						Count        int `json:"count"`
						UserReposted int `json:"user_reposted"`
					} `json:"reposts"`
					Comments struct {
						Count int `json:"count"`
					} `json:"comments"`
					CanComment int `json:"can_comment"`
					CanRepost  int `json:"can_repost"`
				} `json:"items"`
			} `json:"photos,omitempty"`
			ExtID   string `json:"ext_id,omitempty"`
			Friends struct {
				Count int `json:"count"`
				Items []struct {
					UserID int `json:"user_id"`
				} `json:"items"`
			} `json:"friends,omitempty"`
			CopyHistory []struct {
				ID          int    `json:"id"`
				OwnerID     int    `json:"owner_id"`
				FromID      int    `json:"from_id"`
				Date        int    `json:"date"`
				PostType    string `json:"post_type"`
				Text        string `json:"text"`
				Attachments []struct {
					Type  string `json:"type"`
					Photo struct {
						AlbumID   int    `json:"album_id"`
						Date      int    `json:"date"`
						ID        int    `json:"id"`
						OwnerID   int    `json:"owner_id"`
						HasTags   bool   `json:"has_tags"`
						AccessKey string `json:"access_key"`
						PostID    int    `json:"post_id"`
						Sizes     []struct {
							Height int    `json:"height"`
							URL    string `json:"url"`
							Type   string `json:"type"`
							Width  int    `json:"width"`
						} `json:"sizes"`
						Text   string `json:"text"`
						UserID int    `json:"user_id"`
					} `json:"photo"`
				} `json:"attachments"`
				PostSource struct {
					Type string `json:"type"`
				} `json:"post_source"`
			} `json:"copy_history,omitempty"`
		} `json:"items"`
		Profiles []struct {
			FirstName       string `json:"first_name"`
			ID              int    `json:"id"`
			LastName        string `json:"last_name"`
			CanAccessClosed bool   `json:"can_access_closed"`
			IsClosed        bool   `json:"is_closed"`
			Sex             int    `json:"sex"`
			ScreenName      string `json:"screen_name"`
			Photo50         string `json:"photo_50"`
			Photo100        string `json:"photo_100"`
			OnlineInfo      struct {
				Visible  bool `json:"visible"`
				IsOnline bool `json:"is_online"`
				IsMobile bool `json:"is_mobile"`
			} `json:"online_info,omitempty"`
			Online           int  `json:"online"`
			IsService        bool `json:"is_service,omitempty"`
			CanInviteToChats bool `json:"can_invite_to_chats"`
			OnlineMobile     int  `json:"online_mobile,omitempty"`
			OnlineApp        int  `json:"online_app,omitempty"`
		} `json:"profiles"`
		Groups []struct {
			ID           int    `json:"id"`
			Name         string `json:"name"`
			ScreenName   string `json:"screen_name"`
			IsClosed     int    `json:"is_closed"`
			Type         string `json:"type"`
			IsAdmin      int    `json:"is_admin"`
			IsMember     int    `json:"is_member"`
			IsAdvertiser int    `json:"is_advertiser"`
			Photo50      string `json:"photo_50"`
			Photo100     string `json:"photo_100"`
			Photo200     string `json:"photo_200"`
		} `json:"groups"`
		NextFrom string `json:"next_from"`
	} `json:"response"`
}
