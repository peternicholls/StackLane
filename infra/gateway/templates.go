// Templated nginx config for the shared gateway. Replaces the heredoc
// string-interpolation in stageserve_gateway_block_for_route /
// stageserve_write_gateway_config with a typed text/template.
package gateway

import (
	"bytes"
	"strings"
	"text/template"
)

const gatewayTpl = `{{- if .TLSEnabled -}}
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

    add_header X-StageServe-Gateway "shared" always;
    add_header X-StageServe-Route-Target "{{ if eq (len .Routes) 0 }}stageserve-no-route{{ else }}unmatched-host{{ end }}" always;

    location = /__stageserve_gateway_health {
        default_type text/plain;
        return 200 "gateway ok\n";
    }

    location / {
        default_type text/plain;
{{- if eq (len .Routes) 0 }}
        return 503 "StageServe shared gateway has no hostname routes.\n";
{{- else }}
        add_header X-StageServe-Route-State "unmatched-host" always;
        return 404 "StageServe shared gateway has no route for host '$host'.\n";
{{- end }}
    }
}
{{- else -}}
server {
    listen 80 default_server;
    listen 443 default_server;
    server_name _;

    access_log /var/log/nginx/access.log;
    error_log /var/log/nginx/error.log warn;

    add_header X-StageServe-Gateway "shared" always;
    add_header X-StageServe-Route-Target "{{ if eq (len .Routes) 0 }}stageserve-no-route{{ else }}unmatched-host{{ end }}" always;

    location = /__stageserve_gateway_health {
        default_type text/plain;
        return 200 "gateway ok\n";
    }

    location / {
        default_type text/plain;
{{- if eq (len .Routes) 0 }}
        return 503 "StageServe shared gateway has no hostname routes.\n";
{{- else }}
        add_header X-StageServe-Route-State "unmatched-host" always;
        return 404 "StageServe shared gateway has no route for host '$host'.\n";
{{- end }}
    }
}
{{- end }}
{{ range .Routes }}
{{- if $.TLSEnabled }}
server {
    listen 80;
    server_name {{ .Hostname }};
    return 301 {{ if eq $.HTTPSPort 443 }}https://$host$request_uri{{ else }}https://$host:{{ $.HTTPSPort }}$request_uri{{ end }};
}
server {
    listen 443 ssl;
    server_name {{ .Hostname }};

    ssl_certificate     /etc/nginx/certs/tls.pem;
    ssl_certificate_key /etc/nginx/certs/tls-key.pem;
    ssl_protocols       TLSv1.2 TLSv1.3;
    ssl_ciphers         HIGH:!aNULL:!MD5;

    access_log /var/log/nginx/access.log;
    error_log /var/log/nginx/error.log warn;

    add_header X-StageServe-Gateway "shared" always;
    add_header X-StageServe-Route-Target "{{ .WebNetworkAlias }}" always;
    add_header X-StageServe-Hostname "{{ .Hostname }}" always;

    location / {
        resolver 127.0.0.11 valid=5s;
        set $upstream http://{{ .WebNetworkAlias }}:80;

        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-Host $host;
        proxy_set_header X-Forwarded-Proto https;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_connect_timeout 2s;
        proxy_read_timeout 600s;
        proxy_pass $upstream;
    }
}
{{- else }}
server {
    listen 80;
    listen 443;
    server_name {{ .Hostname }};

    access_log /var/log/nginx/access.log;
    error_log /var/log/nginx/error.log warn;

    add_header X-StageServe-Gateway "shared" always;
    add_header X-StageServe-Route-Target "{{ .WebNetworkAlias }}" always;
    add_header X-StageServe-Hostname "{{ .Hostname }}" always;

    location / {
        resolver 127.0.0.11 valid=5s;
        set $upstream http://{{ .WebNetworkAlias }}:80;

        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-Host $host;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_connect_timeout 2s;
        proxy_read_timeout 600s;
        proxy_pass $upstream;
    }
}
{{- end }}
{{ end }}
`

var tpl = template.Must(template.New("gateway").Parse(gatewayTpl))

// Render produces the full nginx config for input. Trailing whitespace is
// trimmed so the output is deterministic across rebuilds (golden tests).
func Render(input RenderInput) (string, error) {
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, input); err != nil {
		return "", err
	}
	out := strings.TrimRight(buf.String(), "\n") + "\n"
	return out, nil
}
