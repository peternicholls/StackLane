#!/usr/bin/env bash

set -euo pipefail

twentyi_script_dir() {
    local source_path="${BASH_SOURCE[0]}"
    while [[ -h "$source_path" ]]; do
        local source_dir
        source_dir="$(cd "$(dirname "$source_path")" && pwd -P)"
        source_path="$(readlink "$source_path")"
        [[ "$source_path" != /* ]] && source_path="$source_dir/$source_path"
    done

    cd "$(dirname "$source_path")/.." && pwd -P
}

twentyi_trim() {
    local value="$1"
    value="${value#${value%%[![:space:]]*}}"
    value="${value%${value##*[![:space:]]}}"
    printf '%s' "$value"
}

twentyi_load_env_file() {
    local env_file="$1"
    local mode="$2"

    [[ -f "$env_file" ]] || return 0

    while IFS= read -r raw_line || [[ -n "$raw_line" ]]; do
        local line key value

        line="$(twentyi_trim "$raw_line")"
        [[ -z "$line" || "${line#\#}" != "$line" ]] && continue

        line="${line#export }"
        [[ "$line" == *=* ]] || continue

        key="$(twentyi_trim "${line%%=*}")"
        value="$(twentyi_trim "${line#*=}")"
        [[ "$key" =~ ^[A-Za-z_][A-Za-z0-9_]*$ ]] || continue

        if [[ "$value" == \"*\" && "$value" == *\" ]]; then
            value="${value#\"}"
            value="${value%\"}"
        elif [[ "$value" == \'*\' ]]; then
            value="${value#\'}"
            value="${value%\'}"
        fi

        if [[ "$mode" == "preserve" && -n "${!key+x}" ]]; then
            continue
        fi

        printf -v "$key" '%s' "$value"
        export "$key"
    done < "$env_file"
}

twentyi_default_stack_home() {
    local repo_home
    repo_home="$(twentyi_script_dir)"

    if [[ -n "${STACK_HOME:-}" ]]; then
        printf '%s' "$STACK_HOME"
    elif [[ -f "$repo_home/docker-compose.yml" ]]; then
        printf '%s' "$repo_home"
    else
        printf '%s' "$HOME/docker/20i-stack"
    fi
}

twentyi_abs_dir() {
    local path="$1"

    if [[ "$path" == /* ]]; then
        cd "$path" && pwd -P
    else
        cd "$PWD/$path" && pwd -P
    fi
}

twentyi_abs_path_from_base() {
    local base_dir="$1"
    local path="$2"

    if [[ "$path" == /* ]]; then
        printf '%s' "$path"
    else
        printf '%s/%s' "$base_dir" "$path"
    fi
}

twentyi_slugify() {
    local input="$1"
    local value

    value="$(printf '%s' "$input" | tr '[:upper:]' '[:lower:]')"
    value="$(printf '%s' "$value" | sed -E 's/[^a-z0-9]+/-/g; s/^-+//; s/-+$//; s/-{2,}/-/g')"
    value="${value:0:63}"
    value="${value%-}"

    if [[ -z "$value" ]]; then
        value="site"
    fi

    printf '%s' "$value"
}

twentyi_project_state_file() {
    printf '%s/projects/%s.env' "$TWENTYI_STATE_DIR" "$PROJECT_SLUG"
}

twentyi_shared_env_file() {
    printf '%s/shared/gateway.env' "$TWENTYI_STATE_DIR"
}

twentyi_shared_gateway_config_file() {
    printf '%s/shared/gateway.conf' "$TWENTYI_STATE_DIR"
}

twentyi_load_state_file() {
    local state_file="$1"

    [[ -f "$state_file" ]] || return 1
    # shellcheck disable=SC1090
    source "$state_file"
}

twentyi_state_files() {
    local found=0
    local state_file

    for state_file in "$TWENTYI_STATE_DIR"/projects/*.env; do
        [[ -e "$state_file" ]] || continue
        found=1
        printf '%s\n' "$state_file"
    done

    return $((found == 0))
}

twentyi_count_state_files() {
    local count=0
    local state_file

    for state_file in "$TWENTYI_STATE_DIR"/projects/*.env; do
        [[ -e "$state_file" ]] || continue
        count=$((count + 1))
    done

    printf '%s' "$count"
}

twentyi_port_in_use() {
    local port="$1"

    if command -v lsof >/dev/null 2>&1; then
        lsof -nP -iTCP:"$port" -sTCP:LISTEN >/dev/null 2>&1
    else
        return 1
    fi
}

twentyi_port_reserved() {
    local var_name="$1"
    local port="$2"
    local state_file current_port

    for state_file in "$TWENTYI_STATE_DIR"/projects/*.env; do
        [[ -e "$state_file" ]] || continue
        unset HOST_PORT MYSQL_PORT PMA_PORT PROJECT_DIR HOSTNAME ATTACHMENT_STATE COMPOSE_PROJECT_NAME
        twentyi_load_state_file "$state_file"
        current_port="${!var_name:-}"
        [[ -n "$current_port" && "$current_port" == "$port" ]] && return 0
    done

    return 1
}

twentyi_find_available_port() {
    local var_name="$1"
    local start_port="$2"
    local port="$start_port"

    while [[ "$port" -lt 65535 ]]; do
        if ! twentyi_port_in_use "$port" && ! twentyi_port_reserved "$var_name" "$port"; then
            printf '%s' "$port"
            return 0
        fi
        port=$((port + 1))
    done

    return 1
}

twentyi_resolve_docroot() {
    if [[ -n "${DOCROOT:-}" ]]; then
        DOCROOT="$(twentyi_abs_path_from_base "$PROJECT_DIR" "$DOCROOT")"
    elif [[ -n "${CODE_DIR:-}" ]]; then
        DOCROOT="$(twentyi_abs_path_from_base "$PROJECT_DIR" "$CODE_DIR")"
    elif [[ -d "$PROJECT_DIR/public_html" ]]; then
        DOCROOT="$PROJECT_DIR/public_html"
    else
        DOCROOT="$PROJECT_DIR"
    fi

    if [[ ! -d "$DOCROOT" ]]; then
        printf 'Error: document root not found: %s\n' "$DOCROOT" >&2
        exit 1
    fi

    DOCROOT="$(cd "$DOCROOT" && pwd -P)"

    if [[ "$DOCROOT" == "$PROJECT_DIR" ]]; then
        DOCROOT_RELATIVE=""
    elif [[ "$DOCROOT" == "$PROJECT_DIR/"* ]]; then
        DOCROOT_RELATIVE="${DOCROOT#$PROJECT_DIR/}"
    else
        printf 'Error: document root must live inside the project directory for the current 20i-style container layout: %s\n' "$DOCROOT" >&2
        exit 1
    fi
}

twentyi_resolve_hostname() {
    SITE_SUFFIX="${SITE_SUFFIX:-test}"
    SITE_SUFFIX="$(twentyi_slugify "$SITE_SUFFIX")"

    if [[ -n "${SITE_HOSTNAME:-}" ]]; then
        HOSTNAME="$SITE_HOSTNAME"
    else
        HOSTNAME="$PROJECT_SLUG.$SITE_SUFFIX"
    fi
}

twentyi_resolve_ports() {
    local project_count
    project_count="$(twentyi_count_state_files)"

    if [[ -z "${HOST_PORT:-}" ]]; then
        if [[ "$TWENTYI_COMMAND" == "up" && "$project_count" -eq 0 ]] && ! twentyi_port_in_use 80; then
            HOST_PORT=80
        else
            HOST_PORT="$(twentyi_find_available_port HOST_PORT 8080)"
        fi
    fi

    if [[ -z "${MYSQL_PORT:-}" ]]; then
        if [[ "$project_count" -eq 0 ]] && ! twentyi_port_in_use 3306; then
            MYSQL_PORT=3306
        else
            MYSQL_PORT="$(twentyi_find_available_port MYSQL_PORT 3307)"
        fi
    fi

    if [[ -z "${PMA_PORT:-}" ]]; then
        if [[ "$project_count" -eq 0 ]] && ! twentyi_port_in_use 8081; then
            PMA_PORT=8081
        else
            PMA_PORT="$(twentyi_find_available_port PMA_PORT 8082)"
        fi
    fi
}

twentyi_require_docker() {
    if ! command -v docker >/dev/null 2>&1; then
        printf 'Error: docker is required for %s\n' "$TWENTYI_COMMAND" >&2
        exit 1
    fi
}

twentyi_validate_requested_ports() {
    local port_var current_value state_file existing_project_dir
    local current_state_file existing_port allow_current_project_port
    local current_project_dir
    local resolved_host_port resolved_mysql_port resolved_pma_port
    local current_project_name current_project_slug current_compose_project_name current_hostname current_web_network_alias current_site_suffix current_docroot current_docroot_relative current_container_site_root current_container_docroot current_mysql_database current_mysql_user current_mysql_password current_mysql_version current_mysql_root_password current_php_version

    current_state_file="$(twentyi_project_state_file)"
    current_project_dir="$PROJECT_DIR"
    current_project_name="$PROJECT_NAME"
    current_project_slug="$PROJECT_SLUG"
    current_compose_project_name="$COMPOSE_PROJECT_NAME"
    current_hostname="$HOSTNAME"
    current_web_network_alias="$WEB_NETWORK_ALIAS"
    current_site_suffix="$SITE_SUFFIX"
    current_docroot="$DOCROOT"
    current_docroot_relative="$DOCROOT_RELATIVE"
    current_container_site_root="$CONTAINER_SITE_ROOT"
    current_container_docroot="$CONTAINER_DOCROOT"
    current_mysql_database="$MYSQL_DATABASE"
    current_mysql_user="$MYSQL_USER"
    current_mysql_password="$MYSQL_PASSWORD"
    current_mysql_version="$MYSQL_VERSION"
    current_mysql_root_password="$MYSQL_ROOT_PASSWORD"
    current_php_version="$PHP_VERSION"
    resolved_host_port="${HOST_PORT:-}"
    resolved_mysql_port="${MYSQL_PORT:-}"
    resolved_pma_port="${PMA_PORT:-}"

    for port_var in HOST_PORT MYSQL_PORT PMA_PORT; do
        current_value="${!port_var:-}"
        [[ -n "$current_value" ]] || continue

        allow_current_project_port=0

        if [[ -f "$current_state_file" ]]; then
            unset PROJECT_DIR HOST_PORT MYSQL_PORT PMA_PORT
            twentyi_load_state_file "$current_state_file"
            existing_port="${!port_var:-}"
            if [[ "${PROJECT_DIR:-}" == "$current_project_dir" && "$existing_port" == "$current_value" ]]; then
                allow_current_project_port=1
            fi
        fi

        if [[ "$allow_current_project_port" -eq 0 ]] && twentyi_port_in_use "$current_value"; then
            printf 'Error: %s is already listening on port %s\n' "$port_var" "$current_value" >&2
            exit 1
        fi

        for state_file in "$TWENTYI_STATE_DIR"/projects/*.env; do
            [[ -e "$state_file" ]] || continue
            unset PROJECT_DIR HOST_PORT MYSQL_PORT PMA_PORT
            twentyi_load_state_file "$state_file"
            existing_project_dir="${PROJECT_DIR:-}"

            if [[ "$existing_project_dir" != "$PROJECT_DIR" && "${!port_var:-}" == "$current_value" ]]; then
                printf 'Error: %s port %s is already reserved by %s\n' "$port_var" "$current_value" "$existing_project_dir" >&2
                exit 1
            fi
        done

        PROJECT_DIR="$current_project_dir"
        PROJECT_NAME="$current_project_name"
        PROJECT_SLUG="$current_project_slug"
        COMPOSE_PROJECT_NAME="$current_compose_project_name"
        HOSTNAME="$current_hostname"
        WEB_NETWORK_ALIAS="$current_web_network_alias"
        SITE_SUFFIX="$current_site_suffix"
        DOCROOT="$current_docroot"
        DOCROOT_RELATIVE="$current_docroot_relative"
        CONTAINER_SITE_ROOT="$current_container_site_root"
        CONTAINER_DOCROOT="$current_container_docroot"
        MYSQL_DATABASE="$current_mysql_database"
        MYSQL_USER="$current_mysql_user"
        MYSQL_PASSWORD="$current_mysql_password"
        MYSQL_VERSION="$current_mysql_version"
        MYSQL_ROOT_PASSWORD="$current_mysql_root_password"
        PHP_VERSION="$current_php_version"
        HOST_PORT="$resolved_host_port"
        MYSQL_PORT="$resolved_mysql_port"
        PMA_PORT="$resolved_pma_port"
    done
}

twentyi_validate_collision() {
    local state_file
    local existing_project_dir existing_hostname
    local current_project_name current_project_slug current_compose_project_name current_hostname current_web_network_alias current_site_suffix current_docroot current_project_dir current_docroot_relative current_container_site_root current_container_docroot current_mysql_database current_mysql_user current_mysql_password current_mysql_version current_mysql_root_password current_php_version current_mysql_port current_pma_port

    current_project_name="$PROJECT_NAME"
    current_project_slug="$PROJECT_SLUG"
    current_compose_project_name="$COMPOSE_PROJECT_NAME"
    current_hostname="$HOSTNAME"
    current_web_network_alias="$WEB_NETWORK_ALIAS"
    current_site_suffix="$SITE_SUFFIX"
    current_docroot="$DOCROOT"
    current_project_dir="$PROJECT_DIR"
    current_docroot_relative="$DOCROOT_RELATIVE"
    current_container_site_root="$CONTAINER_SITE_ROOT"
    current_container_docroot="$CONTAINER_DOCROOT"
    current_mysql_database="$MYSQL_DATABASE"
    current_mysql_user="$MYSQL_USER"
    current_mysql_password="$MYSQL_PASSWORD"
    current_mysql_version="$MYSQL_VERSION"
    current_mysql_root_password="$MYSQL_ROOT_PASSWORD"
    current_php_version="$PHP_VERSION"
    current_mysql_port="$MYSQL_PORT"
    current_pma_port="$PMA_PORT"

    for state_file in "$TWENTYI_STATE_DIR"/projects/*.env; do
        [[ -e "$state_file" ]] || continue
        unset PROJECT_DIR HOSTNAME ATTACHMENT_STATE PROJECT_NAME PROJECT_SLUG COMPOSE_PROJECT_NAME HOST_PORT MYSQL_PORT PMA_PORT
        twentyi_load_state_file "$state_file"
        existing_project_dir="${PROJECT_DIR:-}"
        existing_hostname="${HOSTNAME:-}"

        if [[ "$state_file" == "$(twentyi_project_state_file)" && "$existing_project_dir" != "$PROJECT_DIR" ]]; then
            printf 'Error: project slug collision for %s (already registered by %s)\n' "$PROJECT_SLUG" "$existing_project_dir" >&2
            exit 1
        fi

        if [[ "$state_file" != "$(twentyi_project_state_file)" && "$existing_hostname" == "$HOSTNAME" ]]; then
            printf 'Error: hostname collision for %s (%s already registered by %s)\n' "$HOSTNAME" "$existing_hostname" "$existing_project_dir" >&2
            exit 1
        fi

        PROJECT_NAME="$current_project_name"
        PROJECT_SLUG="$current_project_slug"
        COMPOSE_PROJECT_NAME="$current_compose_project_name"
        HOSTNAME="$current_hostname"
        WEB_NETWORK_ALIAS="$current_web_network_alias"
        SITE_SUFFIX="$current_site_suffix"
        DOCROOT="$current_docroot"
        DOCROOT_RELATIVE="$current_docroot_relative"
        CONTAINER_SITE_ROOT="$current_container_site_root"
        CONTAINER_DOCROOT="$current_container_docroot"
        MYSQL_DATABASE="$current_mysql_database"
        MYSQL_USER="$current_mysql_user"
        MYSQL_PASSWORD="$current_mysql_password"
        MYSQL_VERSION="$current_mysql_version"
        MYSQL_ROOT_PASSWORD="$current_mysql_root_password"
        PHP_VERSION="$current_php_version"
        MYSQL_PORT="$current_mysql_port"
        PMA_PORT="$current_pma_port"
        PROJECT_DIR="$current_project_dir"
    done
}

twentyi_export_runtime_env() {
    export CODE_DIR="$DOCROOT"
    export PROJECT_ROOT="$PROJECT_DIR"
    export CONTAINER_SITE_ROOT
    export CONTAINER_DOCROOT
    export COMPOSE_PROJECT_NAME
    export MYSQL_PORT
    export PMA_PORT
    export MYSQL_VERSION
    export MYSQL_ROOT_PASSWORD
    export MYSQL_DATABASE
    export MYSQL_USER
    export MYSQL_PASSWORD
    export PHP_VERSION
    export WEB_NETWORK_ALIAS
    export SHARED_GATEWAY_NETWORK
}

twentyi_export_shared_env() {
    export SHARED_GATEWAY_NETWORK
    export SHARED_GATEWAY_HTTP_PORT
    export SHARED_GATEWAY_HTTPS_PORT
    export SHARED_GATEWAY_COMPOSE_PROJECT_NAME
    export SHARED_GATEWAY_CONFIG_FILE
}

twentyi_print_runtime_summary() {
    cat <<EOF
Project name:      $PROJECT_NAME
Project slug:      $PROJECT_SLUG
Project dir:       $PROJECT_DIR
Document root:     $DOCROOT
Container root:    $CONTAINER_SITE_ROOT
Container docroot: $CONTAINER_DOCROOT
Compose project:   $COMPOSE_PROJECT_NAME
Planned hostname:  $HOSTNAME
Gateway alias:     $WEB_NETWORK_ALIAS
Current access:    http://localhost:$SHARED_GATEWAY_HTTP_PORT
Database port:     $MYSQL_PORT
phpMyAdmin port:   $PMA_PORT
PHP version:       $PHP_VERSION
MySQL database:    $MYSQL_DATABASE
State dir:         $TWENTYI_STATE_DIR
EOF
}

twentyi_write_state() {
    local state_file
    state_file="$(twentyi_project_state_file)"

    mkdir -p "$TWENTYI_STATE_DIR/projects"
    : > "$state_file"

    {
        printf 'PROJECT_NAME=%q\n' "$PROJECT_NAME"
        printf 'PROJECT_SLUG=%q\n' "$PROJECT_SLUG"
        printf 'PROJECT_DIR=%q\n' "$PROJECT_DIR"
        printf 'DOCROOT=%q\n' "$DOCROOT"
        printf 'DOCROOT_RELATIVE=%q\n' "$DOCROOT_RELATIVE"
        printf 'HOSTNAME=%q\n' "$HOSTNAME"
        printf 'SITE_SUFFIX=%q\n' "$SITE_SUFFIX"
        printf 'COMPOSE_PROJECT_NAME=%q\n' "$COMPOSE_PROJECT_NAME"
        printf 'MYSQL_PORT=%q\n' "$MYSQL_PORT"
        printf 'PMA_PORT=%q\n' "$PMA_PORT"
        printf 'WEB_NETWORK_ALIAS=%q\n' "$WEB_NETWORK_ALIAS"
        printf 'CONTAINER_SITE_ROOT=%q\n' "$CONTAINER_SITE_ROOT"
        printf 'CONTAINER_DOCROOT=%q\n' "$CONTAINER_DOCROOT"
        printf 'MYSQL_DATABASE=%q\n' "$MYSQL_DATABASE"
        printf 'MYSQL_USER=%q\n' "$MYSQL_USER"
        printf 'MYSQL_PASSWORD=%q\n' "$MYSQL_PASSWORD"
        printf 'MYSQL_VERSION=%q\n' "$MYSQL_VERSION"
        printf 'MYSQL_ROOT_PASSWORD=%q\n' "$MYSQL_ROOT_PASSWORD"
        printf 'PHP_VERSION=%q\n' "$PHP_VERSION"
        printf 'ATTACHMENT_STATE=%q\n' "$ATTACHMENT_STATE"
    } >> "$state_file"
}

twentyi_remove_state() {
    local state_file
    state_file="$(twentyi_project_state_file)"
    [[ -f "$state_file" ]] && rm -f "$state_file"
}

twentyi_docker_status() {
    local compose_project="$1"
    local status_lines

    if ! command -v docker >/dev/null 2>&1; then
        printf 'docker-unavailable'
        return 0
    fi

    status_lines="$(docker ps --filter "label=com.docker.compose.project=$compose_project" --format '{{.Names}} ({{.Status}})' 2>/dev/null || true)"
    if [[ -n "$status_lines" ]]; then
        printf '%s' "$status_lines"
    else
        printf 'stopped'
    fi
}

twentyi_shared_gateway_status() {
    local status_lines

    if ! command -v docker >/dev/null 2>&1; then
        printf 'docker-unavailable'
        return 0
    fi

    status_lines="$(docker ps --filter "label=com.docker.compose.project=$SHARED_GATEWAY_COMPOSE_PROJECT_NAME" --format '{{.Names}} ({{.Status}})' 2>/dev/null || true)"
    if [[ -n "$status_lines" ]]; then
        printf '%s' "$status_lines"
    else
        printf 'stopped'
    fi
}

twentyi_compose() {
    docker compose -f "$TWENTYI_STACK_FILE" -p "$COMPOSE_PROJECT_NAME" "$@"
}

twentyi_shared_compose() {
    docker compose --env-file "$(twentyi_shared_env_file)" -f "$TWENTYI_SHARED_STACK_FILE" -p "$SHARED_GATEWAY_COMPOSE_PROJECT_NAME" "$@"
}

twentyi_wait_for_gateway_ready() {
    local attempt

    for attempt in $(seq 1 25); do
        if twentyi_shared_compose exec -T gateway sh -c 'wget -qO- http://127.0.0.1/__20i_gateway_health >/dev/null 2>&1' >/dev/null 2>&1; then
            return 0
        fi
        sleep 0.2
    done

    return 1
}

twentyi_wait_for_route_target() {
    local route_target="$1"
    local attempt

    if [[ "$route_target" == "twentyi-no-route" ]]; then
        return 0
    fi

    for attempt in $(seq 1 25); do
        if twentyi_shared_compose exec -T gateway sh -c "wget -qO- http://$route_target >/dev/null 2>&1" >/dev/null 2>&1; then
            return 0
        fi
        sleep 0.2
    done

    return 1
}

twentyi_wait_for_gateway_route() {
    local route_target="$1"
    local attempt

    if [[ "$route_target" == "twentyi-no-route" ]]; then
        return 0
    fi

    for attempt in $(seq 1 25); do
        if twentyi_shared_compose exec -T gateway sh -c 'wget -qO- http://127.0.0.1/ >/dev/null 2>&1' >/dev/null 2>&1; then
            return 0
        fi
        sleep 0.2
    done

    return 1
}

twentyi_write_gateway_config() {
    local route_target="$1"
    local config_file

    config_file="$(twentyi_shared_gateway_config_file)"
    mkdir -p "$(dirname "$config_file")"

    if [[ "$route_target" == "twentyi-no-route" ]]; then
        cat > "$config_file" <<EOF
server {
    listen 80 default_server;
    listen 443 default_server;
    server_name _;

    access_log /var/log/nginx/access.log;
    error_log /var/log/nginx/error.log warn;

    add_header X-20i-Gateway "shared" always;
    add_header X-20i-Route-Target "twentyi-no-route" always;

    location = /__20i_gateway_health {
        default_type text/plain;
        return 200 "gateway ok\\n";
    }

    location / {
        default_type text/plain;
        return 503 "20i shared gateway has no attached project route.\\n";
    }
}
EOF
        return 0
    fi

    cat > "$config_file" <<EOF
server {
    listen 80 default_server;
    listen 443 default_server;
    server_name _;

    access_log /var/log/nginx/access.log;
    error_log /var/log/nginx/error.log warn;

    add_header X-20i-Gateway "shared" always;
    add_header X-20i-Route-Target "$route_target" always;

    location = /__20i_gateway_health {
        default_type text/plain;
        return 200 "gateway ok\\n";
    }

    location / {
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Forwarded-Host \$host;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_connect_timeout 2s;
        proxy_read_timeout 600s;
        proxy_pass http://$route_target:80;

        error_page 502 503 504 = @route_unavailable;
    }

    location @route_unavailable {
        default_type text/plain;
        return 503 "20i shared gateway could not reach '$route_target'.\\n";
    }
}
EOF
}

twentyi_write_shared_env() {
    local shared_env_file

    shared_env_file="$(twentyi_shared_env_file)"
    mkdir -p "$(dirname "$shared_env_file")"
    : > "$shared_env_file"

    {
        printf 'SHARED_GATEWAY_NETWORK=%q\n' "$SHARED_GATEWAY_NETWORK"
        printf 'SHARED_GATEWAY_HTTP_PORT=%q\n' "$SHARED_GATEWAY_HTTP_PORT"
        printf 'SHARED_GATEWAY_HTTPS_PORT=%q\n' "$SHARED_GATEWAY_HTTPS_PORT"
        printf 'SHARED_GATEWAY_CONFIG_FILE=%q\n' "$SHARED_GATEWAY_CONFIG_FILE"
    } >> "$shared_env_file"
}

twentyi_pick_gateway_target() {
    local preferred_slug="${1:-}"
    local state_file

    if [[ -n "$preferred_slug" ]]; then
        state_file="$TWENTYI_STATE_DIR/projects/$preferred_slug.env"
        if [[ -f "$state_file" ]]; then
            unset PROJECT_NAME PROJECT_SLUG PROJECT_DIR DOCROOT HOSTNAME SITE_SUFFIX COMPOSE_PROJECT_NAME MYSQL_PORT PMA_PORT MYSQL_DATABASE MYSQL_USER MYSQL_PASSWORD MYSQL_VERSION MYSQL_ROOT_PASSWORD PHP_VERSION ATTACHMENT_STATE WEB_NETWORK_ALIAS
            twentyi_load_state_file "$state_file"
            if [[ "${ATTACHMENT_STATE:-}" == "attached" && -n "${WEB_NETWORK_ALIAS:-}" ]]; then
                printf '%s|%s' "$WEB_NETWORK_ALIAS" "$HOSTNAME"
                return 0
            fi
        fi
    fi

    for state_file in "$TWENTYI_STATE_DIR"/projects/*.env; do
        [[ -e "$state_file" ]] || continue
        unset PROJECT_NAME PROJECT_SLUG PROJECT_DIR DOCROOT HOSTNAME SITE_SUFFIX COMPOSE_PROJECT_NAME MYSQL_PORT PMA_PORT MYSQL_DATABASE MYSQL_USER MYSQL_PASSWORD MYSQL_VERSION MYSQL_ROOT_PASSWORD PHP_VERSION ATTACHMENT_STATE WEB_NETWORK_ALIAS
        twentyi_load_state_file "$state_file"
        if [[ "${ATTACHMENT_STATE:-}" == "attached" && -n "${WEB_NETWORK_ALIAS:-}" ]]; then
            printf '%s|%s' "$WEB_NETWORK_ALIAS" "$HOSTNAME"
            return 0
        fi
    done

    printf 'twentyi-no-route|localhost'
}

twentyi_update_gateway_route() {
    local preferred_slug="${1:-}"
    local gateway_target gateway_hostname route_target route_hostname

    gateway_target="$(twentyi_pick_gateway_target "$preferred_slug")"
    route_target="${gateway_target%%|*}"
    route_hostname="${gateway_target#*|}"

    twentyi_write_gateway_config "$route_target"
    twentyi_write_shared_env
    twentyi_export_shared_env

    if docker ps --filter "label=com.docker.compose.project=$SHARED_GATEWAY_COMPOSE_PROJECT_NAME" --filter "label=com.docker.compose.service=gateway" --format '{{.Names}}' | grep -q .; then
        twentyi_wait_for_route_target "$route_target"
        twentyi_shared_compose exec -T gateway nginx -s reload >/dev/null
    else
        twentyi_shared_compose up -d
    fi

    twentyi_wait_for_gateway_ready
    twentyi_wait_for_gateway_route "$route_target"
}

twentyi_ensure_shared_infra() {
    twentyi_require_docker

    if ! docker network inspect "$SHARED_GATEWAY_NETWORK" >/dev/null 2>&1; then
        docker network create "$SHARED_GATEWAY_NETWORK" >/dev/null
    fi

    twentyi_write_gateway_config "twentyi-no-route"
    twentyi_write_shared_env
    twentyi_export_shared_env
    twentyi_shared_compose up -d
    twentyi_wait_for_gateway_ready
}

twentyi_note_phase_status() {
    printf 'Routing mode: shared gateway on host ports with a single default route (hostname-aware routing and .test DNS land in later phases)\n'
}

twentyi_usage() {
    cat <<EOF
Usage: $(basename "$0") [options]

Common options:
  --project-dir PATH       Use a project directory other than the current directory
  --site-name NAME         Override the hostname/project basename source
  --site-hostname HOST     Override the full planned hostname
  --site-suffix SUFFIX     Override the suffix used for planned hostnames (default: test)
  --docroot PATH           Override the document root (default: public_html when present)
  --php-version VERSION    Override the PHP image version
  --mysql-database NAME    Override the project database name
  --mysql-user USER        Override the project database user
  --mysql-password PASS    Override the project database password
  --mysql-port PORT        Override the database port used in Phase 1
  --pma-port PORT          Override the phpMyAdmin port used in Phase 1
  --dry-run                Resolve config and print the docker command without executing it
  --help                   Show this help

Compatibility aliases:
  version=8.4              Same as --php-version 8.4

State model:
    up / attach              Ensure the shared gateway exists, mark the project as attached, and start its runtime
  down                     Stop the project runtime and retain a down state record
  detach                   Stop the project runtime and remove its attachment record
  down --all               Stop all known runtimes and clear all attachment state
EOF
}

twentyi_parse_initial_args() {
    local args=("$@")
    local index=0

    while [[ $index -lt ${#args[@]} ]]; do
        case "${args[$index]}" in
            --project-dir)
                index=$((index + 1))
                PROJECT_DIR="${args[$index]}"
                ;;
        esac
        index=$((index + 1))
    done
}

twentyi_parse_args() {
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --project-dir)
                shift
                PROJECT_DIR="$1"
                ;;
            --site-name)
                shift
                SITE_NAME="$1"
                ;;
            --site-hostname)
                shift
                SITE_HOSTNAME="$1"
                ;;
            --site-suffix)
                shift
                SITE_SUFFIX="$1"
                ;;
            --docroot|--document-root)
                shift
                DOCROOT="$1"
                ;;
            --php-version)
                shift
                PHP_VERSION="$1"
                ;;
            --mysql-database)
                shift
                MYSQL_DATABASE="$1"
                ;;
            --mysql-user)
                shift
                MYSQL_USER="$1"
                ;;
            --mysql-password)
                shift
                MYSQL_PASSWORD="$1"
                ;;
            --mysql-port)
                shift
                MYSQL_PORT="$1"
                ;;
            --pma-port)
                shift
                PMA_PORT="$1"
                ;;
            --all)
                TWENTYI_ALL=1
                ;;
            --dry-run)
                TWENTYI_DRY_RUN=1
                ;;
            --help|-h)
                twentyi_usage
                exit 0
                ;;
            version=*)
                PHP_VERSION="${1#version=}"
                ;;
            --)
                shift
                break
                ;;
            *)
                if [[ -z "${TWENTYI_POSITIONAL_1:-}" ]]; then
                    TWENTYI_POSITIONAL_1="$1"
                else
                    printf 'Error: unrecognized argument: %s\n' "$1" >&2
                    exit 1
                fi
                ;;
        esac
        shift || true
    done
}

twentyi_init_defaults() {
    PROJECT_DIR="$PWD"
    STACK_HOME="$(twentyi_default_stack_home)"
    TWENTYI_STATE_DIR="${STACK_STATE_DIR:-$STACK_HOME/.20i-state}"
    TWENTYI_STACK_FILE="$STACK_HOME/docker-compose.yml"
    TWENTYI_SHARED_STACK_FILE="$STACK_HOME/docker-compose.shared.yml"
    TWENTYI_DRY_RUN=0
    TWENTYI_ALL=0
    SHARED_GATEWAY_NETWORK="${SHARED_GATEWAY_NETWORK:-twentyi-shared}"
    SHARED_GATEWAY_HTTP_PORT="${SHARED_GATEWAY_HTTP_PORT:-80}"
    SHARED_GATEWAY_HTTPS_PORT="${SHARED_GATEWAY_HTTPS_PORT:-443}"
    SHARED_GATEWAY_COMPOSE_PROJECT_NAME="${SHARED_GATEWAY_COMPOSE_PROJECT_NAME:-20i-shared}"
    SHARED_GATEWAY_CONFIG_FILE="$(twentyi_shared_gateway_config_file)"
    DOCROOT_RELATIVE=""
    CONTAINER_SITE_ROOT=""
    CONTAINER_DOCROOT=""
    MYSQL_VERSION="${MYSQL_VERSION:-10.6}"
    MYSQL_ROOT_PASSWORD="${MYSQL_ROOT_PASSWORD:-root}"
    MYSQL_DATABASE="${MYSQL_DATABASE:-devdb}"
    MYSQL_USER="${MYSQL_USER:-devuser}"
    MYSQL_PASSWORD="${MYSQL_PASSWORD:-devpass}"
    PHP_VERSION="${PHP_VERSION:-8.5}"
}

twentyi_load_stack_and_project_config() {
    PROJECT_DIR="$(twentyi_abs_dir "$PROJECT_DIR")"
    STACK_HOME="$(twentyi_abs_dir "$STACK_HOME")"
    TWENTYI_STATE_DIR="$(twentyi_abs_path_from_base "$STACK_HOME" "$TWENTYI_STATE_DIR")"
    TWENTYI_STACK_FILE="$STACK_HOME/docker-compose.yml"
    TWENTYI_SHARED_STACK_FILE="$STACK_HOME/docker-compose.shared.yml"

    if [[ ! -f "$TWENTYI_STACK_FILE" ]]; then
        printf 'Error: docker-compose.yml not found in %s\n' "$STACK_HOME" >&2
        exit 1
    fi

    if [[ ! -f "$TWENTYI_SHARED_STACK_FILE" ]]; then
        printf 'Error: docker-compose.shared.yml not found in %s\n' "$STACK_HOME" >&2
        exit 1
    fi

    mkdir -p "$TWENTYI_STATE_DIR/projects"

    twentyi_load_env_file "$STACK_HOME/.env" preserve
    twentyi_load_env_file "$PROJECT_DIR/.20i-local" override
}

twentyi_finalize_context() {
    PROJECT_DIR="$(twentyi_abs_dir "$PROJECT_DIR")"
    STACK_HOME="$(twentyi_abs_dir "$STACK_HOME")"
    TWENTYI_STATE_DIR="$(twentyi_abs_path_from_base "$STACK_HOME" "$TWENTYI_STATE_DIR")"
    TWENTYI_STACK_FILE="$STACK_HOME/docker-compose.yml"
    TWENTYI_SHARED_STACK_FILE="$STACK_HOME/docker-compose.shared.yml"
    SHARED_GATEWAY_CONFIG_FILE="$(twentyi_shared_gateway_config_file)"

    PROJECT_NAME="${SITE_NAME:-$(basename "$PROJECT_DIR")}"
    PROJECT_SLUG="$(twentyi_slugify "$PROJECT_NAME")"
    COMPOSE_PROJECT_NAME="${COMPOSE_PROJECT_NAME:-20i-$PROJECT_SLUG}"
    WEB_NETWORK_ALIAS="${WEB_NETWORK_ALIAS:-twentyi-$PROJECT_SLUG-web}"
    CONTAINER_SITE_ROOT="/home/sites/$PROJECT_SLUG"

    if [[ "$MYSQL_DATABASE" == "devdb" ]]; then
        MYSQL_DATABASE="$PROJECT_SLUG"
    fi
    if [[ "$MYSQL_USER" == "devuser" ]]; then
        MYSQL_USER="$PROJECT_SLUG"
    fi

    twentyi_resolve_docroot
    if [[ -n "$DOCROOT_RELATIVE" ]]; then
        CONTAINER_DOCROOT="$CONTAINER_SITE_ROOT/$DOCROOT_RELATIVE"
    else
        CONTAINER_DOCROOT="$CONTAINER_SITE_ROOT"
    fi
    twentyi_resolve_hostname
    twentyi_resolve_ports

    if [[ "$TWENTYI_COMMAND" == "up" || "$TWENTYI_COMMAND" == "attach" ]]; then
        twentyi_validate_requested_ports
    fi

    twentyi_validate_collision
}

twentyi_up_like() {
    ATTACHMENT_STATE="attached"
    twentyi_export_runtime_env

    printf 'Resolved runtime configuration for %s\n' "$TWENTYI_COMMAND"
    twentyi_print_runtime_summary
    twentyi_note_phase_status

    if [[ "$TWENTYI_DRY_RUN" -eq 1 ]]; then
        printf 'Dry run: docker network create %s (if missing)\n' "$SHARED_GATEWAY_NETWORK"
        printf 'Dry run: docker compose --env-file %s -f %s -p %s up -d\n' "$(twentyi_shared_env_file)" "$TWENTYI_SHARED_STACK_FILE" "$SHARED_GATEWAY_COMPOSE_PROJECT_NAME"
        printf 'Dry run: docker compose -f %s -p %s up -d\n' "$TWENTYI_STACK_FILE" "$COMPOSE_PROJECT_NAME"
        return 0
    fi

    twentyi_ensure_shared_infra
    twentyi_compose up -d
    twentyi_write_state
    twentyi_update_gateway_route "$PROJECT_SLUG"

    printf 'Attached: %s\n' "$HOSTNAME"
    printf 'Current access URL: http://localhost:%s\n' "$SHARED_GATEWAY_HTTP_PORT"
}

twentyi_down_like() {
    local state_file

    state_file="$(twentyi_project_state_file)"
    if [[ -f "$state_file" ]]; then
        unset PROJECT_NAME PROJECT_SLUG PROJECT_DIR DOCROOT DOCROOT_RELATIVE HOSTNAME SITE_SUFFIX COMPOSE_PROJECT_NAME HOST_PORT MYSQL_PORT PMA_PORT MYSQL_DATABASE MYSQL_USER MYSQL_PASSWORD MYSQL_VERSION MYSQL_ROOT_PASSWORD PHP_VERSION ATTACHMENT_STATE WEB_NETWORK_ALIAS CONTAINER_SITE_ROOT CONTAINER_DOCROOT
        twentyi_load_state_file "$state_file"
    fi

    if [[ "$TWENTYI_DRY_RUN" -eq 1 ]]; then
        printf 'Dry run: docker compose -f %s -p %s down\n' "$TWENTYI_STACK_FILE" "$COMPOSE_PROJECT_NAME"
        return 0
    fi

    twentyi_export_runtime_env
    twentyi_require_docker
    twentyi_compose down

    if [[ "$TWENTYI_COMMAND" == "detach" ]]; then
        twentyi_remove_state
        twentyi_update_gateway_route
        printf 'Detached: %s\n' "$PROJECT_NAME"
    else
        ATTACHMENT_STATE="down"
        twentyi_write_state
        twentyi_update_gateway_route
        printf 'Stopped: %s\n' "$PROJECT_NAME"
    fi
}

twentyi_down_all() {
    local state_file

    if [[ "$TWENTYI_DRY_RUN" -eq 1 ]]; then
        printf 'Dry run: stop all attached projects and remove %s\n' "$TWENTYI_STATE_DIR"
        return 0
    fi

    twentyi_require_docker

    for state_file in "$TWENTYI_STATE_DIR"/projects/*.env; do
        [[ -e "$state_file" ]] || continue
        unset PROJECT_NAME PROJECT_SLUG PROJECT_DIR DOCROOT DOCROOT_RELATIVE HOSTNAME SITE_SUFFIX COMPOSE_PROJECT_NAME MYSQL_PORT PMA_PORT MYSQL_DATABASE MYSQL_USER MYSQL_PASSWORD MYSQL_VERSION MYSQL_ROOT_PASSWORD PHP_VERSION ATTACHMENT_STATE WEB_NETWORK_ALIAS CONTAINER_SITE_ROOT CONTAINER_DOCROOT
        twentyi_load_state_file "$state_file"
        twentyi_export_runtime_env
        twentyi_compose down || true
    done

    if [[ -f "$(twentyi_shared_env_file)" ]]; then
        twentyi_export_shared_env
        twentyi_shared_compose down || true
    fi

    docker network rm "$SHARED_GATEWAY_NETWORK" >/dev/null 2>&1 || true
    rm -rf "$TWENTYI_STATE_DIR"
    printf 'Global teardown complete\n'
}

twentyi_status() {
    local state_file
    local found=0

    printf '20i stack status\n'
    printf 'Stack home: %s\n' "$STACK_HOME"
    printf 'State dir: %s\n' "$TWENTYI_STATE_DIR"
    printf 'Shared gateway: %s\n' "$(twentyi_shared_gateway_status)"
    printf 'Shared network: %s\n' "$SHARED_GATEWAY_NETWORK"
    twentyi_note_phase_status
    printf '\n'

    for state_file in "$TWENTYI_STATE_DIR"/projects/*.env; do
        [[ -e "$state_file" ]] || continue
        found=1
        unset PROJECT_NAME PROJECT_SLUG PROJECT_DIR DOCROOT DOCROOT_RELATIVE HOSTNAME SITE_SUFFIX COMPOSE_PROJECT_NAME MYSQL_PORT PMA_PORT MYSQL_DATABASE MYSQL_USER MYSQL_PASSWORD MYSQL_VERSION MYSQL_ROOT_PASSWORD PHP_VERSION ATTACHMENT_STATE WEB_NETWORK_ALIAS CONTAINER_SITE_ROOT CONTAINER_DOCROOT
        twentyi_load_state_file "$state_file"

        printf '%s\n' "[$PROJECT_NAME]"
        printf '  state: %s\n' "$ATTACHMENT_STATE"
        printf '  hostname: %s\n' "$HOSTNAME"
        printf '  access: http://localhost:%s\n' "$SHARED_GATEWAY_HTTP_PORT"
        printf '  gateway alias: %s\n' "$WEB_NETWORK_ALIAS"
        printf '  docroot: %s\n' "$DOCROOT"
        printf '  container docroot: %s\n' "$CONTAINER_DOCROOT"
        printf '  project dir: %s\n' "$PROJECT_DIR"
        printf '  docker: %s\n' "$(twentyi_docker_status "$COMPOSE_PROJECT_NAME")"
        printf '\n'
    done

    if [[ "$found" -eq 0 ]]; then
        printf 'No attached projects recorded.\n'
    fi
}

twentyi_logs() {
    local service_name="${TWENTYI_POSITIONAL_1:-}"
    local state_file

    state_file="$(twentyi_project_state_file)"
    if [[ -f "$state_file" ]]; then
        unset PROJECT_NAME PROJECT_SLUG PROJECT_DIR DOCROOT DOCROOT_RELATIVE HOSTNAME SITE_SUFFIX COMPOSE_PROJECT_NAME HOST_PORT MYSQL_PORT PMA_PORT MYSQL_DATABASE MYSQL_USER MYSQL_PASSWORD MYSQL_VERSION MYSQL_ROOT_PASSWORD PHP_VERSION ATTACHMENT_STATE WEB_NETWORK_ALIAS CONTAINER_SITE_ROOT CONTAINER_DOCROOT
        twentyi_load_state_file "$state_file"
    fi

    twentyi_require_docker

    if [[ -n "$service_name" ]]; then
        twentyi_compose logs -f "$service_name"
    else
        twentyi_compose logs -f
    fi
}

twentyi_main() {
    TWENTYI_COMMAND="$1"
    shift

    twentyi_init_defaults
    twentyi_parse_initial_args "$@"
    twentyi_load_stack_and_project_config
    twentyi_parse_args "$@"
    twentyi_finalize_context

    case "$TWENTYI_COMMAND" in
        up|attach)
            twentyi_up_like
            ;;
        down|detach)
            if [[ "$TWENTYI_ALL" -eq 1 ]]; then
                twentyi_down_all
            else
                twentyi_down_like
            fi
            ;;
        status)
            twentyi_status
            ;;
        logs)
            twentyi_logs
            ;;
        *)
            printf 'Error: unsupported command %s\n' "$TWENTYI_COMMAND" >&2
            exit 1
            ;;
    esac
}