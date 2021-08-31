module main

go 1.15

replace bot.my/tamtam => ./tamtam

replace bot.my/telegram => ./telegram

replace bot.my/utils => ./utils

replace bot.my/vk => ./vk

replace bot.my/vk/users => ./vk/users

replace bot.my/vk/photo => ./vk/photo

replace bot.my/vk/groups => ./vk/groups

require (
	bot.my/tamtam v0.0.0-00010101000000-000000000000
	bot.my/telegram v0.0.0-00010101000000-000000000000
	bot.my/utils v0.0.0-00010101000000-000000000000
	bot.my/vk/groups v0.0.0-00010101000000-000000000000
	bot.my/vk/photo v0.0.0-00010101000000-000000000000 // indirect
	bot.my/vk/users v0.0.0-00010101000000-000000000000
	github.com/PuerkitoBio/goquery v1.6.0 // indirect
	github.com/lib/pq v1.9.0 // indirect
	github.com/onsi/ginkgo v1.14.2 // indirect
	github.com/sclevine/agouti v3.0.0+incompatible // indirect
	github.com/sirupsen/logrus v1.7.0
)
