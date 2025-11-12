#Тербования 
Go 1.21 или выше
# Запуск 
для запуска сервера а так же устанновки зависимостей вам потрепуются две команды
``` 
go mod tidy - установка зависимостей 
go run  main.go storage.go handler.go link_checker.go pdf_generator.go
```
или же вы можете создать "production" версию для исполняемых файлов что пусть и не значительно но может облегчить запуск сервера 
```
go build -o "ваше желаемое название файла"
./link-checker - запуск через командную строку bash  
```
сервер будет доступен по адресу http://localhost:8080
# Тестирование
вариант команды для теста 
```
curl -X POST http://localhost:8080/check \
  -H "Content-Type: application/json" \
  -d '{"links": ["google.com","yandex.ru", "invalid-test-domain-12345.com"]}' 'сюда вы можете вставить любые свои ссылки как например "github.com"'
```
Генерация же отчёта проводится по команде 
```
curl -X POST http://localhost:8080/report \
  -H "Content-Type: application/json" \
  -d '{"links_list":[1]}' \
  --output report.pdf
```
# Основные использованные зависимости и библитеки 
```net/http
sync
encoding/json
context
github.com/jung-kurt/gofpdf - генерация PDF
```
сервер останавливается через ctrl + C что вызывает Graceful shutdown процесс с сохранением обрабатываемой информации 
в дальнейшем желатьельно перейти на jwt аутентификацию для большей безопасности а так же на https
