#!/bin/bash

IMAGE_NAME="landing"
CONTAINER_NAME="landing"

function build_and_start_container {
  # Сборка докер-контейнера
  docker build -t $IMAGE_NAME .

  # Остановка и удаление контейнера, если он уже существует
  docker stop $CONTAINER_NAME 2>/dev/null
  docker rm $CONTAINER_NAME 2>/dev/null

  # Запуск контейнера с автоматическим перезапуском после каждого краша на порту 8443
  docker run -d --restart always --name $CONTAINER_NAME -p 8443:8443 $IMAGE_NAME
  echo "Container started successfully!"
}

function stop_and_remove_container {
  # Остановка и удаление контейнера
  docker stop $CONTAINER_NAME
  docker rm $CONTAINER_NAME
  echo "Container stopped and removed successfully!"
}

# Проверка команды, переданной в скрипт
if [ "$1" == "start" ]; then
  build_and_start_container
elif [ "$1" == "stop" ]; then
  stop_and_remove_container
else
  echo "Usage: $0 [start|stop]"
fi
