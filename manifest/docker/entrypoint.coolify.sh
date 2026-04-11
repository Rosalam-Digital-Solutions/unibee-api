#!/bin/sh
set -eu

# Build CLI args from conventional env vars so we keep config.yaml defaults intact.
set -- ./main

if [ -n "${APP_ENV:-}" ]; then
    set -- "$@" "--env=${APP_ENV}"
elif [ -n "${ENV:-}" ]; then
    set -- "$@" "--env=${ENV}"
fi

if [ -n "${MODE:-}" ]; then
    set -- "$@" "--mode=${MODE}"
fi

if [ -n "${SERVER_ADDRESS:-}" ]; then
    set -- "$@" "--server-address=${SERVER_ADDRESS}"
fi

if [ -n "${UNIBEE_API_URL:-}" ]; then
    set -- "$@" "--unibee-api-url=${UNIBEE_API_URL}"
fi

if [ -n "${SERVER_JWT_KEY:-}" ]; then
    set -- "$@" "--server-jwtKey=${SERVER_JWT_KEY}"
fi

if [ -n "${SERVER_SWAGGER_PATH:-}" ]; then
    set -- "$@" "--server-swaggerPath=${SERVER_SWAGGER_PATH}"
fi

if [ -n "${DATABASE_LINK:-}" ]; then
    set -- "$@" "--database-link=${DATABASE_LINK}"
elif [ -n "${MYSQL_URL:-}" ]; then
    MYSQL_TMP="${MYSQL_URL#mysql://}"
    MYSQL_CREDS="${MYSQL_TMP%@*}"
    MYSQL_HOSTDB="${MYSQL_TMP#*@}"

    MYSQL_USER="${MYSQL_CREDS%%:*}"
    MYSQL_PASSWORD="${MYSQL_CREDS#*:}"

    MYSQL_HOSTPORT="${MYSQL_HOSTDB%%/*}"
    MYSQL_DB="${MYSQL_HOSTDB#*/}"

    MYSQL_HOST="${MYSQL_HOSTPORT%%:*}"
    MYSQL_PORT="${MYSQL_HOSTPORT#*:}"
    if [ "${MYSQL_PORT}" = "${MYSQL_HOSTPORT}" ]; then
        MYSQL_PORT="3306"
    fi

    DB_QUERY="${DB_QUERY:-loc=UTC&parseTime=false}"
    DB_LINK="mysql:${MYSQL_USER}:${MYSQL_PASSWORD}@tcp(${MYSQL_HOST}:${MYSQL_PORT})/${MYSQL_DB}"
    if [ -n "${DB_QUERY}" ]; then
        DB_LINK="${DB_LINK}?${DB_QUERY}"
    fi

    set -- "$@" "--database-link=${DB_LINK}"
elif [ -n "${DB_HOST:-}" ]; then
    DB_PORT="${DB_PORT:-3306}"
    DB_NAME="${DB_DATABASE:-unib}"
    DB_USER="${DB_USER:-unibee}"
    DB_PASSWORD="${DB_PASSWORD:-}"
    DB_QUERY="${DB_QUERY:-}"

    DB_LINK="mysql:${DB_USER}:${DB_PASSWORD}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}"
    if [ -n "${DB_QUERY}" ]; then
        DB_LINK="${DB_LINK}?${DB_QUERY}"
    fi

    set -- "$@" "--database-link=${DB_LINK}"
fi

if [ -n "${REDIS_ADDRESS:-}" ]; then
    set -- "$@" "--redis-address=${REDIS_ADDRESS}"
elif [ -n "${REDIS_URL:-}" ]; then
    REDIS_TMP="${REDIS_URL#redis://}"
    REDIS_CREDS="${REDIS_TMP%@*}"
    REDIS_HOSTDB="${REDIS_TMP#*@}"

    REDIS_USER="${REDIS_CREDS%%:*}"
    REDIS_PASS_FROM_URL="${REDIS_CREDS#*:}"

    REDIS_HOSTPORT="${REDIS_HOSTDB%%/*}"
    REDIS_DB_FROM_URL="${REDIS_HOSTDB#*/}"

    REDIS_HOST_FROM_URL="${REDIS_HOSTPORT%%:*}"
    REDIS_PORT_FROM_URL="${REDIS_HOSTPORT#*:}"
    if [ "${REDIS_PORT_FROM_URL}" = "${REDIS_HOSTPORT}" ]; then
        REDIS_PORT_FROM_URL="6379"
    fi

    set -- "$@" "--redis-address=${REDIS_HOST_FROM_URL}:${REDIS_PORT_FROM_URL}"
    if [ -n "${REDIS_PASS_FROM_URL}" ] && [ "${REDIS_PASS_FROM_URL}" != "${REDIS_USER}" ]; then
        set -- "$@" "--redis-password=${REDIS_PASS_FROM_URL}"
    fi
    if [ -n "${REDIS_DB_FROM_URL}" ] && [ "${REDIS_DB_FROM_URL}" != "${REDIS_HOSTDB}" ]; then
        set -- "$@" "--redis-database=${REDIS_DB_FROM_URL}"
    fi
elif [ -n "${REDIS_HOST:-}" ]; then
    REDIS_PORT="${REDIS_PORT:-6379}"
    set -- "$@" "--redis-address=${REDIS_HOST}:${REDIS_PORT}"
fi

if [ -n "${REDIS_PASSWORD:-}" ]; then
    set -- "$@" "--redis-password=${REDIS_PASSWORD}"
fi

if [ -n "${REDIS_DATABASE:-}" ]; then
    set -- "$@" "--redis-database=${REDIS_DATABASE}"
fi

if [ -n "${NACOS_IP:-}" ]; then
    set -- "$@" "--nacos-ip=${NACOS_IP}"
fi
if [ -n "${NACOS_PORT:-}" ]; then
    set -- "$@" "--nacos-port=${NACOS_PORT}"
fi
if [ -n "${NACOS_NAMESPACE:-}" ]; then
    set -- "$@" "--nacos-namespace=${NACOS_NAMESPACE}"
fi
if [ -n "${NACOS_GROUP:-}" ]; then
    set -- "$@" "--nacos-group=${NACOS_GROUP}"
fi
if [ -n "${NACOS_DATA_ID:-}" ]; then
    set -- "$@" "--nacos-data-id=${NACOS_DATA_ID}"
fi

exec "$@"
