package user

import "fst/backend/app/services"

// AdminUserRealnameSummary 管理端用户详情中的实名摘要
type AdminUserRealnameSummary struct {
	HasVerification bool   `json:"has_verification"`
	ID              uint64 `json:"id,omitempty"`
	Status          uint8  `json:"status,omitempty"`
	RealName        string `json:"real_name,omitempty"`
	CertificateType uint8  `json:"certificate_type,omitempty"`
	CertificateNo   string `json:"certificate_no,omitempty"`
	SubmittedAt     *int64 `json:"submitted_at,omitempty"`
	ReviewedAt      *int64 `json:"reviewed_at,omitempty"`
	RejectReason    string `json:"reject_reason,omitempty"`
}
//@name 管理端用户实名认证摘要

// AdminUserDetailResponse 管理端用户详情响应
type AdminUserDetailResponse struct {
	User     *services.AdminUserListItem `json:"user"`
	Realname AdminUserRealnameSummary    `json:"realname"`
}
//@name 管理端用户详情响应
