## Для запуска приложения в консоли выполняем:
1. chmod +x run_docker.sh
2. ./run_docker.sh

## Проверка подключения к бд:
docker exec -it postgres_cont psql -U postgres -d chat_service

/dt - see all tables in psql
