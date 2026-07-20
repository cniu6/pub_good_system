package sms_plugin

import (
	"sort"
	"strconv"
	"strings"
)

const (
	metaTemplateNameKey  = "__template_name"
	metaTemplateOrderKey = "__template_order"
	metaUserIDKey        = "__user_id"
)

// ExtractMetaUserID 从模板参数中取出关联用户 ID（匿名发送为 0）。
func ExtractMetaUserID(templateParams map[string]string) uint64 {
	if templateParams == nil {
		return 0
	}
	raw := strings.TrimSpace(templateParams[metaUserIDKey])
	if raw == "" {
		return 0
	}
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0
	}
	return id
}

func normalizeTemplateParams(code string, expireMinutes int, templateParams map[string]string) (string, map[string]string, []string) {
	payload := map[string]string{
		"code":   code,
		"expire": strconv.Itoa(expireMinutes),
	}
	templateName := "send_code"

	for k, v := range templateParams {
		key := strings.TrimSpace(k)
		if key == "" {
			continue
		}
		switch key {
		case metaTemplateNameKey:
			if strings.TrimSpace(v) != "" {
				templateName = strings.TrimSpace(v)
			}
		case metaTemplateOrderKey, metaUserIDKey:
			// 元数据，不传给云厂商模板参数
		default:
			payload[key] = v
		}
	}

	order := buildTemplateParamOrder(templateParams, payload)
	return templateName, payload, order
}

func buildTemplateParamOrder(templateParams map[string]string, payload map[string]string) []string {
	var order []string
	seen := map[string]struct{}{}

	if raw := strings.TrimSpace(templateParams[metaTemplateOrderKey]); raw != "" {
		for _, item := range strings.Split(raw, ",") {
			key := strings.TrimSpace(item)
			if key == "" {
				continue
			}
			if _, ok := payload[key]; !ok {
				continue
			}
			if _, ok := seen[key]; ok {
				continue
			}
			order = append(order, key)
			seen[key] = struct{}{}
		}
	}

	for _, key := range []string{"code", "expire"} {
		if _, ok := payload[key]; !ok {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		order = append(order, key)
		seen[key] = struct{}{}
	}

	var extras []string
	for key := range payload {
		if _, ok := seen[key]; ok {
			continue
		}
		extras = append(extras, key)
	}
	sort.Strings(extras)
	order = append(order, extras...)
	return order
}

