#!/bin/sh

# prepare janus configuration
sed -i "s|@HTTP_PORT@|${HTTP_PORT:-8080}|g" ${JANUS_CONF}
sed -i "s|@LOG_LEVEL@|${LOG_LEVEL:-info}|g" ${JANUS_CONF}
sed -i "s|@ADMIN_HTTP_PORT@|${ADMIN_HTTP_PORT:-8081}|g" ${JANUS_CONF}
sed -i "s|@ADMIN_JWT_SECRET@|${ADMIN_JWT_SECRET}|g" ${JANUS_CONF}
sed -i "s|@ADMIN_BASIC_PASS@|${ADMIN_BASIC_PASS}|g" ${JANUS_CONF}

# clean the old api configs
rm -f ${API_CONF_DIR}/*

# prepare api new api configs
i=0
API_ENV=WB_API_${i}
eval API_NAME=\$${API_ENV}
while [ -n "${API_NAME}" ]; do
    ############### prepare env names ##############
    API_CONF="${API_CONF_DIR}/${API_NAME}.json"
    eval ENABLED=\$${API_ENV}_ENABLED
    eval PRESERVE_HOST=\$${API_ENV}_PRESERVE_HOST
    eval LISTEN_PATH=\$${API_ENV}_LISTEN_PATH
    eval UPSTREAM_TARGET=\$${API_ENV}_UPSTREAM_TARGET
    eval STRIP_PATH=\$${API_ENV}_STRIP_PATH
    eval APPEND_PATH=\$${API_ENV}_APPEND_PATH
    eval HTTP_METHODS=\$${API_ENV}_HTTP_METHODS

    eval RATE_LIMIT_ENABLED=\$${API_ENV}_RATE_LIMIT_ENABLED
    eval RATE_LIMIT_VALUE=\$${API_ENV}_RATE_LIMIT_VALUE

    eval WB_AUTH_ENABLED=\$${API_ENV}_WB_AUTH_ENABLED
    eval WB_AUTH_LOGIN_ENDPOINT=\$${API_ENV}_WB_AUTH_LOGIN_ENDPOINT
    eval WB_AUTH_CACHE_TTL_SECS=\$${API_ENV}_WB_AUTH_CACHE_TTL_SECS
    eval WB_AUTH_CACHE_CLEANUP_SECS=\$${API_ENV}_WB_AUTH_CACHE_CLEANUP_SECS

    ################# use envs ######################
    cp ${API_CONF_TEMPLATE} ${API_CONF}

    sed -i "s|@NAME@|${API_NAME}|g" ${API_CONF}
    sed -i "s|@ENABLED@|${ENABLED:-true}|g" ${API_CONF}
    sed -i "s|@PRESERVE_HOST@|${PRESERVE_HOST:-false}|g" ${API_CONF}
    sed -i "s|@LISTEN_PATH@|${LISTEN_PATH}|g" ${API_CONF}
    sed -i "s|@UPSTREAM_TARGET@|${UPSTREAM_TARGET}|g" ${API_CONF}
    sed -i "s|@STRIP_PATH@|${STRIP_PATH:-true}|g" ${API_CONF}
    sed -i "s|@APPEND_PATH@|${APPEND_PATH:-true}|g" ${API_CONF}
    sed -i "s|@HTTP_METHODS@|${HTTP_METHODS:-\"GET\",\"POST\",\"PUT\",\"DELETE\"}|g" ${API_CONF}

    sed -i "s|@RATE_LIMIT_ENABLED@|${RATE_LIMIT_ENABLED:-true}|g" ${API_CONF}
    sed -i "s|@RATE_LIMIT_VALUE@|${RATE_LIMIT_VALUE:-5-M}|g" ${API_CONF}

    sed -i "s|@WB_AUTH_ENABLED@|${WB_AUTH_ENABLED:-true}|g" ${API_CONF}
    sed -i "s|@WB_AUTH_LOGIN_ENDPOINT@|${WB_AUTH_LOGIN_ENDPOINT}|g" ${API_CONF}
    sed -i "s|@WB_AUTH_CACHE_TTL_SECS@|${WB_AUTH_CACHE_TTL_SECS:-30}|g" ${API_CONF}
    sed -i "s|@WB_AUTH_CACHE_CLEANUP_SECS@|${WB_AUTH_CACHE_CLEANUP_SECS:-60}|g" ${API_CONF}

    i=$((i+1))
    API_ENV=WB_API_${i}
    eval API_NAME=\$${API_ENV}
done

# start the service
${JANUS_BIN} -c ${JANUS_CONF} start
