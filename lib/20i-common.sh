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

twentyi_registry_file() {
    printf '%s/registry.tsv' "$TWENTYI_STATE_DIR"
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

twentyi_unset_project_state_vars() {
    unset PROJECT_NAME PROJECT_SLUG PROJECT_DIR DOCROOT DOCROOT_RELATIVE HOSTNAME SITE_SUFFIX COMPOSE_PROJECT_NAME HOST_PORT MYSQL_PORT PMA_PORT MYSQL_DATABASE MYSQL_USER MYSQL_PASSWORD MYSQL_VERSION MYSQL_ROOT_PASSWORD PHP_VERSION ATTACHMENT_STATE WEB_NETWORK_ALIAS CONTAINER_SITE_ROOT CONTAINER_DOCROOT PROJECT_RUNTIME_NETWORK PROJECT_DATABASE_VOLUME NGINX_CONTAINER_NAME NGINX_CONTAINER_ID NGINX_CONTAINER_STATUS APACHE_CONTAINER_NAME APACHE_CONTAINER_ID APACHE_CONTAINER_STATUS MARIADB_CONTAINER_NAME MARIADB_CONTAINER_ID MARIADB_CONTAINER_STATUS PHPMYADMIN_CONTAINER_NAME PHPMYADMIN_CONTAINER_ID PHPMYADMIN_CONTAINER_STATUS RUNTIME_CONTAINER_SUMMARY
}

twentyi_registry_escape() {
    printf '%s' "$1" | tr '\t\r\n' '   '
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
        if lsof -nP -iTCP:"$port" -sTCP:LISTEN >/dev/null 2>&1; then
            return 0
        fi
    fi

    if command -v netstat >/dev/null 2>&1; then
        netstat -anv -p tcp 2>/dev/null | grep -E "[.:]$port[[:space:]]" | grep -q LISTEN
        return $?
    fi

    return 1
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

twentyi_resolve_shared_gateway_ports() {
    local requested_https_port="${SHARED_GATEWAY_HTTPS_PORT:-}"
    local shared_env_file existing_https_port

    TWENTYI_HTTPS_PORT_AUTO_FALLBACK=0

    if ! twentyi_tls_available; then
        return 0
    fi

    if [[ "${LOCAL_DNS_SUFFIX:-${SITE_SUFFIX:-}}" == "dev" ]]; then
        if [[ -z "$requested_https_port" || "$requested_https_port" == "443" ]]; then
            SHARED_GATEWAY_HTTPS_PORT=8443
            TWENTYI_HTTPS_PORT_AUTO_FALLBACK=1
        fi
        return 0
    fi

    requested_https_port="${SHARED_GATEWAY_HTTPS_PORT:-443}"

    shared_env_file="$(twentyi_shared_env_file)"
    if [[ -f "$shared_env_file" ]]; then
        existing_https_port="$(grep '^SHARED_GATEWAY_HTTPS_PORT=' "$shared_env_file" | head -1 | cut -d= -f2-)"
        if [[ -n "$existing_https_port" && "$requested_https_port" == "443" ]]; then
            SHARED_GATEWAY_HTTPS_PORT="$existing_https_port"
            if [[ "$existing_https_port" != "443" ]]; then
                TWENTYI_HTTPS_PORT_AUTO_FALLBACK=1
            fi
            return 0
        fi
    fi

    if [[ "$requested_https_port" == "443" ]] && twentyi_port_in_use 443; then
        SHARED_GATEWAY_HTTPS_PORT="$(twentyi_find_available_port SHARED_GATEWAY_HTTPS_PORT 8443)"
        TWENTYI_HTTPS_PORT_AUTO_FALLBACK=1
    fi
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
    export PROJECT_NAME
    export PROJECT_SLUG
    export PROJECT_HOSTNAME="$HOSTNAME"
    export PROJECT_DOCROOT="$DOCROOT"
    export PROJECT_RUNTIME_NETWORK="${COMPOSE_PROJECT_NAME}-runtime"
    export PROJECT_DATABASE_VOLUME="${COMPOSE_PROJECT_NAME}-db-data"
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

twentyi_dns_preview_config_file() {
    printf '%s/shared/dnsmasq-%s.conf' "$TWENTYI_STATE_DIR" "$LOCAL_DNS_SUFFIX"
}

twentyi_dns_preview_resolver_file() {
    printf '%s/shared/resolver-%s.conf' "$TWENTYI_STATE_DIR" "$LOCAL_DNS_SUFFIX"
}

twentyi_dns_resolver_file() {
    printf '/etc/resolver/%s' "$LOCAL_DNS_SUFFIX"
}

twentyi_dnsmasq_conf_dir() {
    local brew_prefix

    if ! command -v brew >/dev/null 2>&1; then
        return 1
    fi

    brew_prefix="$(brew --prefix 2>/dev/null)"
    [[ -n "$brew_prefix" ]] || return 1
    printf '%s/etc/dnsmasq.d' "$brew_prefix"
}

twentyi_dnsmasq_main_conf_file() {
    local brew_prefix

    if ! command -v brew >/dev/null 2>&1; then
        return 1
    fi

    brew_prefix="$(brew --prefix 2>/dev/null)"
    [[ -n "$brew_prefix" ]] || return 1
    printf '%s/etc/dnsmasq.conf' "$brew_prefix"
}

twentyi_dnsmasq_managed_file() {
    local conf_dir

    conf_dir="$(twentyi_dnsmasq_conf_dir)" || return 1
    printf '%s/20i-stack-%s.conf' "$conf_dir" "$LOCAL_DNS_SUFFIX"
}

twentyi_write_dns_support_files() {
    local preview_config preview_resolver

    preview_config="$(twentyi_dns_preview_config_file)"
    preview_resolver="$(twentyi_dns_preview_resolver_file)"
    mkdir -p "$(dirname "$preview_config")"

    cat > "$preview_config" <<EOF
port=$LOCAL_DNS_PORT
listen-address=$LOCAL_DNS_IP
bind-interfaces
address=/.$LOCAL_DNS_SUFFIX/$LOCAL_DNS_IP
EOF

    cat > "$preview_resolver" <<EOF
nameserver $LOCAL_DNS_IP
port $LOCAL_DNS_PORT
EOF
}

twentyi_dns_service_running() {
    if command -v lsof >/dev/null 2>&1; then
        lsof -nP -iUDP:"$LOCAL_DNS_PORT" 2>/dev/null | grep -qi dnsmasq
    else
        return 1
    fi
}

twentyi_dns_status() {
    local managed_file resolver_file

    [[ "$(uname -s)" == "Darwin" ]] || {
        printf 'unsupported-os'
        return 0
    }

    managed_file="$(twentyi_dnsmasq_managed_file 2>/dev/null || true)"
    resolver_file="$(twentyi_dns_resolver_file)"

    if ! command -v brew >/dev/null 2>&1; then
        printf 'brew-missing'
        return 0
    fi

    if ! brew list dnsmasq >/dev/null 2>&1; then
        printf 'dnsmasq-missing'
        return 0
    fi

    if [[ -z "$managed_file" || ! -f "$managed_file" ]]; then
        printf 'dnsmasq-config-missing'
        return 0
    fi

    if ! twentyi_dns_service_running; then
        printf 'dnsmasq-stopped'
        return 0
    fi

    if [[ ! -f "$resolver_file" ]]; then
        printf 'resolver-missing'
        return 0
    fi

    if ! grep -Fq "nameserver $LOCAL_DNS_IP" "$resolver_file" || ! grep -Fq "port $LOCAL_DNS_PORT" "$resolver_file"; then
        printf 'resolver-mismatch'
        return 0
    fi

    printf 'ready'
}

twentyi_dns_status_message() {
    case "$(twentyi_dns_status)" in
        ready)
            printf 'ready (%s on %s:%s for .%s)' "$LOCAL_DNS_PROVIDER" "$LOCAL_DNS_IP" "$LOCAL_DNS_PORT" "$LOCAL_DNS_SUFFIX"
            ;;
        unsupported-os)
            printf 'unsupported-os'
            ;;
        brew-missing)
            printf 'brew missing'
            ;;
        dnsmasq-missing)
            printf 'dnsmasq not installed'
            ;;
        dnsmasq-config-missing)
            printf 'dnsmasq config missing'
            ;;
        dnsmasq-stopped)
            printf 'dnsmasq not running on %s:%s' "$LOCAL_DNS_IP" "$LOCAL_DNS_PORT"
            ;;
        resolver-missing)
            printf 'resolver file missing: %s' "$(twentyi_dns_resolver_file)"
            ;;
        resolver-mismatch)
            printf 'resolver file does not point at %s:%s' "$LOCAL_DNS_IP" "$LOCAL_DNS_PORT"
            ;;
        *)
            printf 'unknown'
            ;;
    esac
}

twentyi_warn_if_dns_not_ready() {
    local dns_status

    dns_status="$(twentyi_dns_status)"
    [[ "$dns_status" == "ready" || "$dns_status" == "unsupported-os" ]] && return 0

    printf 'Local DNS: %s\n' "$(twentyi_dns_status_message)" >&2
    printf 'Run stacklane --dns-setup to bootstrap .%s resolution on macOS.\n' "$LOCAL_DNS_SUFFIX" >&2
}

twentyi_dns_setup() {
    local preview_config preview_resolver managed_file resolver_file resolver_dir dnsmasq_main_conf conf_dir include_line

    if [[ "$(uname -s)" != "Darwin" ]]; then
        printf 'Error: local DNS bootstrap is currently implemented for macOS only\n' >&2
        exit 1
    fi

    if ! command -v brew >/dev/null 2>&1; then
        printf 'Error: Homebrew is required for local DNS bootstrap\n' >&2
        exit 1
    fi

    if ! brew list dnsmasq >/dev/null 2>&1; then
        printf 'Error: dnsmasq is not installed. Run: brew install dnsmasq\n' >&2
        exit 1
    fi

    twentyi_write_dns_support_files
    preview_config="$(twentyi_dns_preview_config_file)"
    preview_resolver="$(twentyi_dns_preview_resolver_file)"
    managed_file="$(twentyi_dnsmasq_managed_file)"
    resolver_file="$(twentyi_dns_resolver_file)"
    resolver_dir="$(dirname "$resolver_file")"

    mkdir -p "$(dirname "$managed_file")"
    rm -f "$(dirname "$managed_file")"/20i-stack-*.conf
    cp "$preview_config" "$managed_file"

    dnsmasq_main_conf="$(twentyi_dnsmasq_main_conf_file)"
    conf_dir="$(twentyi_dnsmasq_conf_dir)"
    include_line="conf-dir=$conf_dir,*.conf"

    if [[ ! -f "$dnsmasq_main_conf" ]]; then
        printf 'Error: dnsmasq main config not found: %s\n' "$dnsmasq_main_conf" >&2
        exit 1
    fi

    if ! grep -Fqx "$include_line" "$dnsmasq_main_conf"; then
        printf '\n# 20i stack managed include\n%s\n' "$include_line" >> "$dnsmasq_main_conf"
    fi

    if ! brew services restart dnsmasq >/dev/null 2>&1; then
        brew services start dnsmasq >/dev/null 2>&1 || {
            printf 'Error: could not start dnsmasq via Homebrew services\n' >&2
            exit 1
        }
    fi

    # For .dev (HSTS-preloaded TLD), generate an exact-hostname TLS cert
    # before the privileged resolver copy. That way a fresh machine still gets
    # the expected cert artifacts even if /etc/resolver/<suffix> needs a manual
    # approval step.
    twentyi_ensure_tls_cert

    _twentyi_resolver_needs_update() {
        [[ ! -f "$resolver_file" ]] || ! diff -q "$preview_resolver" "$resolver_file" >/dev/null 2>&1
    }

    if _twentyi_resolver_needs_update; then
        if [[ -w "$resolver_dir" || ( ! -e "$resolver_dir" && -w /etc ) ]]; then
            mkdir -p "$resolver_dir"
            cp "$preview_resolver" "$resolver_file"
        elif command -v sudo >/dev/null 2>&1 && sudo -n true >/dev/null 2>&1; then
            sudo mkdir -p "$resolver_dir"
            sudo cp "$preview_resolver" "$resolver_file"
        elif command -v osascript >/dev/null 2>&1; then
            local admin_command
            admin_command="mkdir -p $(printf '%q' "$resolver_dir") && cp $(printf '%q' "$preview_resolver") $(printf '%q' "$resolver_file")"
            if ! osascript -e "do shell script \"$admin_command\" with administrator privileges" >/dev/null 2>&1; then
                printf 'Error: administrator approval was required to install %s\n' "$resolver_file" >&2
                exit 1
            fi
        else
            printf 'Resolver file needs elevated privileges. Run:\n' >&2
            printf '  sudo mkdir -p %s && sudo cp %s %s\n' "$resolver_dir" "$preview_resolver" "$resolver_file" >&2
            exit 1
        fi
    fi

    if [[ "$(twentyi_dns_status)" != "ready" ]]; then
        printf 'Error: local DNS bootstrap did not reach a ready state (%s)\n' "$(twentyi_dns_status_message)" >&2
        exit 1
    fi

    printf 'Local DNS ready for .%s via %s\n' "$LOCAL_DNS_SUFFIX" "$LOCAL_DNS_PROVIDER"
}

twentyi_hostname_route_url() {
    if twentyi_tls_available; then
        if [[ "$SHARED_GATEWAY_HTTPS_PORT" == "443" ]]; then
            printf 'https://%s' "$HOSTNAME"
        else
            printf 'https://%s:%s' "$HOSTNAME" "$SHARED_GATEWAY_HTTPS_PORT"
        fi
    elif [[ "$SHARED_GATEWAY_HTTP_PORT" == "80" ]]; then
        printf 'http://%s' "$HOSTNAME"
    else
        printf 'http://%s:%s' "$HOSTNAME" "$SHARED_GATEWAY_HTTP_PORT"
    fi
}

twentyi_gateway_probe_url() {
    if [[ "$SHARED_GATEWAY_HTTP_PORT" == "80" ]]; then
        printf 'http://localhost'
    else
        printf 'http://localhost:%s' "$SHARED_GATEWAY_HTTP_PORT"
    fi
}

twentyi_reset_runtime_identity() {
    NGINX_CONTAINER_NAME=""
    NGINX_CONTAINER_ID=""
    NGINX_CONTAINER_STATUS=""
    APACHE_CONTAINER_NAME=""
    APACHE_CONTAINER_ID=""
    APACHE_CONTAINER_STATUS=""
    MARIADB_CONTAINER_NAME=""
    MARIADB_CONTAINER_ID=""
    MARIADB_CONTAINER_STATUS=""
    PHPMYADMIN_CONTAINER_NAME=""
    PHPMYADMIN_CONTAINER_ID=""
    PHPMYADMIN_CONTAINER_STATUS=""
    RUNTIME_CONTAINER_SUMMARY=""
}

twentyi_capture_runtime_identity() {
    local service line container_name container_id container_status prefix summary=()

    twentyi_reset_runtime_identity

    for service in nginx apache mariadb phpmyadmin; do
        line="$(docker ps -a --filter "label=com.docker.compose.project=$COMPOSE_PROJECT_NAME" --filter "label=com.docker.compose.service=$service" --format '{{.Names}}|{{.ID}}|{{.Status}}' | head -n 1)"

        if [[ -z "$line" ]]; then
            printf 'Error: expected %s container for compose project %s was not created\n' "$service" "$COMPOSE_PROJECT_NAME" >&2
            return 1
        fi

        container_name="${line%%|*}"
        line="${line#*|}"
        container_id="${line%%|*}"
        container_status="${line#*|}"

        case "$service" in
            nginx) prefix="NGINX" ;;
            apache) prefix="APACHE" ;;
            mariadb) prefix="MARIADB" ;;
            phpmyadmin) prefix="PHPMYADMIN" ;;
        esac

        printf -v "${prefix}_CONTAINER_NAME" '%s' "$container_name"
        printf -v "${prefix}_CONTAINER_ID" '%s' "$container_id"
        printf -v "${prefix}_CONTAINER_STATUS" '%s' "$container_status"
        summary+=("$service=$container_name#$container_id [$container_status]")
    done

    RUNTIME_CONTAINER_SUMMARY="${summary[*]}"
}

twentyi_refresh_registry() {
    local registry_file

    registry_file="$(twentyi_registry_file)"
    mkdir -p "$TWENTYI_STATE_DIR"

    # Run the state-file iteration in a subshell so that loading each project's
    # env does not clobber the current project's exported variables in the
    # parent shell (which is critical when called from within 20i-up/attach).
    (
        local state_file
        : > "$registry_file"

        printf 'project_slug\tattachment_state\tproject_name\tproject_dir\thostname\tdocroot\tcompose_project\truntime_network\tdb_volume\tphp_version\tmysql_database\tmysql_port\tpma_port\tweb_network_alias\tcontainer_summary\n' >> "$registry_file"

        for state_file in "$TWENTYI_STATE_DIR"/projects/*.env; do
            [[ -e "$state_file" ]] || continue
            twentyi_unset_project_state_vars
            twentyi_load_state_file "$state_file"
            printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n' \
                "$(twentyi_registry_escape "$PROJECT_SLUG")" \
                "$(twentyi_registry_escape "$ATTACHMENT_STATE")" \
                "$(twentyi_registry_escape "$PROJECT_NAME")" \
                "$(twentyi_registry_escape "$PROJECT_DIR")" \
                "$(twentyi_registry_escape "$HOSTNAME")" \
                "$(twentyi_registry_escape "$DOCROOT")" \
                "$(twentyi_registry_escape "$COMPOSE_PROJECT_NAME")" \
                "$(twentyi_registry_escape "${PROJECT_RUNTIME_NETWORK:-${COMPOSE_PROJECT_NAME}-runtime}")" \
                "$(twentyi_registry_escape "${PROJECT_DATABASE_VOLUME:-${COMPOSE_PROJECT_NAME}-db-data}")" \
                "$(twentyi_registry_escape "$PHP_VERSION")" \
                "$(twentyi_registry_escape "$MYSQL_DATABASE")" \
                "$(twentyi_registry_escape "$MYSQL_PORT")" \
                "$(twentyi_registry_escape "$PMA_PORT")" \
                "$(twentyi_registry_escape "$WEB_NETWORK_ALIAS")" \
                "$(twentyi_registry_escape "${RUNTIME_CONTAINER_SUMMARY:-}")" >> "$registry_file"
        done
    )
}

twentyi_state_file_for_selector() {
    local selector="$1"
    local state_file

    for state_file in "$TWENTYI_STATE_DIR"/projects/*.env; do
        [[ -e "$state_file" ]] || continue
        twentyi_unset_project_state_vars
        twentyi_load_state_file "$state_file"

        if [[ "$selector" == "$PROJECT_SLUG" || "$selector" == "$PROJECT_NAME" || "$selector" == "$HOSTNAME" || "$selector" == "$PROJECT_DIR" ]]; then
            printf '%s' "$state_file"
            return 0
        fi
    done

    return 1
}

twentyi_live_container_summary() {
    local compose_project="$1"
    local service line summary=()

    for service in nginx apache mariadb phpmyadmin; do
        line="$(docker ps -a --filter "label=com.docker.compose.project=$compose_project" --filter "label=com.docker.compose.service=$service" --format '{{.Names}}#{{.ID}} [{{.Status}}]' | head -n 1)"
        if [[ -n "$line" ]]; then
            summary+=("$service=$line")
        fi
    done

    if [[ ${#summary[@]} -eq 0 ]]; then
        printf '%s' ''
    else
        printf '%s' "${summary[*]}"
    fi
}

twentyi_registry_drift_status() {
    local live_summary normalized_recorded normalized_live

    if ! command -v docker >/dev/null 2>&1; then
        printf 'docker-unavailable'
        return 0
    fi

    live_summary="$(twentyi_live_container_summary "$COMPOSE_PROJECT_NAME")"

    if [[ "$ATTACHMENT_STATE" == "attached" && -z "$live_summary" ]]; then
        printf 'attached-but-missing-runtime'
    elif [[ "$ATTACHMENT_STATE" == "down" && -n "$live_summary" ]]; then
        printf 'state-down-but-runtime-present'
    elif [[ -n "${RUNTIME_CONTAINER_SUMMARY:-}" && -n "$live_summary" ]]; then
        # Strip the mutable Docker status text (e.g. "Up 33 seconds") from both
        # sides before comparing so normal uptime changes don't trigger false
        # identity-mismatch warnings.
        normalized_recorded="$(printf '%s' "${RUNTIME_CONTAINER_SUMMARY}" | sed 's/ \[[^]]*\]//g')"
        normalized_live="$(printf '%s' "$live_summary" | sed 's/ \[[^]]*\]//g')"
        if [[ "$normalized_recorded" != "$normalized_live" ]]; then
            printf 'recorded-container-identity-mismatch'
        else
            printf 'none'
        fi
    else
        printf 'none'
    fi
}

twentyi_validate_runtime_registration() {
    twentyi_capture_runtime_identity
    twentyi_write_state
    twentyi_refresh_registry

    if ! grep -Fq "$PROJECT_SLUG"$'\t' "$(twentyi_registry_file)"; then
        printf 'Error: registry update failed for %s\n' "$PROJECT_SLUG" >&2
        exit 1
    fi
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
Runtime network:   ${COMPOSE_PROJECT_NAME}-runtime
DB volume:         ${COMPOSE_PROJECT_NAME}-db-data
Planned hostname:  $HOSTNAME
Gateway alias:     $WEB_NETWORK_ALIAS
Hostname route:    $(twentyi_hostname_route_url)
Gateway probe:     $(twentyi_gateway_probe_url)
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
        printf 'PROJECT_RUNTIME_NETWORK=%q\n' "${PROJECT_RUNTIME_NETWORK:-${COMPOSE_PROJECT_NAME}-runtime}"
        printf 'PROJECT_DATABASE_VOLUME=%q\n' "${PROJECT_DATABASE_VOLUME:-${COMPOSE_PROJECT_NAME}-db-data}"
        printf 'NGINX_CONTAINER_NAME=%q\n' "${NGINX_CONTAINER_NAME:-}"
        printf 'NGINX_CONTAINER_ID=%q\n' "${NGINX_CONTAINER_ID:-}"
        printf 'NGINX_CONTAINER_STATUS=%q\n' "${NGINX_CONTAINER_STATUS:-}"
        printf 'APACHE_CONTAINER_NAME=%q\n' "${APACHE_CONTAINER_NAME:-}"
        printf 'APACHE_CONTAINER_ID=%q\n' "${APACHE_CONTAINER_ID:-}"
        printf 'APACHE_CONTAINER_STATUS=%q\n' "${APACHE_CONTAINER_STATUS:-}"
        printf 'MARIADB_CONTAINER_NAME=%q\n' "${MARIADB_CONTAINER_NAME:-}"
        printf 'MARIADB_CONTAINER_ID=%q\n' "${MARIADB_CONTAINER_ID:-}"
        printf 'MARIADB_CONTAINER_STATUS=%q\n' "${MARIADB_CONTAINER_STATUS:-}"
        printf 'PHPMYADMIN_CONTAINER_NAME=%q\n' "${PHPMYADMIN_CONTAINER_NAME:-}"
        printf 'PHPMYADMIN_CONTAINER_ID=%q\n' "${PHPMYADMIN_CONTAINER_ID:-}"
        printf 'PHPMYADMIN_CONTAINER_STATUS=%q\n' "${PHPMYADMIN_CONTAINER_STATUS:-}"
        printf 'RUNTIME_CONTAINER_SUMMARY=%q\n' "${RUNTIME_CONTAINER_SUMMARY:-}"
    } >> "$state_file"
}

twentyi_remove_state() {
    local state_file
    state_file="$(twentyi_project_state_file)"
    [[ -f "$state_file" ]] && rm -f "$state_file"
    twentyi_refresh_registry
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
    local route_hostname="$1"
    local attempt

    if [[ -z "$route_hostname" || "$route_hostname" == "localhost" ]]; then
        return 0
    fi

    for attempt in $(seq 1 25); do
        if twentyi_shared_compose exec -T gateway sh -c "wget -S --spider --header='Host: $route_hostname' http://127.0.0.1/ 2>&1 | grep -Eq 'HTTP/[0-9.]+ (200|301)'" >/dev/null 2>&1; then
            return 0
        fi
        sleep 0.2
    done

    return 1
}

twentyi_hostname_valid() {
    local hostname="$1"
    local remaining label

    [[ -n "$hostname" && "$hostname" == *.* ]] || return 1

    case "$hostname" in
        *[!A-Za-z0-9.-]*|.*|*..*|*.)
            return 1
            ;;
    esac

    remaining="$hostname"
    while true; do
        label="${remaining%%.*}"
        [[ -n "$label" ]] || return 1
        [[ "$label" != -* && "$label" != *- ]] || return 1
        [[ "$remaining" == *.* ]] || break
        remaining="${remaining#*.}"
    done

    return 0
}

twentyi_alias_valid() {
    local alias_name="$1"

    [[ "$alias_name" =~ ^[a-z0-9]([a-z0-9-]*[a-z0-9])?$ ]]
}

twentyi_gateway_route_lines() {
    local registry_file line
    local project_slug attachment_state project_name project_dir hostname docroot compose_project runtime_network db_volume php_version mysql_database mysql_port pma_port web_network_alias container_summary

    registry_file="$(twentyi_registry_file)"
    [[ -f "$registry_file" ]] || return 0

    while IFS=$'\t' read -r project_slug attachment_state project_name project_dir hostname docroot compose_project runtime_network db_volume php_version mysql_database mysql_port pma_port web_network_alias container_summary; do
        [[ "$project_slug" == "project_slug" ]] && continue
        [[ "$attachment_state" == "attached" ]] || continue

        if ! twentyi_hostname_valid "$hostname"; then
            continue
        fi

        if ! twentyi_alias_valid "$web_network_alias"; then
            continue
        fi

        printf '%s|%s|%s\n' "$hostname" "$web_network_alias" "$project_slug"
    done < "$registry_file"
}

twentyi_gateway_block_for_route() {
    local hostname="$1"
    local route_target="$2"
    local https_redirect_host='https://$host$request_uri'

    if [[ "$SHARED_GATEWAY_HTTPS_PORT" != "443" ]]; then
        https_redirect_host="https://\$host:$SHARED_GATEWAY_HTTPS_PORT\$request_uri"
    fi

    if twentyi_tls_available; then
        # HTTP → HTTPS redirect
        cat <<EOF
server {
    listen 80;
    server_name $hostname;
    return 301 $https_redirect_host;
}
EOF

        # HTTPS proxy block
        cat <<EOF
server {
    listen 443 ssl;
    server_name $hostname;

    ssl_certificate     /etc/nginx/certs/tls.pem;
    ssl_certificate_key /etc/nginx/certs/tls-key.pem;
    ssl_protocols       TLSv1.2 TLSv1.3;
    ssl_ciphers         HIGH:!aNULL:!MD5;

    access_log /var/log/nginx/access.log;
    error_log /var/log/nginx/error.log warn;

    add_header X-20i-Gateway "shared" always;
    add_header X-20i-Route-Target "$route_target" always;
    add_header X-20i-Hostname "$hostname" always;

    location / {
        resolver 127.0.0.11 valid=5s;
        set \$upstream http://$route_target:80;

        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Forwarded-Host \$host;
        proxy_set_header X-Forwarded-Proto https;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_connect_timeout 2s;
        proxy_read_timeout 600s;
        proxy_pass \$upstream;
    }
}

EOF
    else
        cat <<EOF
server {
    listen 80;
    listen 443;
    server_name $hostname;

    access_log /var/log/nginx/access.log;
    error_log /var/log/nginx/error.log warn;

    add_header X-20i-Gateway "shared" always;
    add_header X-20i-Route-Target "$route_target" always;
    add_header X-20i-Hostname "$hostname" always;

    location / {
        # Use Docker's embedded DNS so the upstream is resolved at request-time,
        # not at nginx startup. This lets the gateway start before project
        # containers exist and recover automatically when they come up.
        resolver 127.0.0.11 valid=5s;
        set \$upstream http://$route_target:80;

        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Forwarded-Host \$host;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_connect_timeout 2s;
        proxy_read_timeout 600s;
        proxy_pass \$upstream;
    }
}

EOF
    fi
}

twentyi_write_gateway_config() {
    local preferred_slug="$1"
    local config_file temp_config route_lines=() route_line preferred_hostname="" preferred_target="" hostname route_target route_slug

    config_file="$(twentyi_shared_gateway_config_file)"
    mkdir -p "$(dirname "$config_file")"
    temp_config="$(mktemp "${config_file}.tmp.XXXXXX")"

    while IFS= read -r route_line; do
        [[ -n "$route_line" ]] || continue
        route_lines+=("$route_line")
        hostname="${route_line%%|*}"
        route_target="${route_line#*|}"
        route_target="${route_target%%|*}"
        route_slug="${route_line##*|}"

        if [[ -n "$preferred_slug" && "$route_slug" == "$preferred_slug" ]]; then
            preferred_hostname="$hostname"
            preferred_target="$route_target"
        fi
    done < <(twentyi_gateway_route_lines)

    if [[ ${#route_lines[@]} -eq 0 ]]; then
        if twentyi_tls_available; then
            cat > "$temp_config" <<EOF
server {
    listen 80 default_server;
    listen 443 ssl default_server;
    server_name _;

    ssl_certificate     /etc/nginx/certs/tls.pem;
    ssl_certificate_key /etc/nginx/certs/tls-key.pem;
    ssl_protocols       TLSv1.2 TLSv1.3;
    ssl_ciphers         HIGH:!aNULL:!MD5;

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
        return 503 "20i shared gateway has no hostname routes.\\n";
    }
}
EOF
        else
            cat > "$temp_config" <<EOF
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
        return 503 "20i shared gateway has no hostname routes.\\n";
    }
}
EOF
        fi
    mv "$temp_config" "$config_file"
        TWENTYI_GATEWAY_PROBE_TARGET="twentyi-no-route"
        TWENTYI_GATEWAY_PROBE_HOSTNAME="localhost"
        return 0
    fi

    if twentyi_tls_available; then
    cat > "$temp_config" <<EOF
server {
    listen 80 default_server;
    listen 443 ssl default_server;
    server_name _;

    ssl_certificate     /etc/nginx/certs/tls.pem;
    ssl_certificate_key /etc/nginx/certs/tls-key.pem;
    ssl_protocols       TLSv1.2 TLSv1.3;
    ssl_ciphers         HIGH:!aNULL:!MD5;

    access_log /var/log/nginx/access.log;
    error_log /var/log/nginx/error.log warn;

    add_header X-20i-Gateway "shared" always;
    add_header X-20i-Route-Target "unmatched-host" always;

    location = /__20i_gateway_health {
        default_type text/plain;
        return 200 "gateway ok\\n";
    }

    location / {
        default_type text/plain;
        add_header X-20i-Route-State "unmatched-host" always;
        return 404 "20i shared gateway has no route for host '\$host'.\\n";
    }
}
EOF
    else
    cat > "$temp_config" <<EOF
server {
    listen 80 default_server;
    listen 443 default_server;
    server_name _;

    access_log /var/log/nginx/access.log;
    error_log /var/log/nginx/error.log warn;

    add_header X-20i-Gateway "shared" always;
    add_header X-20i-Route-Target "unmatched-host" always;

    location = /__20i_gateway_health {
        default_type text/plain;
        return 200 "gateway ok\\n";
    }

    location / {
        default_type text/plain;
        add_header X-20i-Route-State "unmatched-host" always;
        return 404 "20i shared gateway has no route for host '\$host'.\\n";
    }
}
EOF
    fi

    for route_line in "${route_lines[@]}"; do
        hostname="${route_line%%|*}"
        route_target="${route_line#*|}"
        route_target="${route_target%%|*}"
        twentyi_gateway_block_for_route "$hostname" "$route_target" >> "$temp_config"
    done

    mv "$temp_config" "$config_file"

    if [[ -z "$preferred_target" ]]; then
        hostname="${route_lines[0]%%|*}"
        preferred_target="${route_lines[0]#*|}"
        preferred_target="${preferred_target%%|*}"
        preferred_hostname="$hostname"
    fi

    TWENTYI_GATEWAY_PROBE_TARGET="$preferred_target"
    TWENTYI_GATEWAY_PROBE_HOSTNAME="$preferred_hostname"
}

twentyi_tls_cert_file() {
    printf '%s/shared/certs/tls.pem' "$TWENTYI_STATE_DIR"
}

twentyi_tls_key_file() {
    printf '%s/shared/certs/tls-key.pem' "$TWENTYI_STATE_DIR"
}

twentyi_tls_hostnames() {
    local registry_file
    local project_slug attachment_state project_name project_dir hostname docroot compose_project runtime_network db_volume php_version mysql_database mysql_port pma_port web_network_alias container_summary

    printf 'localhost\n127.0.0.1\n'

    if [[ -n "${HOSTNAME:-}" ]] && twentyi_hostname_valid "$HOSTNAME"; then
        printf '%s\n' "$HOSTNAME"
    fi

    registry_file="$(twentyi_registry_file)"
    [[ -f "$registry_file" ]] || return 0

    while IFS=$'\t' read -r project_slug attachment_state project_name project_dir hostname docroot compose_project runtime_network db_volume php_version mysql_database mysql_port pma_port web_network_alias container_summary; do
        [[ "$project_slug" == "project_slug" ]] && continue
        [[ "$attachment_state" == "attached" ]] || continue
        if twentyi_hostname_valid "$hostname"; then
            printf '%s\n' "$hostname"
        fi
    done < "$registry_file"
}

twentyi_ensure_tls_cert() {
    local cert_file key_file certs_dir host_list_file cert_summary cert_target
    local cert_targets=()

    if [[ "$LOCAL_DNS_SUFFIX" != "dev" ]]; then
        return 0
    fi

    if ! command -v mkcert >/dev/null 2>&1; then
        printf 'Error: mkcert is required for .dev TLS. Run: brew install mkcert && mkcert -install\n' >&2
        exit 1
    fi

    cert_file="$(twentyi_tls_cert_file)"
    key_file="$(twentyi_tls_key_file)"
    certs_dir="$(dirname "$cert_file")"
    mkdir -p "$certs_dir"

    host_list_file="$(mktemp)"
    twentyi_tls_hostnames | awk 'NF' | sort -u > "$host_list_file"

    while IFS= read -r cert_target; do
        [[ -n "$cert_target" ]] || continue
        cert_targets+=("$cert_target")
    done < "$host_list_file"
    rm -f "$host_list_file"

    if [[ ${#cert_targets[@]} -eq 0 ]]; then
        cert_targets=("localhost" "127.0.0.1")
    fi

    mkcert -cert-file "$cert_file" -key-file "$key_file" "${cert_targets[@]}" >/dev/null 2>&1 || {
        printf 'Error: mkcert could not generate a TLS certificate for the local .dev hostnames\n' >&2
        exit 1
    }

    cert_summary="$(printf '%s, ' "${cert_targets[@]}")"
    cert_summary="${cert_summary%, }"
    printf 'TLS certificate ready for %s (expires %s)\n' "$cert_summary" "$(openssl x509 -noout -enddate -in "$cert_file" 2>/dev/null | cut -d= -f2 || echo 'see cert file')"
}

twentyi_tls_available() {
    [[ -f "$(twentyi_tls_cert_file)" && -f "$(twentyi_tls_key_file)" ]]
}

twentyi_write_shared_env() {
    local shared_env_file certs_dir

    shared_env_file="$(twentyi_shared_env_file)"
    certs_dir="$(dirname "$(twentyi_tls_cert_file)")"
    mkdir -p "$(dirname "$shared_env_file")"
    : > "$shared_env_file"

    {
        printf 'SHARED_GATEWAY_NETWORK=%q\n' "$SHARED_GATEWAY_NETWORK"
        printf 'SHARED_GATEWAY_HTTP_PORT=%q\n' "$SHARED_GATEWAY_HTTP_PORT"
        printf 'SHARED_GATEWAY_HTTPS_PORT=%q\n' "$SHARED_GATEWAY_HTTPS_PORT"
        printf 'SHARED_GATEWAY_CONFIG_FILE=%q\n' "$SHARED_GATEWAY_CONFIG_FILE"
        if twentyi_tls_available; then
            printf 'SHARED_GATEWAY_CERTS_DIR=%q\n' "$certs_dir"
        else
            printf 'SHARED_GATEWAY_CERTS_DIR=/dev/null\n'
        fi
    } >> "$shared_env_file"
}

twentyi_update_gateway_route() {
    local preferred_slug="${1:-}"
    local config_file backup_file gateway_container

    twentyi_refresh_registry
    twentyi_ensure_tls_cert
    config_file="$(twentyi_shared_gateway_config_file)"
    backup_file="$config_file.bak"

    if [[ -f "$config_file" ]]; then
        cp "$config_file" "$backup_file"
    else
        rm -f "$backup_file"
    fi

    twentyi_write_gateway_config "$preferred_slug"
    twentyi_write_shared_env
    twentyi_export_shared_env

    gateway_container="$(docker ps --filter "label=com.docker.compose.project=$SHARED_GATEWAY_COMPOSE_PROJECT_NAME" --filter "label=com.docker.compose.service=gateway" --format '{{.Names}}' | head -1)"

    if [[ -n "$gateway_container" ]]; then
        if ! twentyi_shared_compose up -d --no-deps --force-recreate gateway >/dev/null 2>&1; then
            if [[ -f "$backup_file" ]]; then
                mv "$backup_file" "$config_file"
                twentyi_write_shared_env
                twentyi_export_shared_env
                twentyi_shared_compose up -d --no-deps --force-recreate gateway >/dev/null 2>&1 || true
            fi
            printf 'Error: could not recreate shared gateway with updated config\n' >&2
            exit 1
        fi
        rm -f "$backup_file"
    else
        twentyi_shared_compose up -d
    fi

    twentyi_wait_for_gateway_ready
    if [[ -n "$preferred_slug" ]]; then
        twentyi_wait_for_route_target "${TWENTYI_GATEWAY_PROBE_TARGET:-twentyi-no-route}"
        twentyi_wait_for_gateway_route "${TWENTYI_GATEWAY_PROBE_HOSTNAME:-localhost}"
    fi
}

twentyi_ensure_shared_infra() {
    twentyi_require_docker

    if ! docker network inspect "$SHARED_GATEWAY_NETWORK" >/dev/null 2>&1; then
        docker network create "$SHARED_GATEWAY_NETWORK" >/dev/null
    fi

    twentyi_refresh_registry
    twentyi_write_gateway_config ""
    twentyi_write_shared_env
    twentyi_export_shared_env
    twentyi_shared_compose up -d
    twentyi_wait_for_gateway_ready
}

twentyi_note_phase_status() {
    printf 'Routing mode: shared gateway on host ports with hostname-aware gateway rules (.test DNS bootstrap still lands in later phases)\n'
    if [[ "${LOCAL_DNS_SUFFIX:-${SITE_SUFFIX:-}}" == "dev" && "$SHARED_GATEWAY_HTTPS_PORT" != "443" ]]; then
        printf 'Local .dev uses HTTPS port %s by default.\n' "$SHARED_GATEWAY_HTTPS_PORT"
    elif [[ "${TWENTYI_HTTPS_PORT_AUTO_FALLBACK:-0}" -eq 1 ]]; then
        printf 'HTTPS port 443 is busy on this machine; using %s instead.\n' "$SHARED_GATEWAY_HTTPS_PORT"
    fi
}

twentyi_stacklane_action_flag() {
    case "$1" in
        up)
            printf '%s' '--up'
            ;;
        attach)
            printf '%s' '--attach'
            ;;
        down)
            printf '%s' '--down'
            ;;
        detach)
            printf '%s' '--detach'
            ;;
        status)
            printf '%s' '--status'
            ;;
        logs)
            printf '%s' '--logs'
            ;;
        dns-setup)
            printf '%s' '--dns-setup'
            ;;
        *)
            return 1
            ;;
    esac
}

twentyi_legacy_command_name() {
    case "$1" in
        up)
            printf '%s' '20i-up'
            ;;
        attach)
            printf '%s' '20i-attach'
            ;;
        down)
            printf '%s' '20i-down'
            ;;
        detach)
            printf '%s' '20i-detach'
            ;;
        status)
            printf '%s' '20i-status'
            ;;
        logs)
            printf '%s' '20i-logs'
            ;;
        dns-setup)
            printf '%s' '20i-dns-setup'
            ;;
        *)
            return 1
            ;;
    esac
}

twentyi_stacklane_usage() {
    cat <<EOF
Stacklane

Usage: stacklane --ACTION [options]

Primary actions (choose exactly one):
  --up                     Start the current project and ensure shared infrastructure
  --attach                 Attach an additional project to the shared gateway
  --down                   Stop the current project runtime
  --detach                 Stop the current project runtime and remove its state record
  --status                 Show shared gateway, DNS, and project runtime state
  --logs                   Follow logs for the current or selected project
  --dns-setup              Bootstrap local DNS on macOS using Homebrew dnsmasq

Shared options:
  --project-dir PATH       Use a project directory other than the current directory
  --project SELECTOR       Target a recorded project for status/logs
  --site-name NAME         Override the hostname/project basename source
  --site-hostname HOST     Override the full planned hostname
  --site-suffix SUFFIX     Override the suffix used for planned hostnames (default: test)
  --docroot PATH           Override the document root (default: public_html when present)
  --php-version VERSION    Override the PHP image version
  --mysql-database NAME    Override the project database name
  --mysql-user USER        Override the project database user
  --mysql-password PASS    Override the project database password
  --mysql-port PORT        Override the database port used in the runtime
  --pma-port PORT          Override the phpMyAdmin port used in the runtime
  --all                    Apply supported commands globally (for example, --down --all)
  --dry-run                Resolve config and print the docker command without executing it
  --help                   Show this help

Compatibility aliases:
  version=8.4              Same as --php-version 8.4

Examples:
  stacklane --up
  stacklane --status --project marketing-site
  stacklane --down --all
  stacklane --logs apache

Migration:
  Legacy wrappers such as 20i-up and 20i-status still work temporarily, but they now forward to Stacklane.
  Prefer stacklane $(twentyi_stacklane_action_flag up), stacklane $(twentyi_stacklane_action_flag status), and related action flags in all new docs and shell aliases.
EOF
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

Additional commands:
    dns-setup                Bootstrap local .test DNS on macOS using Homebrew dnsmasq
EOF
}

twentyi_print_usage() {
    if [[ "${TWENTYI_ENTRYPOINT_MODE:-legacy}" == "stacklane" ]]; then
        twentyi_stacklane_usage
    else
        twentyi_usage
    fi
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
            --up)
                TWENTYI_PRIMARY_ACTIONS+=("up")
                ;;
            --attach)
                TWENTYI_PRIMARY_ACTIONS+=("attach")
                ;;
            --down)
                TWENTYI_PRIMARY_ACTIONS+=("down")
                ;;
            --detach)
                TWENTYI_PRIMARY_ACTIONS+=("detach")
                ;;
            --status)
                TWENTYI_PRIMARY_ACTIONS+=("status")
                ;;
            --logs)
                TWENTYI_PRIMARY_ACTIONS+=("logs")
                ;;
            --dns-setup)
                TWENTYI_PRIMARY_ACTIONS+=("dns-setup")
                ;;
            --project-dir)
                shift
                PROJECT_DIR="$1"
                ;;
            --project)
                shift
                PROJECT_SELECTOR="$1"
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
                twentyi_print_usage
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

twentyi_validate_stacklane_action_selection() {
    local action_count="${#TWENTYI_PRIMARY_ACTIONS[@]}"

    if [[ "$action_count" -eq 0 ]]; then
        printf 'Error: stacklane requires exactly one primary action flag.\n\n' >&2
        twentyi_stacklane_usage >&2
        exit 1
    fi

    if [[ "$action_count" -gt 1 ]]; then
        printf 'Error: primary actions are mutually exclusive:' >&2
        local action
        for action in "${TWENTYI_PRIMARY_ACTIONS[@]}"; do
            printf ' %s' "$(twentyi_stacklane_action_flag "$action")" >&2
        done
        printf '\n\n' >&2
        twentyi_stacklane_usage >&2
        exit 1
    fi

    TWENTYI_COMMAND="${TWENTYI_PRIMARY_ACTIONS[0]}"
}

twentyi_init_defaults() {
    PROJECT_DIR="$PWD"
    TWENTYI_ENTRYPOINT_MODE="${TWENTYI_ENTRYPOINT_MODE:-legacy}"
    TWENTYI_PRIMARY_ACTIONS=()
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
    LOCAL_DNS_PROVIDER="${LOCAL_DNS_PROVIDER:-dnsmasq}"
    LOCAL_DNS_IP="${LOCAL_DNS_IP:-127.0.0.1}"
    LOCAL_DNS_PORT="${LOCAL_DNS_PORT:-53535}"
    TWENTYI_HTTPS_PORT_AUTO_FALLBACK=0
    # LOCAL_DNS_SUFFIX is NOT defaulted here; it derives from SITE_SUFFIX in twentyi_finalize_context
    # so that .20i-local's SITE_SUFFIX=dev flows through correctly.
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
    LOCAL_DNS_SUFFIX="${LOCAL_DNS_SUFFIX:-$SITE_SUFFIX}"
    twentyi_resolve_shared_gateway_ports
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
    twentyi_validate_runtime_registration
    twentyi_update_gateway_route "$PROJECT_SLUG"

    printf 'Attached: %s\n' "$HOSTNAME"
    printf 'Hostname route URL: %s\n' "$(twentyi_hostname_route_url)"
    printf 'Gateway probe URL: %s\n' "$(twentyi_gateway_probe_url)"
    twentyi_warn_if_dns_not_ready
}

twentyi_down_like() {
    local state_file

    state_file="$(twentyi_project_state_file)"
    if [[ -f "$state_file" ]]; then
        twentyi_unset_project_state_vars
        twentyi_load_state_file "$state_file"
    fi

    if [[ "$TWENTYI_DRY_RUN" -eq 1 ]]; then
        printf 'Dry run: docker compose -f %s -p %s down\n' "$TWENTYI_STACK_FILE" "$COMPOSE_PROJECT_NAME"
        return 0
    fi

    twentyi_export_runtime_env
    twentyi_require_docker
    twentyi_compose down
    twentyi_reset_runtime_identity

    if [[ "$TWENTYI_COMMAND" == "detach" ]]; then
        twentyi_remove_state
        twentyi_update_gateway_route
        printf 'Detached: %s\n' "$PROJECT_NAME"
    else
        ATTACHMENT_STATE="down"
        twentyi_write_state
        twentyi_refresh_registry
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
        twentyi_unset_project_state_vars
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

    twentyi_refresh_registry

    if [[ "${TWENTYI_ENTRYPOINT_MODE:-legacy}" == "stacklane" ]]; then
        printf 'Stacklane status\n'
    else
        printf '20i stack status\n'
    fi
    printf 'Stack home: %s\n' "$STACK_HOME"
    printf 'State dir: %s\n' "$TWENTYI_STATE_DIR"
    printf 'Registry file: %s\n' "$(twentyi_registry_file)"
    printf 'Shared gateway: %s\n' "$(twentyi_shared_gateway_status)"
    printf 'Local DNS: %s\n' "$(twentyi_dns_status_message)"
    printf 'Shared network: %s\n' "$SHARED_GATEWAY_NETWORK"
    twentyi_note_phase_status
    printf '\n'

    if [[ -n "${PROJECT_SELECTOR:-}" ]]; then
        state_file="$(twentyi_state_file_for_selector "$PROJECT_SELECTOR" 2>/dev/null || true)"
        if [[ -z "$state_file" ]]; then
            printf 'No project matched selector: %s\n' "$PROJECT_SELECTOR"
            return 1
        fi
        set -- "$state_file"
    else
        set -- "$TWENTYI_STATE_DIR"/projects/*.env
    fi

    for state_file in "$@"; do
        [[ -e "$state_file" ]] || continue
        found=1
        twentyi_unset_project_state_vars
        twentyi_load_state_file "$state_file"

        printf '%s\n' "[$PROJECT_NAME]"
        printf '  state: %s\n' "$ATTACHMENT_STATE"
        printf '  compose project: %s\n' "$COMPOSE_PROJECT_NAME"
        printf '  hostname: %s\n' "$HOSTNAME"
        printf '  route url: %s\n' "$(twentyi_hostname_route_url)"
        printf '  gateway probe: %s\n' "$(twentyi_gateway_probe_url)"
        printf '  gateway alias: %s\n' "$WEB_NETWORK_ALIAS"
        printf '  runtime network: %s\n' "${COMPOSE_PROJECT_NAME}-runtime"
        printf '  db volume: %s\n' "${COMPOSE_PROJECT_NAME}-db-data"
        printf '  docroot: %s\n' "$DOCROOT"
        printf '  container docroot: %s\n' "$CONTAINER_DOCROOT"
        printf '  project dir: %s\n' "$PROJECT_DIR"
        printf '  containers: %s\n' "${RUNTIME_CONTAINER_SUMMARY:-none recorded}"
        printf '  drift: %s\n' "$(twentyi_registry_drift_status)"
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

    if [[ -n "${PROJECT_SELECTOR:-}" ]]; then
        state_file="$(twentyi_state_file_for_selector "$PROJECT_SELECTOR" 2>/dev/null || true)"
        if [[ -z "$state_file" ]]; then
            printf 'Error: no project matched selector %s\n' "$PROJECT_SELECTOR" >&2
            exit 1
        fi
    else
        state_file="$(twentyi_project_state_file)"
    fi

    if [[ -f "$state_file" ]]; then
        twentyi_unset_project_state_vars
        twentyi_load_state_file "$state_file"
        twentyi_export_runtime_env
    fi

    twentyi_require_docker

    if [[ -n "$service_name" ]]; then
        twentyi_compose logs -f "$service_name"
    else
        twentyi_compose logs -f
    fi
}

twentyi_legacy_forward() {
    local runtime_action="$1"
    shift

    local preferred_flag legacy_command
    preferred_flag="$(twentyi_stacklane_action_flag "$runtime_action")"
    legacy_command="$(twentyi_legacy_command_name "$runtime_action")"

    printf 'Notice: %s is deprecated. Use stacklane %s instead.\n' "$legacy_command" "$preferred_flag" >&2

    TWENTYI_ENTRYPOINT_MODE="stacklane"
    stacklane_main "$preferred_flag" "$@"
}

stacklane_main() {
    TWENTYI_ENTRYPOINT_MODE="stacklane"

    twentyi_init_defaults
    twentyi_parse_initial_args "$@"
    twentyi_load_stack_and_project_config
    twentyi_parse_args "$@"
    twentyi_validate_stacklane_action_selection
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
        dns-setup)
            twentyi_dns_setup
            ;;
        logs)
            twentyi_logs
            ;;
        *)
            printf 'Error: unsupported Stacklane action %s\n' "$TWENTYI_COMMAND" >&2
            exit 1
            ;;
    esac
}

twentyi_main() {
    TWENTYI_ENTRYPOINT_MODE="legacy"
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
        dns-setup)
            twentyi_dns_setup
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