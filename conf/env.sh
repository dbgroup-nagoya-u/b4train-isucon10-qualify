# server declaration
readonly WORKERS="127.0.0.1"
#readonly WORKERS="192.168.207.51 192.168.207.52 192.168.207.53" # use space for delimiter
readonly WEB_HOST="127.0.0.1"
readonly APP_HOSTS="127.0.0.1"
readonly DB_HOSTS="127.0.0.1"

# app declaration
readonly APP_LANG="go"
readonly SERVICE_NAME="isuumo"
readonly SERVICE_FILE="${SERVICE_NAME}.${APP_LANG}.service"

# app static files settings
readonly APP_DIR="${WORKSPACE}/webapp/${APP_LANG}"
readonly APP_OUT_LOG_PATH="${APP_DIR}/log/app.out"
readonly APP_ERR_LOG_PATH="${APP_DIR}/log/app.err"
readonly APP_SERVICE_PATH="/etc/systemd/system/${SERVICE_FILE}"

# Nginx static file settings
readonly NGINX_LOG_PATH="/var/log/nginx/access.log"
readonly NGINX_CONF_PATH="/etc/nginx/nginx.conf"
readonly NGINX_SITE_PATH="/etc/nginx/sites-available/${SERVICE_NAME}.conf"

# DB settings
readonly PG_HOST="127.0.0.1"
readonly PG_PORT="5432"
readonly PG_USER="isucon"
readonly PG_DBNAME="${SERVICE_NAME}"

# DB static file settings
readonly PG_LOG_PATH="/var/log/postgresql/postgresql-12-main.log"
readonly PG_OVERRIDE_PATH="/etc/postgresql/12/main/conf.d/override.conf"
readonly PG_HBA_PATH="/etc/postgresql/12/main/pg_hba.conf"
