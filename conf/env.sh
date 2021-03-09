# contest workspace (root of git repository)
WORKSPACE="/home/isucon/isuumo"

# server declaration
WORKERS="127.0.0.1"
#WORKERS="192.168.207.51 192.168.207.52 192.168.207.53" # use space for delimiter
WEB_HOST="127.0.0.1"
APP_HOSTS="127.0.0.1"
DB_HOSTS="127.0.0.1"

# app declaration
APP_LANG="go"
SERVICE_NAME="isuumo"
SERVICE_FILE="${SERVICE_NAME}.${APP_LANG}.service"

# app static files settings
APP_DIR="${WORKSPACE}/webapp/${APP_LANG}"
APP_OUT_LOG_PATH="${APP_DIR}/log/app.out"
APP_ERR_LOG_PATH="${APP_DIR}/log/app.err"
APP_SERVICE_PATH="/etc/systemd/system/${SERVICE_FILE}"

# Nginx static file settings
NGINX_LOG_PATH="/var/log/nginx/access.log"
NGINX_CONF_PATH="/etc/nginx/nginx.conf"
NGINX_SITE_PATH="/etc/nginx/sites-available/${SERVICE_NAME}.conf"

# DB settings
PG_HOST="127.0.0.1"
PG_PORT="5432"
PG_USER="isucon"
PG_DBNAME="isuumo"

# DB static file settings
PG_LOG_PATH="/var/log/postgresql/postgresql-12-main.log"
PG_OVERRIDE_PATH="/etc/postgresql/12/main/conf.d/override.conf"
PG_HBA_PATH="/etc/postgresql/12/main/pg_hba.conf"
