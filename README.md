[![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/ineverbee/adverts-go-api.svg)](https://github.com/ineverbee/adverts-go-api) [![Go Report Card](https://goreportcard.com/badge/github.com/ineverbee/adverts-go-api)](https://goreportcard.com/report/github.com/ineverbee/adverts-go-api) <a href='https://github.com/jpoles1/gopherbadger' target='_blank'>![gopherbadger-tag-do-not-edit](https://img.shields.io/badge/Go%20Coverage-74%25-brightgreen.svg?longCache=true&style=flat)</a>

# adverts-go-api

## Тестовое задание от Avito Advertising

### Сервис
Сервис для хранения и подачи объявлений, хранящихся в базе данных postgres. Сервис предоставляет API, работающее поверх HTTP в формате JSON.

### Запуск
Сервис можно запустить командой:
```
docker-compose up
```

### Методы
**Метод получения списка объявлений**
* Пагинация: на одной странице 10 объявлений (параметр `page=[pageNumber]`);
* Cортировки: по цене (возрастание/убывание, параметры `sort=price&order=[asc/desc]`) и по дате создания (возрастание/убывание, параметры `sort=date&order=[asc/desc]`, по умолчанию сортировка по дате создания);
* Поля в ответе: название объявления, ссылка на главное фото (первое в списке), цена.

**Метод получения конкретного объявления**
* Поля в ответе: название объявления, цена, ссылка на главное фото;
* Опциональные поля (можно запросить, передав параметр с перечисленными через запятую полями `fields=[title|ad_description|price|photos]`): описание, ссылки на все фото.

**Метод создания объявления:**
* Принимает все вышеперечисленные поля: название, описание, несколько ссылок на фотографии, цена;
* Возвращает ID созданного объявления и код результата (ошибка или успех).

### API

#### /advert
* `POST` : Создать новое объявление

Запрос:
```
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{
        "title": "iPhone 12",
        "ad_description": "Base variant of iPhone 12",
        "price": 80000,
        "photos": [
            "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcTUriZwT-g07Lhj-UyVC3jxUgzdAP5pA1vVRbudohz23XI0JcNg2yhg48QsQLj__8_4LMM&usqp=CAU",
            "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcTctAyOkB4bRGA8nbAXBezBm-E7kCFiMqSEvQ&usqp=CAU"
        ]
    }' \
  http://localhost:8080/advert
```
Ответ: `id` - id нового объявления, `message` - сообщение об успешном создании объявления
```
 {
   "id":1,
   "message":"Success! Added new advert"
   }
```

#### /adverts?sort=[date/price]&order=[asc/desc]&page=[pageNumber]
* `GET` : Запросить список всех объявлений

Запрос:
```
curl --header "Content-Type: application/json" \
  --request GET \
  http://localhost:8080/adverts
```
Ответ: `title` - название объявления, `price` - цена, `photos` - массив с ссылкой на главное фото
```
 [{
    "title":"iPhone 12",
    "price":80000,
    "photos":["https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcTUriZwT-g07Lhj-UyVC3jxUgzdAP5pA1vVRbudohz23XI0JcNg2yhg48QsQLj__8_4LMM\u0026usqp=CAU"]
         }]
```

#### /adverts/{id}?fields=[title,ad_description,price,photos]
* `GET` : Запросить одно объявление по id

Запрос:
```
curl --header "Content-Type: application/json" \
  --request GET \
  http://localhost:8080/adverts/1
```
Ответ: `title` - название объявления, `price` - цена, `photos` - массив с ссылкой на главное фото
```
 {
    "title":"iPhone 12",
    "price":80000,
    "photos":["https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcTUriZwT-g07Lhj-UyVC3jxUgzdAP5pA1vVRbudohz23XI0JcNg2yhg48QsQLj__8_4LMM\u0026usqp=CAU"]
         }
```

### Усложнения
* Юнит тесты: покрытие в 70% и больше;
* Контейнеризация: поднятие проекта с помощью команды `docker-compose up`;
* Документация: структурированное описание методов сервиса.