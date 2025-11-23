CREATE DATABASE IF NOT EXISTS auth_service_dev;

CREATE DATABASE IF NOT EXISTS auth_service_db;

GRANT ALL PRIVILEGES ON *.* TO 'auth_service_user'@'%';
