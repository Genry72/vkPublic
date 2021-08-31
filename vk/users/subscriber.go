package users

import (
	"log"
)

//Subscriber подписываеся на указанные группы
func Subscriber(sex, versinonAPIVk, token, myID string, vkErrUserChan chan error) {
	//Группы подписок для мужика
	menSubscrybers := []string{
		// "39153701",  //Злой Гений
		// "122149207", //Экспериментатор https://vk.com/experimentato
		// "49297041",  //ЕБ*НУТЬСЯ https://vk.com/fucking_bitching_21_plus
		// "125011298", //Иллюзия обмана https://vk.com/illusionobman
		"197513246", //В Ж(*)ПУ РеАлЬнОсТь https://vk.com/fuck_realnost
	}
	var subscrybers []string
	if sex == "Мужчина" {
		subscrybers = menSubscrybers
	}
	//Получаем все подписки пользователя
	userSubscribers, err := GetSubscriptions(versinonAPIVk, token, myID, myID, vkErrUserChan)
	if err != nil {
		vkErrUserChan <- err
	}
	//Получаем список не достающих подписок
	var newSubs []string
	for _, subs := range subscrybers { //Идем по группе подписок
		for i, subsU := range userSubscribers { //Идем по списку подписок пользоватлея
			if subs == subsU { //Нашли совпадение - скипаем
				break
			}
			if i == len(userSubscribers)-1 {
				newSubs = append(newSubs, subs)
			}
		}
		if len(userSubscribers) == 0 { //Если у пользоватлея совсем нет подписок, то подписываемся на все
			newSubs = subscrybers
		}
	}
	//Подписываемся на все подписки из списка
	for _, groupID := range newSubs {
		err = GroupsJoin(versinonAPIVk, token, groupID, myID, vkErrUserChan)
		if err != nil {
			vkErrUserChan <- err
			continue
		}
	}
	log.Printf("Пользователь %v подписан на все подписки", myID)
}
