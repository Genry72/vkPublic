package utils

import (
	"math/rand"
	"time"
)

//GeneratePWD генерирует случайный пароль
func GeneratePWD() (string) {
	rand.Seed(time.Now().UnixNano())
	digits := "0123456789"
	specials := "!@#$%^&*()_+-"
	all := "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		digits + specials
	length := 12
	buf := make([]byte, length)
	buf[0] = digits[rand.Intn(len(digits))]
	buf[1] = specials[rand.Intn(len(specials))]
	for i := 2; i < length; i++ {
		buf[i] = all[rand.Intn(len(all))]
	}
	rand.Shuffle(len(buf), func(i, j int) {
		buf[i], buf[j] = buf[j], buf[i]
	})
	str := string(buf) // Например "3i[g0|)z"
	return str
}
