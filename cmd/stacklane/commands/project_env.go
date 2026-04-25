package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/peternicholls/stacklane/core/config"
)

const projectEnvFileName = ".env.stacklane"

var safeEnvValue = regexp.MustCompile(`^[A-Za-z0-9_./:-]+$`)

func ensureProjectEnvFile(cfg config.ProjectConfig, flags *SharedFlags) error {
	path := filepath.Join(cfg.Dir, projectEnvFileName)
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("check project env: %w", err)
	}
	body := renderProjectEnvFile(cfg, flags)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return fmt.Errorf("write project env: %w", err)
	}
	return nil
}

func renderProjectEnvFile(cfg config.ProjectConfig, flags *SharedFlags) string {
	var b strings.Builder
	stackKind := cfg.StackKind
	if stackKind == "" {
		stackKind = "20i"
	}
	b.WriteString("# Stacklane project config\n")
	b.WriteString("# Created automatically on first `stacklane up` or `stacklane attach`.\n")
	b.WriteString("# Keep project-specific overrides here.\n")
	b.WriteString("# Shared defaults stay in <stack-home>/.env.stacklane.\n")
	b.WriteString("# Project .env remains application-owned.\n\n")
	b.WriteString("STACKLANE_STACK=")
	b.WriteString(renderEnvValue(stackKind))
	b.WriteString("\n\n")

	overrides := explicitProjectOverrides(cfg, flags)
	if len(overrides) > 0 {
		b.WriteString("# Persisted from the first command run\n")
		for _, line := range overrides {
			b.WriteString(line)
			b.WriteByte('\n')
		}
		b.WriteByte('\n')
	}

	b.WriteString("# Uncomment and edit only what this project should own\n")
	b.WriteString("# SITE_NAME=")
	b.WriteString(renderCommentValue(cfg.Name))
	b.WriteByte('\n')
	b.WriteString("# SITE_HOSTNAME=")
	b.WriteString(renderCommentValue(cfg.Hostname))
	b.WriteByte('\n')
	b.WriteString("# SITE_SUFFIX=")
	b.WriteString(renderCommentValue(cfg.SiteSuffix))
	b.WriteByte('\n')
	if cfg.DocRootRelative != "" {
		b.WriteString("# DOCROOT=")
		b.WriteString(renderCommentValue(cfg.DocRootRelative))
		b.WriteByte('\n')
	} else {
		b.WriteString("# DOCROOT=public_html\n")
	}
	b.WriteString("# PHP_VERSION=")
	b.WriteString(renderCommentValue(cfg.PHPVersion))
	b.WriteByte('\n')
	b.WriteString("# MYSQL_DATABASE=")
	b.WriteString(renderCommentValue(cfg.MySQL.Database))
	b.WriteByte('\n')
	b.WriteString("# MYSQL_USER=")
	b.WriteString(renderCommentValue(cfg.MySQL.User))
	b.WriteByte('\n')
	b.WriteString("# MYSQL_PASSWORD=")
	b.WriteString(renderCommentValue(cfg.MySQL.Password))
	b.WriteByte('\n')
	b.WriteString("# MYSQL_PORT=33060\n")
	b.WriteString("# PMA_PORT=33061\n")
	b.WriteString("# STACKLANE_POST_UP_COMMAND=php artisan migrate --force --no-interaction\n")
	return b.String()
}

func explicitProjectOverrides(cfg config.ProjectConfig, flags *SharedFlags) []string {
	if flags == nil {
		return nil
	}
	var lines []string
	add := func(key, value string) {
		if value == "" {
			return
		}
		lines = append(lines, key+"="+renderEnvValue(value))
	}
	add("SITE_NAME", flags.SiteName)
	add("SITE_HOSTNAME", flags.SiteHostname)
	add("SITE_SUFFIX", flags.SiteSuffix)
	add("DOCROOT", flags.DocRoot)
	add("PHP_VERSION", flags.PHPVersion)
	add("MYSQL_DATABASE", flags.MySQLDatabase)
	add("MYSQL_USER", flags.MySQLUser)
	add("MYSQL_PASSWORD", flags.MySQLPassword)
	add("MYSQL_PORT", flags.MySQLPort)
	add("PMA_PORT", flags.PMAPort)
	add("HOST_PORT", flags.HostPort)
	if flags.WaitTimeoutSecs > 0 {
		lines = append(lines, fmt.Sprintf("STACKLANE_WAIT_TIMEOUT=%d", flags.WaitTimeoutSecs))
	}
	if cfg.PostUpCommand != "" {
		add("STACKLANE_POST_UP_COMMAND", cfg.PostUpCommand)
	}
	return lines
}

func renderCommentValue(value string) string {
	value = strings.ReplaceAll(value, "\n", " ")
	return strings.TrimSpace(value)
}

func renderEnvValue(value string) string {
	value = renderCommentValue(value)
	if safeEnvValue.MatchString(value) {
		return value
	}
	if !strings.Contains(value, "'") {
		return "'" + value + "'"
	}
	if !strings.Contains(value, `"`) {
		return `"` + value + `"`
	}
	return value
}
