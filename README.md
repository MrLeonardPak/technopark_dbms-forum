# technopark_dbms-forum

[Репозиторий задания](https://github.com/mailcourses/technopark-dbms-forum)

[Описание API](https://app.swaggerhub.com/apis/MrLeonardPak/forum/0.1.0)

[Пост задания](https://park.vk.company/blog/topic/view/21180/)

## Сборка и запуск проекта

```
docker build --no-cache -t park .

docker run -d --memory 2G --log-opt max-size=5M --log-opt max-file=3 --name park_perf -p 5000:5000 park
```

## Сборка тестов

```
go get -u -v github.com/mailcourses/technopark-dbms-forum@master

go build github.com/mailcourses/technopark-dbms-forum
```

## Функциональное тестирование
```
go get -u -v github.com/mailcourses/technopark-dbms-forum@master

go build github.com/mailcourses/technopark-dbms-forum
./technopark-dbms-forum func -u http://localhost:5000/api -r report.html
```

Поддерживаются следующие параметры:
Параметр                              | Описание
---                                   | ---
-h, --help                            | Вывод списка поддерживаемых параметров
-u, --url[=http://localhost:5000/api] | Указание базовой URL тестируемого приложения
-k, --keep                            | Продолжить тестирование после первого упавшего теста
-t, --tests[=.*]                      | Маска запускаемых тестов (регулярное выражение)
-r, --report[=report.html]            | Имя файла для детального отчета о функциональном тестировании

## Нагрузочного тестирования
```
// заполнение:
./technopark-dbms-forum fill --url=http://localhost:5000/api --timeout=900

// тестирование:
./technopark-dbms-forum perf --url=http://localhost:5000/api --duration=600 --step=60
```
Параметры в примере означают:
- Лимит времени на заполнение базы - 15-ти минут;
- Нагрузка идёт 10 раз в течение 1-ой минуты. Учитывается лучший результат.