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
