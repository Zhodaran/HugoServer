# Используем образ с Hugo
FROM klakegg/hugo:ext-alpine

# Устанавливаем рабочую директорию
WORKDIR /src

# Копируем все файлы проекта в контейнер
COPY . .

ENTRYPOINT [""]
# Команда для сборки сайта
RUN hugo --minify --gc

# Указываем, что контейнер будет слушать на порту 1313
EXPOSE 1313

# Запускаем сервер Hugo
CMD ["hugo", "server", "--bind=0.0.0.0", "--port=1313", "--watch"]
