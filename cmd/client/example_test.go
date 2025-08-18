// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package client_test

import (
	"fmt"

	"github.com/h1067675/shortUrl/cmd/client"
)

func Example() {
	// Создаем клиента и при необходимости прописывааем настройки сервера.
	cl := client.NewClient().SetLevelLogging("debug")

	// Запускаем сервер если еще не запущен.
	cl.StartServer()

	// Ссылка для сокращения:
	URL := "http://yandex.ru"

	// Делаем POST запрос к серверу сокращения ссылок
	fmt.Println(cl.Post(cl.NetAddressServer, URL, client.TextData))

	// Выведет:
	// ссылка вида cl.NetAddressExpand/XXXXXXXX
	// где cl.NetAddressExpand адрес сервера извлечения оригинальных ссылок
	// и XXXXXXXX случайно сгенерированный уникальный код ссылки
}
