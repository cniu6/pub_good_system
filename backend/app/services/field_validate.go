package services

import "fst/backend/utils"

// validateClientRuneLen 用户/管理端输入按 rune 长度校验；超长返回 ClientError（HTTP 400）。
func validateClientRuneLen(s, fieldName string, max int) error {
	if err := utils.ValidateRuneLen(s, fieldName, max); err != nil {
		return NewClientError(err.Error())
	}
	return nil
}
